package funcmap

import (
	"bytes"
	"encoding/xml"
)

func EscapeXML(s string) (string, error) {
	var b bytes.Buffer
	if err := xml.EscapeText(&b, []byte(s)); err != nil {
		return "", err
	}
	return b.String(), nil
}
