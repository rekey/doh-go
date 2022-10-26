package main

import (
	"doh-go/lib"
	"log"
)

func main() {
	//	service.Store.AddDomain("bilibili.com", "china")
	//	service.Store.AddDomain("douyincdn.com", "china")
	//	service.Store.AddDomain("youtube.com", "gfw")
	//	service.Store.AddDomain("google.com", "gfw")
	//	service.Store.Save()
	//	service.Store.Update()
	//	resp := lib.Resolve("cdn.jsdelivr.net")
	result, _ := lib.Request("https://cdn.jsdelivr.net/gh/rekey/doh-go@main/store/dns.json")
	log.Println(result.String())

}
