package main

import (
	"encoding/json"
	"errors"
	"time"
	"github.com/mediocregopher/radix.v2/redis"
)

type GuestKey struct {
	Key string `json:"key"`
	Expiration int `json:"expiration"`
}

func createGuestKeyInRedis(redisClient *redis.Client, attempts int) (GuestKey, error) {
	key := randomString(32, randomStringUcLcD)
	guestKey := GuestKey{key, int(time.Now().Unix()) + config.GuestOneTimeKeyExpirationTime}
	dat, err := json.Marshal(guestKey)
	if err != nil {
		return GuestKey{}, err
	}

	resp := redisClient.Cmd("SET", config.RedisKeyPrefix + "GUEST_KEY:" + key, dat, "EX", config.GuestOneTimeKeyExpirationTime, "NX")
	if resp.Err != nil {
		return GuestKey{}, resp.Err
	}
	val, _ := resp.Str()
	if val != "OK" {
		if attempts < 3 {
			return createGuestKeyInRedis(redisClient, attempts + 1)
		} else {
			return GuestKey{}, errors.New("this guest key value is already in use")
		}
	}
	return guestKey, nil
}
