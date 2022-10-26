package main

import (
	"doh-go/lib"
	"encoding/base64"
	"github.com/flamego/flamego"
	"github.com/levigross/grequests"
	"github.com/miekg/dns"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var cwd, _ = os.Getwd()
var storeDir = path.Join(cwd, "store")
var store = lib.Store{
	Dir: storeDir,
}

func parseDomain(dnsQuery string) string {
	msg := dns.Msg{}
	pack, _ := base64.RawURLEncoding.DecodeString(dnsQuery)
	_ = msg.Unpack(pack)
	domain := msg.Question[0].Name
	return domain
}

func createLogWriter() io.Writer {
	cwd, _ := os.Getwd()
	storeDir := path.Join(cwd, "store")
	file, _ := os.OpenFile(path.Join(storeDir, "app.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	return file
}

func createApp() *flamego.Flame {
	flamego.SetEnv(flamego.EnvTypeProd)
	file := createLogWriter()
	return flamego.NewWithLogger(file)
}

func main() {
	store.Init(createLogWriter())
	store.Update()
	go func() {
		for {
			time.Sleep(time.Hour * 12)
			store.Update()
		}
	}()
	app := createApp()
	app.Get("/dns-query", func(c flamego.Context, logger *log.Logger) {
		dnsQuery := c.Query("dns", "")
		domain := parseDomain(dnsQuery)
		arr := dns.SplitDomainName(domain)
		host := store.GetDNS(strings.Join(arr, "."))
		url := "https://" + host + c.Request().URL.String()
		now := time.Now().UnixNano()
		resp, _ := grequests.Get(url, &grequests.RequestOptions{})
		defer func() {
			_ = resp.Close()
			logger.Printf("%s: %s %s %s",
				time.Now().Format("2006-01-02 15:04:05"),
				domain,
				host,
				strconv.FormatInt((time.Now().UnixNano()-now)/1e6, 10)+"ms",
			)
		}()
		w := c.ResponseWriter()
		_, _ = w.Write(resp.Bytes())
	})
	app.Run(54413)
}
