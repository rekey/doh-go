package main

import (
	"doh-go/service"
	"log"
)

func main() {
	//	service.Store.AddDomain("bilibili.com", "china")
	//	service.Store.AddDomain("douyincdn.com", "china")
	//	service.Store.AddDomain("youtube.com", "gfw")
	//	service.Store.AddDomain("google.com", "gfw")
	//	service.Store.Save()
	//	service.Store.Update()
	log.Println(service.Store.GetDNS("s1.bilivideo.com"))
	log.Println(service.Store.GetDNS("s1.hdslb.com"))
}
