package config

import (
	"log"

	"github.com/jinzhu/configor"
	"github.com/labstack/echo/v4/middleware"
)

type C struct {
	parsed bool
	Hosts  map[string]string
	Listen struct {
		Addr     string
		AppArmor struct {
			Serve string
			SSL   string
		}
		Ssl struct {
			Use  bool
			Cert string
			Key  string
		}
	}
	DB struct {
		Badger DBBadger
	}
	FS      FileStore
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
	Comm      Communication
	CanvasLMS CanvasLMS
	Auth      struct {
		ValidMinutes int
		DevMode      struct {
			GrantAll bool
		}
		Users map[string]struct {
			Password    string
			PublicKeyId []string
			Roles       []string
		}
		Yubikey struct {
			ClientId  string
			ClientKey string
		}
	}
}

var parsedC C

var c C

func Config() C {
	if !c.parsed {
		log.Panicln("Config() called without calling ParseConfig() first")
	}
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
	parsedC.parsed = true
	c = parsedC
}
