package gotify

type Message struct {
	Message  string `json:"message"`
	Title    string `json:"title"`
	Priority int    `json:"priority"`

	Extras struct {
		ClientDisplay struct {
			ContentType string `json:"contentType"`
		} `json:"client::display,omitempty"`
		ClientNotification struct {
			Click struct {
				URL string `json:"url"`
			} `json:"click,omitempty"`
		} `json:"client::notification,omitempty"`
	} `json:"extras"`
}

type Health struct {
	Database string `json:"database"`
	Health   string `json:"health"`
}

type Version struct {
	BuildDate string `json:"buildDate"`
	Commit    string `json:"commit"`
	Version   string `json:"version"`
}
