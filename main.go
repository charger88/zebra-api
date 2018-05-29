package main

import (
	"net/http"
	"log"
)

func main() {
	loadConfig()
	establishRedisConnection()
	initRouting("/stripe-create", mStripeCreate, false)
	initRouting("/stripe-get", mStripeGet, false)
	initRouting("/app-config", mInfoConfig, true)
	initRouting("/", mInfoPing, true)
	log.Fatal(http.ListenAndServe(":8080", nil))
}