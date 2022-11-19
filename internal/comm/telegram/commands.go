package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func (b *Bot) RegisterCommand(cmd string, description string, handler CommandHandler) error {
	existingCommand, err := b.client.GetMyCommands()
	if err != nil {
		return err
	}

	found := false
	for i, command := range existingCommand {
		if command.Command == cmd {
			found = true
			existingCommand[i].Description = description
		}
	}

	if !found {
		existingCommand = append(existingCommand, tgbotapi.BotCommand{
			Command:     cmd,
			Description: description,
		})
	}

	if _, err := b.client.Request(tgbotapi.NewSetMyCommands(existingCommand...)); err != nil {
		return err
	}

	b.cmdHandlers[cmd] = handler
	return nil

}
