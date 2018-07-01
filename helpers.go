package main

import (
	"time"
	"math/rand"
	"fmt"
	"github.com/mediocregopher/radix.v2/redis"
	"net/http"
	"log"
	"strings"
	"net"
)

var randomInitialized = false

const randomStringUcLcD = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const randomStringUcD = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const randomStringUc = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const randomStringD = "0123456789"

func randomString(n int, chars string) string {
	if !randomInitialized {
		rand.Seed(time.Now().UTC().UnixNano())
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

func deleteRedisKey(key string) error {
	resp := redisClient.Cmd("DEL", key)
	return resp.Err
}

func getIp(r *http.Request) string {
	ip := r.RemoteAddr
	ip = strings.Split(ip, ":")[0]
	if isTrustedProxy(ip) {
		if r.Header.Get("X-Forwarded-For") != "" {
			ip = r.Header.Get("X-Forwarded-For")
		}
	}
	return ip
}

func isTrustedProxy(ip string) bool {
	ipObj := net.ParseIP(ip)
	var tpNet *net.IPNet
	for _, tp := range config.TrustedProxy {
		_, tpNet, _ = net.ParseCIDR(tp)
		if tpNet.Contains(ipObj) {
			return true
		}
	}
	return false
}

func extendedLog(r *http.Request, message string){
	if config.ExtendedLogs {
		log.Print(r.Method + " " + r.RequestURI + " - " + getIp(r) + " - " + message)
	}
}