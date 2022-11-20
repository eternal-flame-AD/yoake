package tglogin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/internal/db"
	"github.com/labstack/echo/v4"
)

type LoginCallbackForm struct {
	ID        uint64 `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserName  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  uint64 `json:"auth_date"`
}

type verifyForm map[string]interface{}

func sha256Str(input string) []byte {
	h := sha256.New()
	h.Write([]byte(input))
	return h.Sum(nil)
}

func VerifyLoginCallback(c echo.Context, database db.DB) (*LoginCallbackForm, error) {
	txn := database.NewTransaction(true)
	defer txn.Discard()
	tgToken := config.Config().Comm.Telegram.Token
	if tgToken == "" {
		return nil, fmt.Errorf("telegram token not set")
	}
	verifyForm := make(verifyForm)
	if err := c.Bind(&verifyForm); err != nil {
		return nil, err
	}
	verifyKeys := make([]string, 0, len(verifyForm))
	for k := range verifyForm {
		if k != "hash" {
			verifyKeys = append(verifyKeys, k)
		}
	}
	sort.Strings(verifyKeys)
	hmac := hmac.New(sha256.New, sha256Str(tgToken))
	for i, k := range verifyKeys {
		if i != 0 {
			hmac.Write([]byte("\n"))
		}
		switch val := verifyForm[k].(type) {
		case string:
			fmt.Fprintf(hmac, "%s=%s", k, val)
		case float64:
			fmt.Fprintf(hmac, "%s=%v", k, uint64(val))
		default:
			return nil, fmt.Errorf("invalid verify form, unexpected type %T", val)
		}
	}
	if fmt.Sprintf("%x", hmac.Sum(nil)) != verifyForm["hash"] {
		return nil, echo.NewHTTPError(403, "invalid hash")
	}

	form := new(LoginCallbackForm)
	remarshaled, err := json.Marshal(verifyForm)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(remarshaled, form); err != nil {
		return nil, err
	}

	lastAuthDateDBKey := fmt.Sprintf("auth_telegram_last_auth_date_%d", form.ID)
	var dbLastAuthDate uint64
	if err := db.GetJSON(txn, []byte(lastAuthDateDBKey), &dbLastAuthDate); db.IsNotFound(err) {
		dbLastAuthDate = 0
	} else if err != nil {
		return nil, err
	}
	if form.AuthDate <= dbLastAuthDate || time.Since(time.Unix(int64(form.AuthDate), 0)) > 10*time.Minute {
		return nil, echo.NewHTTPError(403, "authentication payload expired")
	}
	if err := db.SetJSON(txn, []byte(lastAuthDateDBKey), form.AuthDate); err != nil {
		return nil, err
	}

	return form, nil
}
