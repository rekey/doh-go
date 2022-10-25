package main

import (
	"doh-go/service"
	"log"
)

func main() {
	service.Store.AddDomain("bilibili.com", "china")
	service.Store.AddDomain("douyincdn.com", "china")
	service.Store.AddDomain("youtube.com", "gfw")
	service.Store.AddDomain("google.com", "gfw")
	service.Store.Save()
	log.Println(service.Store.GetDNS("google.com"))
	log.Println(service.Store.GetDNS("youtube.com"))
	log.Println(service.Store.GetDNS("twitter.com"))
	log.Println(service.Store.GetDNS("facebook.com"))
}
