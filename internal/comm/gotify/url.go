package gotify

import "net/url"

var (
	urlMessage = urlMustParse("/message")
	urlHealth  = urlMustParse("/health")
	urlVersion = urlMustParse("/version")
)

func urlMustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
