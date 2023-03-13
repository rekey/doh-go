package main

import (
	"doh-go/app"
	"log"
)

func main() {
	server := app.Get()
	log.Println("app", "lister", 54413)
	server.Run(54413)
}
