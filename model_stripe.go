package main

import (
	"encoding/json"
	"errors"
	"time"
	"math"
)

type Stripe struct {
	Key string `json:"key"`
	Data string `json:"data"`
	Expiration int `json:"expiration"`
	Password string `json:"password"`
	Burn bool `json:"burn"`
}

func loadStripeFromRedis(key string) (Stripe, error) {
	stripe := Stripe{}
	resp := redisClient.Cmd("GET", "STRIPE:" + key)
	if resp.Err != nil {
		return stripe, resp.Err
	}
	dat, err := resp.Bytes()
	if err != nil {
		return stripe, nil
	}
	err = json.Unmarshal(dat, &stripe)
	if err != nil {
		return stripe, err
	}
	return stripe, nil
}

func createStripeInRedis(data string, expiration int, mode string, password string, burn bool, attempts int) (Stripe, error) {
	var chars string
	if mode == "uppercase-lowercase-digits" {
		chars = randomStringUcLcD
	} else if mode == "uppercase-digits" {
		chars = randomStringUcD
	} else if mode == "uppercase" {
		chars = randomStringUc
	} else if mode == "digits" {
		chars = randomStringD
	} else {
		chars = randomStringUcLcD
	}
	key := randomString(getRequiredKeyLength(chars, expiration), chars)
	stripe := Stripe{key, data, int(time.Now().Unix()) + expiration, password, burn}
	dat, err := json.Marshal(stripe)
	if err != nil {
		return Stripe{}, err
	}
	resp := redisClient.Cmd("SET", "STRIPE:" + key, dat, "EX", expiration, "NX")
	if resp.Err != nil {
		return Stripe{}, resp.Err
	}
	val, _ := resp.Str()
	if val != "OK" {
		if attempts < 3 {
			return createStripeInRedis(data, expiration, mode, password, burn, attempts + 1)
		} else {
			return Stripe{}, errors.New("this code is already in use")
		}
	}
	return stripe, nil
}

func getRequiredKeyLength(chars string, expiration int) int {
	hours := math.Max(float64(expiration) / 3600.0, 0.25)
	requiredKeysTotalNumber := float64(config.AppropriateChanceToGuess) * float64(config.ExpectedStripesPerHour) * float64(config.AllowedBadAttempts) * hours
	keyLength := math.Log(float64(requiredKeysTotalNumber)) / math.Log(float64(len(chars)))
	return int(math.Max(math.Ceil(keyLength), float64(config.MinimalKeyLength)))
}