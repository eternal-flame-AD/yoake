package main

import (
	"flag"
	"log"

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
func main() {
	listen := config.Config().Listen
	if listen.Ssl.Use {
		log.Fatalln(server.Server.StartTLS(listen.Addr, listen.Ssl.Cert, listen.Ssl.Key))
	} else {
		log.Fatalln(server.Server.Start(listen.Addr))
	}
}
