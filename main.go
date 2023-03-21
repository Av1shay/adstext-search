package main

import (
	"crypto/tls"
	"github.com/Av1shay/adstext-search/search"
	"github.com/Av1shay/adstext-search/server"
	"log"
	"net/http"
	"time"
)

const defaultPort = "8080"

func main() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MaxVersion: tls.VersionTLS12,
		},
	}
	httpClient := &http.Client{Timeout: 30 * time.Second, Transport: tr}
	searchClient := search.NewClient(httpClient)

	serv, err := server.New(searchClient)
	if err != nil {
		log.Fatal("server.New() error", err)
	}

	log.Fatal(http.ListenAndServe(":"+defaultPort, serv))
}
