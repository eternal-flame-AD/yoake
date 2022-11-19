package canvaslms

import (
	"log"
	"strings"
	"time"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/internal/comm/model"
)

func (h *Handler) SendGradeMessage(conf config.CanvasLMSMessage) error {
	if conf.Template == "" {
		return nil
	}
	mime := "text/plain+text/template"
	if strings.HasSuffix(conf.Template, ".html") && strings.HasPrefix(conf.Template, "@") {
		mime = "text/html+html/template"
	}
	grades, err := h.sortResponse(GraphSubmissionCompareByGradeTime)
	if err != nil {
		return err
	}

	if err := h.comm.SendGenericMessage(conf.Comm, model.GenericMessage{
		Subject: conf.Subject,
		Body:    conf.Template,
		MIME:    mime,
		Context: GetGradesResponse{
			Grades:      grades,
			LastRefresh: h.respCache.requestTime.Format(time.RFC3339),
		},
	}, false); err != nil {
		log.Printf("error sending grade message: %v", err)
		return err
	}
	return nil
}
