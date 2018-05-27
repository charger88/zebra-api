package main

import (
	"net/http"
	"log"
)

func main() {
	loadConfig()
	establishRedisConnection()
	initRouting("/stripe-create", mStripeCreate)
	initRouting("/stripe-get", mStripeGet)
	initRouting("/app-config", mInfoConfig)
	initRouting("/", mInfoPing)
	log.Fatal(http.ListenAndServe(":8080", nil))
}