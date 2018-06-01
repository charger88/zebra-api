package main

import (
	"time"
	"math/rand"
	"fmt"
	"github.com/mediocregopher/radix.v2/redis"
	"net/http"
)

var randomInitialized = false

const randomStringUcLcD = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const randomStringLcD = "abcdefghijklmnopqrstuvwxyz0123456789"
const randomStringLc = "abcdefghijklmnopqrstuvwxyz"
const randomStringD = "0123456789"

func randomString(n int, chars string) string {
	if !randomInitialized {
		rand.Seed(time.Now().Unix())
		randomInitialized = true
	}
	letterRunes := []rune(chars)
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func rateLimit(prefix string, value string, limit int, period int) (string, bool) {
	var key string
	var resp *redis.Resp
	var redisRespValue string
	for n := 0; n < limit; n++ {
		key = fmt.Sprintf("RATE-LIMIT:%s:%s:%d", prefix, value, n)
		resp = redisClient.Cmd("SET", key, 1, "EX", period, "NX")
		if resp.Err == nil {
			redisRespValue, _ = resp.Str()
			if redisRespValue == "OK" {
				return key, true
			}
		} else {
			return "", false
		}
	}
	return "", false
}

func deleteRateLimitRecord(key string){
	redisClient.Cmd("DEL", key)
}

func getIp(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		ip = r.RemoteAddr
	}
	return ip
}