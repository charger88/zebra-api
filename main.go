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
	initRouting("/ping", map[string]Endpoint{http.MethodGet: mInfoPing}, true)
	initRouting("/", map[string]Endpoint{http.MethodGet: mInfoConfig}, true)
	log.Fatal(http.ListenAndServe(":" + config.HttpPort, nil))
}