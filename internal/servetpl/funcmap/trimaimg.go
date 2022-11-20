package funcmap

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func TrimaImg(path string, retType string) (string, error) {
	url := "https://yumechi.jp/img/trima/" + path
	download := func() ([]byte, error) {
		resp, err := httpClient.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	switch retType {
	case "url":
		return url, nil
	case "data":
		data, err := download()
		if err != nil {
			return "", err
		}
		return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data), nil
	case "raw":
		data, err := download()
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	return "", fmt.Errorf("unknown return type: %s", retType)
}
