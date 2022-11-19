package model

type GenericMessage struct {
	Subject  string `json:"subject" form:"subject" query:"subject"`
	Body     string `json:"body" form:"body" query:"body"`
	MIME     string `json:"mime" form:"mime" query:"mime"`
	Priority int    `json:"priority" form:"priority" query:"priority"`
	ThreadID uint64 `json:"thread_id" form:"thread_id" query:"thread_id"`

	Context interface{}
}
