package auth

import (
	"log"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yubigo"
)

var yubiAuth *yubigo.YubiAuth

func yubiAuthLazyInit() {
	if yubiAuth != nil {
		return
	}
	conf := config.Config()
	if conf.Auth.Yubikey.ClientId != "" {
		if a, err := yubigo.NewYubiAuth(conf.Auth.Yubikey.ClientId, conf.Auth.Yubikey.ClientKey); err != nil {
			log.Panicf("failed to load yubigo: %v", err)
		} else {
			yubiAuth = a
		}
	}
}
