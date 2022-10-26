package main

import (
	"doh-go/app"
)

func main() {
	server := app.Get()
	server.Run(54413)
}
