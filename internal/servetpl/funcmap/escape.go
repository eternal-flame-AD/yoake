package funcmap

import (
	"bytes"
	"encoding/xml"
	"net/url"
)

func EscapeXML(s string) (string, error) {
	var b bytes.Buffer
	if err := xml.EscapeText(&b, []byte(s)); err != nil {
		return "", err
	}
	return b.String(), nil
}

func EscapeQuery(s string) string {
	return url.QueryEscape(s)
}
