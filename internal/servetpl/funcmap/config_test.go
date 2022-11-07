package funcmap

import (
	"github.com/eternal-flame-AD/yoake/config"
)

func init() {
	config.ParseConfig("../../../config-test.yml")
	config.MockConfig(true, func(config *config.C) {
		config.Twilio.AuthToken = "12345"
		config.Twilio.SkipVerify = false
	})
}
