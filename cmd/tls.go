package main

import (
	"doh-go/app"
	"flag"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/kabukky/httpscerts"
)

func main() {
	domain := flag.String("domain", "", "dns.local")
	flag.Parse()
	ssl := *domain
	if os.Getenv("ssl") != "" {
		ssl = os.Getenv("ssl")
	}
	certPath := path.Join(app.StoreDir, "cert.pem")
	keyPath := path.Join(app.StoreDir, "key.pem")
	// Check if the cert files are available.
	err := httpscerts.Check(certPath, keyPath)
	// If they are not available, generate new ones.
	if err != nil {
		err = httpscerts.Generate(certPath, keyPath, ssl)
		if err != nil {
			log.Fatal("Error: Couldn't create https certs.")
		}
	}
	server := app.Get()
	http.Handle("/", server)
	err = http.ListenAndServeTLS(":443", certPath, keyPath, nil)
	log.Println(err)
}
