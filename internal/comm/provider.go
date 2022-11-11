package comm

import (
	"errors"
	"log"

	"github.com/eternal-flame-AD/yoake/internal/comm/email"
	"github.com/eternal-flame-AD/yoake/internal/comm/gotify"
	"github.com/eternal-flame-AD/yoake/internal/comm/model"
)

type CommProvider struct {
	communicators         map[string]Communicator
	fallbackCommunicators []string
}

var (
	errMethodNotSupported = errors.New("method not supported")
)

func (c *CommProvider) actualSendGenericMessage(tryMethod string, message model.GenericMessage) error {
	if comm, ok := c.communicators[tryMethod]; ok {
		if convertedMsg, err := ConvertGenericMessage(&message, comm.SupportedMIME()); err == nil {
			return comm.SendGenericMessage(*convertedMsg)
		} else {
			return err
		}
	}
	return errMethodNotSupported
}

func (c *CommProvider) SendGenericMessage(preferredMethod string, message model.GenericMessage) error {
	if preferredMethod == "" {
		preferredMethod = c.fallbackCommunicators[0]
	}
	if err := c.actualSendGenericMessage(preferredMethod, message); err != nil {
		log.Printf("Failed to send message using preferred method %s: %v. trying fallback methods", preferredMethod, err)
		for _, fallback := range c.fallbackCommunicators {
			if fallback == preferredMethod {
				continue
			}
			if err := c.actualSendGenericMessage(fallback, message); err == nil {
				log.Printf("Sent message using fallback method %s", fallback)
				return nil
			} else {
				log.Printf("Failed to send message using fallback method %s: %v", fallback, err)
			}
		}
		return err
	}
	return nil
}

func InitializeCommProvider() *CommProvider {
	comm := &CommProvider{
		communicators: make(map[string]Communicator),
	}
	if emailHandler, err := email.NewHandler(); err == nil {
		comm.communicators["email"] = emailHandler
		comm.fallbackCommunicators = append(comm.fallbackCommunicators, "email")
	} else {
		log.Printf("Failed to initialize email communicator: %v", err)
	}
	if gotifyHandler, err := gotify.NewClient(); err == nil {
		comm.communicators["gotify"] = gotifyHandler
		comm.fallbackCommunicators = append(comm.fallbackCommunicators, "gotify")
	} else {
		log.Printf("Failed to initialize gotify communicator: %v", err)
	}

	return comm
}
