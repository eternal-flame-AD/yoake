package gotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/internal/comm/model"
	"github.com/eternal-flame-AD/yoake/internal/echoerror"
)

type Client struct {
	conf          *config.CommGotify
	httpClient    *http.Client
	baseURL       url.URL
	serverVersion Version
}

func NewClient() (*Client, error) {
	conf := config.Config().Comm.Gotify
	baseURL, err := url.Parse(conf.BaseURL)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := httpClient.Get(baseURL.ResolveReference(urlVersion).String())
	if err != nil {
		return nil, fmt.Errorf("failed to obtain gotify version: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gotify returned with %d", resp.StatusCode)
	}
	var version Version
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return nil, fmt.Errorf("failed to decode gotify version: %w", err)
	}
	resp, err = httpClient.Get(baseURL.ResolveReference(urlHealth).String())
	if err != nil {
		return nil, fmt.Errorf("failed to obtain gotify health: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gotify returned with %d", resp.StatusCode)
	}
	var health Health
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode gotify health: %w", err)
	}

	return &Client{
		conf:          &conf,
		httpClient:    httpClient,
		baseURL:       *baseURL,
		serverVersion: version,
	}, nil
}

func (c *Client) SendGenericMessage(gmsg model.GenericMessage) error {
	msg := Message{
		Message:  gmsg.Body,
		Title:    gmsg.Subject,
		Priority: gmsg.Priority + 5,
	}
	msg.Extras.ClientDisplay.ContentType = gmsg.MIME
	return c.SendMessage(msg)
}

func (c *Client) SupportedMIME() []string {
	return []string{"text/plain", "text/markdown"}
}

func (c *Client) SendMessage(msg Message) error {

	reader, writer := io.Pipe()
	e := json.NewEncoder(writer)
	req, err := http.NewRequest("POST", c.baseURL.ResolveReference(urlMessage).String(), reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", c.conf.AppToken)
	go func() {
		defer writer.Close()
		e.Encode(msg)
	}()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return echoerror.NewHttp(resp.StatusCode, fmt.Errorf("gotify returned with %d", resp.StatusCode))
	}
	return nil
}
