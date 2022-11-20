package twilio

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/internal/comm/model"
	"github.com/eternal-flame-AD/yoake/internal/comm/telegram"
	"github.com/eternal-flame-AD/yoake/internal/filestore"
	"github.com/eternal-flame-AD/yoake/internal/servetpl/funcmap"
	"github.com/labstack/echo/v4"
	"github.com/spf13/afero"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const rfc2822 = "Mon, 02 Jan 2006 15:04:05 -0700"

type RecordingStatusCallbackForm struct {
	AccountSid         string `form:"AccountSid"  json:"AccountSid"`
	CallSid            string `form:"CallSid"  json:"CallSid"`
	RecordingSid       string `form:"RecordingSid"  json:"RecordingSid"`
	RecordingUrl       string `form:"RecordingUrl"  json:"RecordingUrl"`
	RecordingStatus    string `form:"RecordingStatus"  json:"RecordingStatus"`
	RecodgingDuration  int    `form:"RecordingDuration"  json:"RecordingDuration"`
	RecordingChannels  int    `form:"RecordingChannels"  json:"RecordingChannels"`
	RecodgingStartTime string `form:"RecordingStartTime"  json:"RecordingStartTime"`
	RecordingSource    string `form:"RecordingSource"  json:"RecordingSource"`
	RecordingTrack     string `form:"RecordingTrack"  json:"RecordingTrack"`
}

type CallStatusCallbackForm struct {
	CallSid       string `form:"CallSid"  json:"CallSid"`
	AccountSid    string `form:"AccountSid"  json:"AccountSid"`
	From          string `form:"From"  json:"From"`
	To            string `form:"To"  json:"To"`
	CallStatus    string `form:"CallStatus"  json:"CallStatus"`
	ApiVersion    string `form:"ApiVersion"  json:"ApiVersion"`
	Direction     string `form:"Direction"  json:"Direction"`
	ForwardedFrom string `form:"ForwardedFrom"  json:"ForwardedFrom"`
	CallerName    string `form:"CallerName"  json:"CallerName"`
	ParentCallSid string `form:"ParentCallSid"  json:"ParentCallSid"`
}

type CallStatusCallbackProgressForm struct {
	CallStatusCallbackForm
	// https://www.twilio.com/docs/voice/api/call-resource#statuscallbackevent
	CallStatus        string `form:"CallStatus"  json:"CallStatus"`
	Duration          int    `form:"Duration"  json:"Duration"`
	CallDuration      int    `form:"CallDuration"  json:"CallDuration"`
	SipResponseCode   int    `form:"SipResponseCode"  json:"SipResponseCode"`
	RecordingUrl      string `form:"RecordingUrl"  json:"RecordingUrl"`
	RecordingSid      string `form:"RecordingSid"  json:"RecordingSid"`
	RecordingDuration int    `form:"RecordingDuration"  json:"RecordingDuration"`
	TimeStamp         string `form:"TimeStamp"  json:"TimeStamp"`
	CallbackSource    string `form:"CallbackSource"  json:"CallbackSource"`
	SequenceNumber    int    `form:"SequenceNumber"  json:"SequenceNumber"`
}

func findCallDir(callsDir filestore.FS, callSid string, from string, to string) (filestore.FS, string, error) {
	if callSid == "" {
		return nil, "", fmt.Errorf("callSid is empty")
	}
	dirs, err := afero.ReadDir(callsDir, ".")
	if err != nil {
		return nil, "", err
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			if strings.HasSuffix(dir.Name(), callSid) {
				return filestore.ChrootFS(callsDir, dir.Name()), dir.Name(), nil
			}
		}
	}
	now := time.Now()
	newName := fmt.Sprintf("%s_%s_%s_%s", now.Format("2006-01-02T15.04.05"), from, to, callSid)
	if err := callsDir.Mkdir(newName, 0770); err != nil {
		return nil, "", err
	}
	return filestore.ChrootFS(callsDir, newName), newName, nil
}

func fetchRecording(apiClient *twilio.RestClient, callDir filestore.FS, sid string, recType string) error {
	prm := new(openapi.ListRecordingParams)
	prm.SetCallSid(sid)
	recordings, err := apiClient.Api.ListRecording(prm)
	if err != nil {
		return err
	}
	if err := callDir.Mkdir("recordings", 0770); err != nil && !os.IsExist(err) {
		return err
	}
	recDir := filestore.ChrootFS(callDir, "recordings")

	wg := new(sync.WaitGroup)
	for i, recording := range recordings {
		wg.Add(1)
		go func(i int, recording openapi.ApiV2010Recording) {
			defer wg.Done()
			rSid := recording.Sid
			if rSid == nil {
				rSidS := fmt.Sprintf("unknown-%d-%d", time.Now().Unix(), i)
				rSid = &rSidS
			}
			jsonF, err := recDir.OpenFile(recType+"_"+*recording.Sid+".json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0660)
			if err != nil {
				log.Printf("failed to open json file for recording %s: %v", *rSid, err)
				return
			}
			defer jsonF.Close()
			enc := json.NewEncoder(jsonF)
			enc.SetIndent("", "  ")
			if err := enc.Encode(recording); err != nil {
				log.Printf("failed to write json file for recording %s: %v", *rSid, err)
				return
			}

			mediaResp, err := http.Get(*recording.MediaUrl + ".wav?requestedChannels=2")
			if err != nil {
				log.Printf("failed to download media file for recording %s: %v", *rSid, err)
				return
			}
			defer mediaResp.Body.Close()
			if mediaResp.StatusCode != http.StatusOK {
				log.Printf("failed to download media file for recording %s: http status %d", *rSid, mediaResp.StatusCode)
			}

			mediaF, err := recDir.OpenFile(recType+"_"+*recording.Sid+".wav", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0660)
			if err != nil {
				log.Printf("failed to open media file for recording %s: %v", *rSid, err)
				return
			}
			defer mediaF.Close()

			if _, err := io.Copy(mediaF, mediaResp.Body); err != nil {
				log.Printf("failed to write media file for recording %s: %v", *rSid, err)
				return
			}

			if err := apiClient.Api.DeleteRecording(*rSid, new(openapi.DeleteRecordingParams)); err != nil {
				log.Printf("failed to delete recording %s: %v", *rSid, err)
			}

		}(i, recording)
	}
	wg.Wait()

	return nil
}
func Register(g *echo.Group, fs filestore.FS, comm model.Communicator) {
	apiClient := twilio.NewRestClientWithParams(twilio.ClientParams{
		AccountSid: config.Config().Twilio.AccountSid,
		Username:   config.Config().Twilio.AccountSid,
		Password:   config.Config().Twilio.AuthToken,
	})

	if err := fs.Mkdir("calls", 0770); err != nil && !os.IsExist(err) {
		log.Panicf("failed to create calls directory: %v", err)
	}
	calls := filestore.ChrootFS(fs, "calls")
	tg, hasTelegram := comm.GetMethod("telegram").(*telegram.Bot)

	process := g.Group("/process", VerifyMiddleware("", config.Config().Twilio.BaseURL))
	{
		process.POST("/voicemail/:type", func(c echo.Context) error {
			stateForm := new(CallStatusCallbackForm)
			err := c.Bind(stateForm)
			if err != nil {
				return err
			}
			thenUrl := c.QueryParam("then")
			typeStr := c.Param("type")
			switch typeStr {
			case "message":
				err = tg.SendHTML(tg.OwnerChatID, "New voicemail request\n\ncallSid: %s\nFrom: %s\n", stateForm.CallSid, stateForm.From)
			case "callback":
				err = tg.SendHTML(tg.OwnerChatID, "New callback request\n\ncallSid: %s\nFrom: %s\n", stateForm.CallSid, stateForm.From)
			default:
				return c.String(http.StatusBadRequest, "invalid type")
			}
			if err != nil {
				return fmt.Errorf("failed to send message: %v", err)
			}
			return c.Redirect(http.StatusTemporaryRedirect, thenUrl)
		})
		process.POST("/incoming_owner", func(c echo.Context) error {
			return c.Redirect(http.StatusTemporaryRedirect, c.QueryParam("then"))
		})
	}

	cb := g.Group("/callback", VerifyMiddleware("", config.Config().Twilio.BaseURL), func(next echo.HandlerFunc) echo.HandlerFunc {

		return func(c echo.Context) error {
			if err := next(c); err != nil {
				log.Printf("failed to process twilio callback: %v", err)
				return err
			}
			return nil
		}
	})
	{
		cb.POST("/recording/:type", func(c echo.Context) error {
			form := new(RecordingStatusCallbackForm)
			if err := c.Bind(form); err != nil {
				return err
			}
			sid := form.CallSid
			if sid == "" {
				return c.String(http.StatusBadRequest, "missing call sid")
			}
			callDir, _, err := findCallDir(calls, form.CallSid, "", "")
			if err != nil {
				return err
			}
			return fetchRecording(apiClient, callDir, form.CallSid, c.Param("type"))
		})

		cb.POST("/voice", func(c echo.Context) error {
			form := new(CallStatusCallbackForm)
			if err := c.Bind(form); err != nil {
				return err
			}
			callDir, callDirName, err := findCallDir(calls, form.CallSid, form.From, form.To)
			if err != nil {
				return err
			}
			callDirAbs := fmt.Sprintf("/calls/%s", callDirName)
			log.Printf("call %s: %s -> %s dirAbs=%s", form.CallSid, form.From, form.To, callDirAbs)
			if hasTelegram {
				msg := tgbotapi.NewMessage(tg.OwnerChatID, fmt.Sprintf("Call From %s (%s):\n\nTo: %s\nSid: %s\nCallDir: <a href=\"%s\">%s</a>",
					form.From, form.CallStatus, form.To, form.CallSid, funcmap.FileAccess(callDirAbs+"/"), callDirAbs))
				msg.ParseMode = tgbotapi.ModeHTML
				if _, err := tg.Client().Send(msg); err != nil {
					log.Printf("failed to send telegram message: %v", err)
				}
			}

			fileName := fmt.Sprintf("status.%d.json", time.Now().UnixNano())
			f, err := callDir.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0660)
			if err != nil {
				return err
			}
			defer f.Close()
			enc := json.NewEncoder(f)
			enc.SetIndent("", "  ")
			if err := enc.Encode(form); err != nil {
				return err
			}
			return nil
		})
	}
}
