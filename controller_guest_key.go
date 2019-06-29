package main

import (
	"net/http"
	"errors"
)

type GuestKeyCreateResponse struct {
	Key string `json:"key"`
	Expiration int `json:"expiration"`
}

func mGuestKeyCreate(r *http.Request, c Context) (int, JsonResponse, error) {
	if !config.GuestOneTimeKey {
		extendedLog(r, "Guest key feature is not enabled")
		return 404, GuestKeyCreateResponse{}, errors.New("feature is not enabled")
	}
	guestKey, err := createGuestKeyInRedis(c.redisClient, 0)
	if err != nil {
		extendedLog(r, "Guest key was not saved in Redis: " + err.Error())
		return 503, GuestKeyCreateResponse{}, err
	}
	extendedLog(r, "stripe " + guestKey.Key + " was created")
	return 201, GuestKeyCreateResponse{guestKey.Key, guestKey.Expiration}, nil
}
