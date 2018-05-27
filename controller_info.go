package main

import (
	"net/http"
	"time"
	"log"
)

type InfoPingResponse struct {
	Timestamp int64 `json:"timestamp"`
	HttpStatus bool `json:"http-status"`
	RedisStatus bool `json:"redis-status"`
}

func mInfoPing(r *http.Request, c Context) (int, JsonResponse, error) {
	log.Print("Ping was called by " + r.UserAgent())
	return 200, InfoPingResponse{time.Now().Unix(), true, testRedisConnection()}, nil
}

type InfoConfigResponse struct {
}

func mInfoConfig(r *http.Request, c Context) (int, JsonResponse, error) {
	return 200, InfoConfigResponse{}, nil
}