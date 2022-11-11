package comm

import (
	"github.com/eternal-flame-AD/yoake/internal/comm/model"
	"github.com/labstack/echo/v4"
)

type Communicator interface {
	SupportedMIME() []string
	SendGenericMessage(message model.GenericMessage) error
}

type CommunicatorWithRoute interface {
	RegisterRoute(g *echo.Group)
}
