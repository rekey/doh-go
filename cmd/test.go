package main

import (
	"doh-go/lib/db"
	"doh-go/lib/dns"
	"log"
	"os"
	"path"
)

func main() {
	cwd, _ := os.Getwd()
	StoreDir := path.Join(cwd, "store")
	dnsdb := &db.DB{
		Dir: StoreDir,
		DNS: "223.5.5.5",
	}
	ddns := &dns.DNS{
		DB: dnsdb,
	}
	ddns.Init()

	log.Println(ddns.GetDNS("stun.chat.bilibili.com."))
	log.Println(ddns.GetDNS("youtube.com"))
}
