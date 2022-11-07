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
	if conf.Auth.Method.Yubikey.ClientId != "" {
		if a, err := yubigo.NewYubiAuth(conf.Auth.Method.Yubikey.ClientId, conf.Auth.Method.Yubikey.ClientKey); err != nil {
			log.Panicf("failed to load yubigo: %v", err)
		} else {
			yubiAuth = a
		}
	}
}
