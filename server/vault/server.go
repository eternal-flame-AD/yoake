package vault

import (
	"log"
	"net/url"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/server"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Init(hostname string) {
	e := echo.New()

	vaultApiURL, err := url.Parse(config.Config().Upstream.Vault.API)
	if err != nil {
		log.Panicf("vault URL not valid: %s", err)
	}
	e.Group("/v1").Use(middleware.Proxy(middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
		{URL: vaultApiURL},
	})))
	if vaultUiURLStr := config.Config().Upstream.Vault.UI; vaultUiURLStr != "" {
		vaultUiURL, err := url.Parse(vaultUiURLStr)
		if err != nil {
			log.Panicf("vault URL not valid: %s", err)
		}
		if vaultUiURL.Scheme == "file" {
			log.Println(vaultUiURL.Path)
			e.Use(middleware.Rewrite(map[string]string{"^/ui/*": "/$1"}), middleware.StaticWithConfig(middleware.StaticConfig{HTML5: true, Root: vaultUiURL.Path}))
		} else {
			e.Use(middleware.Proxy(middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
				{URL: vaultUiURL},
			})))
		}
	}

	server.RegisterHostname(hostname, &server.Host{Echo: e})
}
