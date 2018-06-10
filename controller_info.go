package main

import (
	"net/http"
	"time"
)

type InfoPingResponse struct {
	Timestamp int64 `json:"timestamp"`
}

func mInfoPing(r *http.Request, c Context) (int, JsonResponse, error) {
	return 200, InfoPingResponse{time.Now().Unix()}, nil
}

type InfoConfigResponse struct {
	Name string `json:"name"`
	URL string `json:"url"`
}

func mInfoConfig(r *http.Request, c Context) (int, JsonResponse, error) {
	return 200, InfoConfigResponse{
		config.PublicName,
		config.PublicURL,
	}, nil
}