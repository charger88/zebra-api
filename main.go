package main

import (
	"net/http"
	"math/rand"
	"log"
	"time"
)

func main() {
	loadConfig()
	log.Print("Starting Zebra API v. " + config.Version)
	reloadConfig()
	establishRedisConnection(true)
	rand.Seed(time.Now().UTC().UnixNano())
	initRouting("/stripe", map[string]Endpoint{
		http.MethodGet: mStripeGet,
		http.MethodPost: mStripeCreate,
		http.MethodDelete: mStripeDelete,
	}, false)
	initRouting("/ping", map[string]Endpoint{http.MethodGet: mInfoPing}, true)
	initRouting("/config", map[string]Endpoint{http.MethodGet: mInfoConfig}, true)
	initRouting("/", map[string]Endpoint{http.MethodGet: mInfoRoutes}, true)
	log.Print("Started Zebra API v. " + config.Version)
	log.Fatal(http.ListenAndServe(config.HttpInterface + ":" + config.HttpPort, nil))
}