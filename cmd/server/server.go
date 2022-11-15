package main

import (
	"flag"
	"log"
	"os"

	"github.com/eternal-flame-AD/go-apparmor/apparmor"
	"github.com/eternal-flame-AD/go-apparmor/apparmor/magic"
	"github.com/eternal-flame-AD/yoake/config"
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

	comm := comm.InitCommunicator()
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

func main() {
	listen := config.Config().Listen

	Server := server.New()
	if listen.Ssl.Use {
		var sslCertBytes, sslKeyBytes []byte

		readCerts := func() {
			var err error
			sslCertBytes, err = os.ReadFile(listen.Ssl.Cert)
			if err != nil {
				log.Panicf("failed to read ssl cert: %v", err)
			}
			sslKeyBytes, err = os.ReadFile(listen.Ssl.Key)
			if err != nil {
				log.Panicf("failed to read ssl key: %v", err)
			}
		}
		magic, err := magic.Generate(nil)
		if err != nil {
			log.Panicf("failed to generate apparmor magic token: %v", err)
		}

		if listen.AppArmor.SSL != "" {
			if err := apparmor.WithHat(listen.AppArmor.SSL, func() uint64 { return magic }, readCerts); err != nil {
				log.Panicf("failed to read ssl cert/key with apparmor hat: %v", err)
			}

			// defensive programming, try read ssl key
			if _, err := os.ReadFile(listen.Ssl.Key); err == nil {
				log.Panicf("AppArmor profile set for SSL but I could still read %v!", listen.Ssl.Key)
			}
		} else {
			readCerts()
		}

		log.Fatalln(Server.StartTLS(listen.Addr, sslCertBytes, sslKeyBytes))
	} else {
		log.Fatalln(Server.Start(listen.Addr))
	}
}
