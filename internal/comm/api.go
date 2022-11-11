package comm

import (
	"github.com/eternal-flame-AD/yoake/internal/auth"
	"github.com/eternal-flame-AD/yoake/internal/comm/model"
	"github.com/labstack/echo/v4"
)

type CommStatusResponse struct {
	Communicators []struct {
		Method        string
		SupportedMIME []string
	} `json:"communicators"`
}

func (c *CommProvider) RegisterAPIRoute(g *echo.Group) {
	send := g.Group("/send", auth.RequireMiddleware(auth.RoleAdmin))
	{
		send.POST("", func(ctx echo.Context) error {
			var msg model.GenericMessage
			if err := ctx.Bind(&msg); err != nil {
				return err
			}
			if err := c.SendGenericMessage("", msg); err != nil {
				return err
			}
			return nil
		})
		send.POST("/:method", func(ctx echo.Context) error {
			var msg model.GenericMessage
			if err := ctx.Bind(&msg); err != nil {
				return err
			}
			if err := c.SendGenericMessage(ctx.Param("method"), msg); err != nil {
				return err
			}
			return nil
		})
	}

	g.GET("/status", func(ctx echo.Context) error {
		var communicators []struct {
			Method        string
			SupportedMIME []string
		}
		for _, comm := range c.fallbackCommunicators {
			communicators = append(communicators, struct {
				Method        string
				SupportedMIME []string
			}{
				Method:        comm,
				SupportedMIME: c.communicators[comm].SupportedMIME(),
			})
		}
		for key, comm := range c.communicators {
			if !contains(c.fallbackCommunicators, key) {
				communicators = append(communicators, struct {
					Method        string
					SupportedMIME []string
				}{
					Method:        key,
					SupportedMIME: comm.SupportedMIME(),
				})
			}
		}
		return ctx.JSON(200, CommStatusResponse{Communicators: communicators})
	})
}
