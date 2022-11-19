package comm

import (
	"errors"
	"fmt"
	"log"

	"github.com/eternal-flame-AD/yoake/internal/comm/email"
	"github.com/eternal-flame-AD/yoake/internal/comm/gotify"
	"github.com/eternal-flame-AD/yoake/internal/comm/model"
	"github.com/eternal-flame-AD/yoake/internal/comm/telegram"
	"github.com/eternal-flame-AD/yoake/internal/db"
	"github.com/eternal-flame-AD/yoake/internal/util"
)

type Communicator struct {
	commMethods           map[string]model.CommMethod
	fallbackCommunicators []string
}

var (
	errMethodNotSupported = errors.New("method not supported")
)

func (c *Communicator) actualSendGenericMessage(tryMethod string, message *model.GenericMessage) error {
	if comm, ok := c.commMethods[tryMethod]; ok {
		if convertedMsg, err := ConvertGenericMessage(message, comm.SupportedMIME()); err == nil {
			err := comm.SendGenericMessage(convertedMsg)
			if err != nil {
				return err
			}
			message.ThreadID = convertedMsg.ThreadID
			return nil
		} else {
			return err
		}
	}
	return errMethodNotSupported
}

// GetMethod returns the method with the given name.
func (c *Communicator) GetMethod(method string) model.CommMethod {
	return c.commMethods[method]
}

// GetMethodsByMIME returns a list of methods that support the given MIME type as the message type, MIME convertions were considered.
func (c *Communicator) GetMethodsByMIME(mime string) []model.CommMethod {
	var result []model.CommMethod
	for _, comm := range c.commMethods {
		if util.Contain(ConvertOutMIMEToSupportedInMIME(comm.SupportedMIME()), mime) {
			result = append(result, comm)
		}
	}
	return result
}

type ErrorSentWithFallback struct {
	OriginalError  error
	OrignalMethod  string
	FallbackMethod string
}

func (e ErrorSentWithFallback) Error() string {
	return fmt.Sprintf("used fallback method %s because original method %s reeported error: %v", e.FallbackMethod, e.OrignalMethod, e.OriginalError)
}

// SendGenericMethods sends a message using the preferred method
// if the preferred method failed to send the message, fallback methods will be tried,
// and an ErrorSentWithFabback will be returned if any fallback method succeeded.
// if fallback methods failed as well the original error will be returned.
func (c *Communicator) SendGenericMessage(preferredMethod string, message *model.GenericMessage, force bool) error {
	if preferredMethod == "" {
		preferredMethod = c.fallbackCommunicators[0]
	}
	if origErr := c.actualSendGenericMessage(preferredMethod, message); origErr != nil {
		if force {
			log.Printf("Failed to send message using preferred method %s: %v", preferredMethod, origErr)
			return origErr
		}
		log.Printf("Failed to send message using preferred method %s: %v. trying fallback methods", preferredMethod, origErr)
		for _, fallback := range c.fallbackCommunicators {
			if fallback == preferredMethod {
				continue
			}
			if err := c.actualSendGenericMessage(fallback, message); err == nil {
				log.Printf("Sent message using fallback method %s", fallback)
				return ErrorSentWithFallback{
					OriginalError:  origErr,
					OrignalMethod:  preferredMethod,
					FallbackMethod: fallback,
				}
			} else {
				log.Printf("Failed to send message using fallback method %s: %v", fallback, err)
			}
		}
		return origErr
	}
	return nil
}

func InitCommunicator(database db.DB) *Communicator {
	comm := &Communicator{
		commMethods: make(map[string]model.CommMethod),
	}
	if emailHandler, err := email.NewHandler(); err == nil {
		comm.commMethods["email"] = emailHandler
		comm.fallbackCommunicators = append(comm.fallbackCommunicators, "email")
	} else {
		log.Printf("Failed to initialize email communicator: %v", err)
	}
	if gotifyHandler, err := gotify.NewClient(); err == nil {
		comm.commMethods["gotify"] = gotifyHandler
		comm.fallbackCommunicators = append(comm.fallbackCommunicators, "gotify")
	} else {
		log.Printf("Failed to initialize gotify communicator: %v", err)
	}
	if telegramHandler, err := telegram.NewClient(database); err == nil {
		comm.commMethods["telegram"] = telegramHandler
		comm.fallbackCommunicators = append(comm.fallbackCommunicators, "telegram")
	} else {
		log.Printf("Failed to initialize telegram communicator: %v", err)
	}

	return comm
}
