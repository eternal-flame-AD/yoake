package model

import (
	"github.com/labstack/echo/v4"
)

type CommMethod interface {
	SupportedMIME() []string
	SendGenericMessage(message GenericMessage) error
}

type CommMethodWithRoute interface {
	RegisterRoute(g *echo.Group)
}

type Communicator interface {
	GetMethod(method string) CommMethod
	GetMethodsByMIME(mime string) []CommMethod
	SendGenericMessage(preferredMethod string, message GenericMessage, force bool) error
}
