package app

import (
	"doh-go/lib"
	"doh-go/lib/db"
	"doh-go/lib/dns"
	"doh-go/lib/q"
	"encoding/json"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/flamego/flamego"
)

var cwd, _ = os.Getwd()
var StoreDir = path.Join(cwd, "store")
var dnsdb = &db.DB{
	Dir: StoreDir,
	DNS: "223.5.5.5",
}
var ddns = &dns.DNS{
	DB: dnsdb,
}

var file = createLogWriter()

func createLogWriter() io.Writer {
	cwd, _ := os.Getwd()
	storeDir := path.Join(cwd, "store")
	file, _ := os.OpenFile(path.Join(storeDir, "app.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	return file
}

func createApp() *flamego.Flame {
	flamego.SetEnv(flamego.EnvTypeProd)
	app := flamego.NewWithLogger(file)
	app.Use(func(c flamego.Context, logger *log.Logger) {
		now := time.Now().UnixNano()
		c.Next()
		logger.Printf("%s: %s %s %d %s",
			time.Now().Format("2006-01-02 15:04:05"),
			c.Request().Method,
			c.Request().URL,
			c.ResponseWriter().Status(),
			strconv.FormatInt((time.Now().UnixNano()-now)/1e6, 10)+"ms",
		)
	})
	return app
}

func Get() *flamego.Flame {

	ddns.Init()

	app := createApp()
	app.Get("/test", func(c flamego.Context) string {
		host := c.Query("domain", "")
		base := c.Query("dns", "")
		result := lib.Test(host, base)
		buf, _ := json.Marshal(result)
		return string(buf)
	})

	app.Get("/dns-query", func(c flamego.Context, logger *log.Logger) {
		now := time.Now().UnixNano()
		dnsQuery := c.Query("dns", "")
		host := q.ParseDomain(dnsQuery)
		cate, ip := ddns.GetDNS(host)
		ms := (time.Now().UnixNano() - now) / 1e6
		resp, _ := q.ResolveOrigin(dnsQuery, ip)
		defer func() {
			_ = resp.Close()
		}()
		w := c.ResponseWriter()
		_, _ = w.Write(resp.Bytes())
		logger.Println(host, cate, ip, strconv.FormatInt(ms, 10)+"ms")
	})

	app.Get("/query", func(c flamego.Context, logger *log.Logger) string {
		now := time.Now().UnixNano()
		host := c.Query("domain", "")
		cate, ip := ddns.GetDNS(host)
		resp, _ := q.Resolve(host, ip)
		ms := (time.Now().UnixNano() - now) / 1e6
		logger.Println(host, cate, ip, strconv.FormatInt(ms, 10)+"ms")
		return resp
	})

	return app
}
