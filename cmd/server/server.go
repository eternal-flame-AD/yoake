package main

import (
	"flag"
	"log"

	"github.com/eternal-flame-AD/yoake/config"
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

	conf := config.Config()
	for host, handler := range conf.Hosts {
		switch handler {
		case "vault":
			vault.Init(host)
		case "webroot":
			webroot.Init(host)
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
