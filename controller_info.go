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
	Version string `json:"version"`
	Name string `json:"name"`
	URL string `json:"url"`
	Color string `json:"color"`
	PasswordPolicy string `json:"password-policy"`
	RequireApiKey bool `json:"require-api-key"`
	RequireApiKeyForPostOnly bool `json:"require-api-key-for-post-only"`
}

func mInfoConfig(r *http.Request, c Context) (int, JsonResponse, error) {
	return 200, InfoConfigResponse{
		config.Version,
		config.PublicName,
		config.PublicURL,
		config.PublicColor,
		config.PasswordPolicy,
		config.RequireApiKey,
		config.RequireApiKeyForPostOnly,
	}, nil
}