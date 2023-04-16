package main

import (
	"crypto/tls"
	"github.com/Av1shay/adstext-search/search"
	"github.com/Av1shay/adstext-search/server"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	defaultPort               = "8080"
	defaultUserAgent          = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
	defaultSearchWorkersCount = 10
)

func main() {
	port, found := os.LookupEnv("PORT")
	if !found {
		port = defaultPort
	}

	userAgent, found := os.LookupEnv("USER_AGENT")
	if !found {
		userAgent = defaultUserAgent
	}
	log.Printf("using user agent %q\n", userAgent)

	searchWorkersCount := defaultSearchWorkersCount
	if searchWorkersCountEnv := os.Getenv("SEARCH_WORKERS_COUNT"); searchWorkersCountEnv != "" {
		workersCount, err := strconv.Atoi(searchWorkersCountEnv)
		if err != nil || workersCount == 0 {
			log.Println("SEARCH_WORKERS_COUNT env must be a valid number > 0, using default workers count")
		} else {
			searchWorkersCount = workersCount
		}
	}
	log.Printf("using %d workers for search\n", searchWorkersCount)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MaxVersion: tls.VersionTLS12,
		},
	}
	httpClient := &http.Client{Timeout: 30 * time.Second, Transport: tr}
	searchClient := search.NewClient(search.Config{UserAgent: userAgent, Workers: searchWorkersCount}, httpClient)
	serv, err := server.New(searchClient)
	if err != nil {
		log.Fatal("server.New() error", err)
	}

	log.Println("starting server...")

	log.Fatal(http.ListenAndServe(":"+port, serv))
}
