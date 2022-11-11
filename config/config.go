package config

import (
	"github.com/jinzhu/configor"
	"github.com/labstack/echo/v4/middleware"
)

type C struct {
	Hosts  map[string]string
	Listen struct {
		Addr string
		Ssl  struct {
			Use  bool
			Cert string
			Key  string
		}
	}
	WebRoot struct {
		SiteName   string
		Root       string
		SessionKey string
		SessionDir string
		Secure     *middleware.SecureConfig
		Log        *struct {
			Filter []string
			Indent bool
		}
	}
	Upstream struct {
		Vault struct {
			API string
			UI  string
		}
	}
	Twilio struct {
		AccountSid string
		AuthToken  string
		SkipVerify bool
		BaseURL    string
	}
	Auth struct {
		ValidMinutes int
		Method       struct {
			UserPass struct {
			}
			Yubikey struct {
				ClientId  string
				ClientKey string
				Keys      []struct {
					Name     string
					PublicId string
					Role     string
				}
			}
		}
	}
}

var parsedC C

var c C

func Config() C {
	return c
}

func MockConfig(freshEnv bool, wrapper func(deployedC *C)) {
	if freshEnv {
		c = parsedC
	}
	wrapper(&c)
}

func ParseConfig(files ...string) {
	configor.Load(&parsedC, files...)
	c = parsedC
}
