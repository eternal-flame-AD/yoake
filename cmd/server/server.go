package main

import (
	"flag"
	"log"
	"os"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/internal/apparmor"
	"github.com/eternal-flame-AD/yoake/internal/comm"
	"github.com/eternal-flame-AD/yoake/internal/db"
	"github.com/eternal-flame-AD/yoake/server"
	"github.com/eternal-flame-AD/yoake/server/vault"
	"github.com/eternal-flame-AD/yoake/server/webroot"
)

var (
	flagConfig = flag.String("c", "config.yml", "config file")
)

func init() {
	flag.Parse()
	config.ParseConfig(*flagConfig)

	comm := comm.InitializeCommProvider()
	db, err := db.New(config.Config())
	if err != nil {
		log.Panicf("failed to initialize database: %v", err)
	}
	conf := config.Config()
	for host, handler := range conf.Hosts {
		switch handler {
		case "vault":
			vault.Init(host)
		case "webroot":
			webroot.Init(host, comm, db)
		default:
			log.Panicf("unknown handler for %s: %s", host, handler)
		}
	}
}

func changeHat() {
	profile := config.Config().Listen.AppArmor.Serve
	if profile != "" {
		token, err := apparmor.GetMagicToken()
		if err != nil {
			log.Panicf("failed to get apparmor magic token: %v", err)
		}
		if err := apparmor.ChangeHat(profile, token); err != nil {
			log.Panicf("failed to change apparmor hat: %v", err)
		} else {
			log.Printf("changed apparmor hat to %s", profile)
		}
	}
}

func main() {
	listen := config.Config().Listen
	if listen.Ssl.Use {
		var sslCertBytes, sslKeyBytes []byte
		apparmor.ExecuteInHat(listen.AppArmor.SSL, func() {
			var err error
			sslCertBytes, err = os.ReadFile(listen.Ssl.Cert)
			if err != nil {
				log.Panicf("failed to read ssl cert: %v", err)
			}
			sslKeyBytes, err = os.ReadFile(listen.Ssl.Key)
			if err != nil {
				log.Panicf("failed to read ssl key: %v", err)
			}
		}, true)
		if listen.AppArmor.SSL != "" {
			// defensive programming, try read ssl key
			if _, err := os.ReadFile(listen.Ssl.Key); err == nil {
				log.Panicf("AppArmor profile set for SSL but I could still read %v!", listen.Ssl.Key)
			}
		}

		log.Fatalln(server.Server.StartTLS(listen.Addr, sslCertBytes, sslKeyBytes))
	} else {
		log.Fatalln(server.Server.Start(listen.Addr))
	}
}
