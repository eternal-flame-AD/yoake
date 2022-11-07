package twilio

import (
	"bytes"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTwilioValidate(t *testing.T) {
	params := map[string]string{
		"CallSid": "CA1234567890ABCDE",
		"Caller":  "+12349013030",
		"Digits":  "1234",
		"From":    "+12349013030",
		"To":      "+18005551212",
	}
	bodyForm := make(url.Values)
	for k, v := range params {
		bodyForm.Set(k, v)
	}

	req := httptest.NewRequest("POST", "https://mycompany.com/myapp.php?foo=1&bar=2",
		bytes.NewBufferString(bodyForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, req.ParseForm())

	assert.ErrorContains(t, TwilioValidate(req), "no twilio signature")
	req.Header.Set("X-Twilio-Signature", "garbage")
	assert.ErrorContains(t, TwilioValidate(req), "twilio signature verification failed")
	req.Header.Set("X-Twilio-Signature", "0/KCTR6DLpKmkAf8muzZqo1nDgQ=")
	assert.NoError(t, TwilioValidate(req))
}
