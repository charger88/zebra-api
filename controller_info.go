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
	Email string `json:"email"`
	Color string `json:"color"`
	MaxExpirationTime int `json:"max-expiration-time"`
	PasswordPolicy string `json:"password-policy"`
	RequireApiKey bool `json:"require-api-key"`
	RequireApiKeyForPostOnly bool `json:"require-api-key-for-post-only"`
}

func mInfoConfig(r *http.Request, c Context) (int, JsonResponse, error) {
	return 200, InfoConfigResponse{
		config.Version,
		config.PublicName,
		config.PublicURL,
		config.PublicEmail,
		config.PublicColor,
		config.MaxExpirationTime,
		config.PasswordPolicy,
		config.RequireApiKey,
		config.RequireApiKeyForPostOnly,
	}, nil
}

type InfoRoutesResponse struct {
	Routes map[string][]string `json:"routes"`
}

func mInfoRoutes(r *http.Request, c Context) (int, JsonResponse, error) {
	return 200, InfoRoutesResponse{routeResources}, nil
}