package config

type CanvasLMS struct {
	Token     string
	Frequency string
	MaxN      string
	Endpoint  string

	SubmissionName string

	Message struct {
		OnUpdate  CanvasLMSMessage
		OnStartup CanvasLMSMessage
	}
}
type CanvasLMSMessage struct {
	Comm     string
	Subject  string
	Template string
}
