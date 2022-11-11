package model

type GenericMessage struct {
	Subject string `json:"subject" form:"subject" query:"subject"`
	Body    string `json:"body" form:"body" query:"body"`
	MIME    string `json:"mime" form:"mime" query:"mime"`

	Context interface{}
}
