package telegram

import (
	"fmt"
	"html/template"
	"log"
	"strconv"
	"strings"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/internal/comm/model"
	"github.com/eternal-flame-AD/yoake/internal/db"
	"github.com/eternal-flame-AD/yoake/internal/servetpl/funcmap"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Role int

const (
	RoleAnonymous Role = 0
	RoleFriend    Role = 1
	RoleOwner     Role = 2
)

type CommandHandler func(bot *Bot, role Role, update tgbotapi.Update) error

type Bot struct {
	client      *tgbotapi.BotAPI
	database    db.DB
	cmdHandlers map[string]CommandHandler

	Name string `json:"name"`

	LastUpdateID int   `json:"last_update_id"`
	OwnerChatID  int64 `json:"owner_user_id"`
}

func (b *Bot) SupportedMIME() []string {
	return []string{"text/markdown", "text/html"}
}

func (b *Bot) SendGenericMessage(message *model.GenericMessage) error {
	if b.OwnerChatID == 0 {
		return fmt.Errorf("owner chat id not set")
	}
	chattable := tgbotapi.NewMessage(b.OwnerChatID, message.Body)
	switch message.MIME {
	case "text/markdown":
		chattable.ParseMode = tgbotapi.ModeMarkdownV2
	case "text/html":
		chattable.ParseMode = tgbotapi.ModeHTML
	default:
		return fmt.Errorf("unsupported MIME type %s", message.MIME)
	}

	if message.ThreadID != 0 {
		chattable.ReplyToMessageID = int(message.ThreadID)
	}
	msg, err := b.client.Send(chattable)
	if err != nil {
		return err
	}
	if message.ThreadID == 0 {
		message.ThreadID = uint64(msg.MessageID)
	}

	return nil
}

func (b *Bot) saveConf() error {
	txn := b.database.NewTransaction(true)
	defer txn.Discard()
	if err := db.SetJSON(txn, []byte(fmt.Sprintf("comm_telegram_bot_%d", b.client.Self.ID)), b); err != nil {
		return err
	}
	return txn.Commit()
}

func (b *Bot) Client() *tgbotapi.BotAPI {
	return b.client
}

func (b *Bot) SendHTML(chatID int64, fmtStr string, args ...interface{}) error {
	for i := range args {
		switch v := args[i].(type) {
		case string:
			args[i] = template.HTMLEscapeString(args[i].(string))
		default:
			args[i] = v
		}
	}
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(fmtStr, args...))
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := b.client.Send(msg)
	return err
}
func (b *Bot) handleUpdate(update tgbotapi.Update) error {

	role := RoleAnonymous
	if msg := update.Message; msg != nil {
		log.Printf("received message from %s: %v", update.Message.From.UserName, update.Message)
		msg := *update.Message
		conf := config.Config().Comm.Telegram
		if strings.HasPrefix(conf.Owner, "@") && msg.From.UserName == conf.Owner[1:] {
			if msg.Chat.ID != b.OwnerChatID {
				log.Printf("telegram owner chat id set: %d", msg.Chat.ID)
				b.OwnerChatID = msg.Chat.ID
			}
		} else if id, err := strconv.ParseInt(conf.Owner, 10, 64); err == nil && msg.From.ID == id || msg.Chat.ID == id {
			if msg.Chat.ID != b.OwnerChatID {
				log.Printf("telegram owner chat id set: %d", msg.Chat.ID)
				b.OwnerChatID = msg.Chat.ID
			}
		}
		if msg.Chat.ID == b.OwnerChatID {
			role = RoleOwner
		}
		if msg.IsCommand() {
			if handler, ok := b.cmdHandlers[msg.Command()]; ok {
				return handler(b, role, update)
			} else {
				if err := b.SendHTML(msg.Chat.ID, "unknown command: %s\n", msg.Command()); err != nil {
					return err
				}
			}
		}
	}

	if callback := update.CallbackQuery; callback != nil {
		cb := tgbotapi.NewCallback(callback.ID, "")
		defer b.client.Send(cb)
		if msg := callback.Message; msg != nil {
			if callbackMsg := msg.ReplyToMessage; callbackMsg != nil && callbackMsg.IsCommand() {
				if handler, ok := b.cmdHandlers[callbackMsg.Command()]; ok {
					log.Printf("callback for command %s", callbackMsg.Command())
					return handler(b, role, update)
				} else {
					if err := b.SendHTML(callback.Message.Chat.ID, "unknown command: %s\n", callback.Message.Command()); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (bot *Bot) start() error {
	u := tgbotapi.NewUpdate(bot.LastUpdateID + 1)
	u.Timeout = 60

	bot.RegisterCommand("start", "onboarding command", func(bot *Bot, role Role, upd tgbotapi.Update) error {
		if msg := upd.Message; msg != nil {
			if _, err := bot.client.Send(tgbotapi.NewMessage(msg.Chat.ID, strings.ReplaceAll(banner, "{name}", msg.From.FirstName+" "+msg.From.LastName))); err != nil {
				return err
			}
		}
		return nil
	})

	updates := bot.client.GetUpdatesChan(u)
	go func() {
		for update := range updates {
			if err := bot.handleUpdate(update); err != nil {
				fromChat := update.FromChat()
				if stopImgURL, err := funcmap.TrimaImg("ja/btn_stop.gif", "url"); err != nil {
					log.Printf("failed to get stop image: %v", err)
				} else {
					stopPhoto := tgbotapi.NewPhoto(fromChat.ID, tgbotapi.FileURL(stopImgURL))
					stopPhoto.DisableNotification = true
					bot.client.Send(stopPhoto)
				}
				if fromChat.ID != bot.OwnerChatID {
					bot.SendHTML(fromChat.ID, "<b>Runtime Error</b>\n<pre>%s</pre>\nBot owner has been notified.", err)

				}
				bot.SendHTML(bot.OwnerChatID, "<b>Runtime Error</b>\noriginating chat ID: %d (@%s)\n\n<pre>%s</pre>", update.FromChat().ID, update.FromChat().UserName, err)
				log.Printf("telegram runtime error: %v", err)
			}

			if update.UpdateID > bot.LastUpdateID {
				bot.LastUpdateID = update.UpdateID
			}
			if err := bot.saveConf(); err != nil {
				log.Printf("failed to save telegram bot config: %v", err)
			}
		}
	}()
	return nil
}

func loadConf(confName string, database db.DB) Bot {
	txn := database.NewTransaction(false)
	defer txn.Discard()
	conf := Bot{LastUpdateID: -1}

	if err := db.GetJSON(txn, []byte(confName), &conf); db.IsNotFound(err) {
		log.Printf("telegram bot config  %s not found, creating new one", confName)
		txn.Discard()
		txn = database.NewTransaction(true)
		defer txn.Discard()
		if err := db.SetJSON(txn, []byte(confName), conf); err != nil {
			log.Fatalf("failed to create telegram bot config %s: %v", confName, err)
		}
	} else if err != nil {
		log.Fatalf("failed to load telegram bot config %s: %v",
			confName, err)
	}
	return conf
}

func NewClient(database db.DB) (*Bot, error) {
	token := config.Config().Comm.Telegram.Token
	if token == "" {
		return nil, fmt.Errorf("telegram token not set")
	}

	client, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	bot := loadConf(fmt.Sprintf("comm_telegram_bot_%d", client.Self.ID), database)
	bot.client = client
	bot.database = database
	bot.cmdHandlers = make(map[string]CommandHandler)

	return &bot, bot.start()
}
