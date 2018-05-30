package main

import (
	"net/http"
	"log"
)

func main() {
	loadConfig()
	establishRedisConnection()
	initRouting("/stripe", map[string]Endpoint{
		http.MethodGet: mStripeGet,
		http.MethodPost: mStripeCreate,
	}, false)
	initRouting("/app", map[string]Endpoint{http.MethodGet: mInfoConfig}, true)
	initRouting("/", map[string]Endpoint{http.MethodGet: mInfoPing}, true)
	log.Fatal(http.ListenAndServe(":8080", nil))
}