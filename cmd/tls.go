package main

import (
	"doh-go/app"
	"github.com/kabukky/httpscerts"
	"log"
	"net/http"
	"path"
)

func main() {
	certPath := path.Join(app.StoreDir, "cert.pem")
	keyPath := path.Join(app.StoreDir, "key.pem")
	// Check if the cert files are available.
	err := httpscerts.Check(certPath, keyPath)
	log.Println(err)
	// If they are not available, generate new ones.
	if err != nil {
		err = httpscerts.Generate(certPath, keyPath, "127.0.0.1")
		if err != nil {
			log.Fatal("Error: Couldn't create https certs.")
		}
	}
	server := app.Get()
	log.Println(server)
	http.Handle("/", server)
	err = http.ListenAndServeTLS(":443", certPath, keyPath, nil)
	log.Println(err)
}
