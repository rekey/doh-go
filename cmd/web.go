package main

import (
	"doh-go/service"
	"encoding/base64"
	"github.com/flamego/flamego"
	"github.com/levigross/grequests"
	"github.com/miekg/dns"
	"log"
	"strings"
	"time"
)

func parseDomain(dnsQuery string) string {
	msg := dns.Msg{}
	pack, _ := base64.RawURLEncoding.DecodeString(dnsQuery)
	_ = msg.Unpack(pack)
	domain := msg.Question[0].Name
	return domain
}

func main() {
	flamego.SetEnv(flamego.EnvTypeProd)
	app := flamego.Classic()
	app.Get("/dns-query", func(c flamego.Context, logger *log.Logger) {
		dnsQuery := c.Query("dns", "")
		domain := parseDomain(dnsQuery)
		arr := dns.SplitDomainName(domain)
		host := service.Store.GetDNS(strings.Join(arr, "."))
		url := "https://" + host + c.Request().URL.String()
		resp, _ := grequests.Get(url, &grequests.RequestOptions{})
		logger.Printf("%s: %s %s %s",
			time.Now().Format("2006-01-02 15:04:05"),
			"Query",
			domain,
			host,
		)
		w := c.ResponseWriter()
		_, _ = w.Write(resp.Bytes())
	})
	app.Run(54413)
}
