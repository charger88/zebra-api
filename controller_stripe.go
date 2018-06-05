package main

import (
	"net/http"
	"encoding/json"
	"errors"
	"crypto/md5"
	"fmt"
)

type StripeCreateResponse struct {
	Key string `json:"key"`
	Expiration int `json:"expiration"`
}

type StripeCreateRequest struct {
	Data string `json:"data"`
	Burn bool `json:"burn"`
	Expiration int `json:"expiration"`
	Mode string `json:"mode"`
	Password string `json:"password"`
}

func getStripeCreateRequest(r *http.Request) (StripeCreateRequest, error) {
	var sc StripeCreateRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&sc)
	defer r.Body.Close()
	// TODO Validate, including option "stripe-password-policy"
	if sc.Password != "" {
		sc.Password = fmt.Sprintf("%x", md5.Sum([]byte(sc.Password)))
	}
	return sc, err
}

func mStripeCreate(r *http.Request, c Context) (int, JsonResponse, error) {
	req, err := getStripeCreateRequest(r)
	if err != nil {
		return 503, StripeCreateResponse{}, err
	}
	_, rateLimitStatus := rateLimit("mStripeCreate", getIp(r), config.AllowedSharesNumberInPeriod, config.AllowedSharesPeriod)
	if !rateLimitStatus {
		return 429, StripeCreateResponse{}, errors.New( fmt.Sprintf("try again in %d seconds", config.AllowedSharesPeriod))
	}
	stripe, err := createStripeInRedis(req.Data, req.Expiration, req.Mode, req.Password, req.Burn, 0)
	if err != nil {
		return 503, StripeCreateResponse{}, err
	}
	return 201, StripeCreateResponse{stripe.Key, stripe.Expiration}, nil
}

type StripeGetResponse struct {
	Data string `json:"data"`
	Expiration int `json:"expiration"`
	Burn bool`json:"burn"`
}

type StripeGetRequest struct {
	Key string `json:"key"`
	Password string `json:"password"`
}

func getStripeGetRequest(r *http.Request) (StripeGetRequest, error) {
	var err error
	var sc StripeGetRequest
	sc.Key = r.URL.Query().Get("key")
	sc.Password = r.URL.Query().Get("password")
	if sc.Password != "" {
		sc.Password = fmt.Sprintf("%x", md5.Sum([]byte(sc.Password)))
	}
	// TODO Validate
	return sc, err
}

func mStripeGet(r *http.Request, c Context) (int, JsonResponse, error) {
	req, err := getStripeGetRequest(r)
	if err != nil {
		return 503, StripeGetResponse{}, err
	}
	rateLimitKey, rateLimitStatus := rateLimit("mStripeGet", getIp(r), config.AllowedBadAttempts, 60)
	if !rateLimitStatus {
		return 429, StripeCreateResponse{}, errors.New( fmt.Sprintf("try again in %d seconds", 60))
	}
	stripe, err := loadStripeFromRedis(req.Key)
	if err != nil {
		return 503, StripeGetResponse{}, err
	}
	if stripe.Burn {
		resp := redisClient.Cmd("SET", "BURN:" + stripe.Key, 1, "EX", 3600, "NX")
		if resp.Err != nil {
			return 503, StripeGetResponse{}, resp.Err
		}
		val, _ := resp.Str()
		if val != "OK" {
			return 404, StripeGetResponse{}, errors.New("key not found")
		}
	}
	if stripe.Key == "" {
		return 404, StripeGetResponse{}, errors.New("key not found")
	}
	if stripe.Key != req.Key {
		return 503, StripeGetResponse{}, nil
	}
	if stripe.Password != req.Password {
		return 401, StripeGetResponse{}, errors.New("incorrect password")
	}
	if stripe.Burn {
		deleteRedisKey("STRIPE:" + stripe.Key)
		deleteRedisKey("BURN:" + stripe.Key)
	}
	deleteRedisKey(rateLimitKey)
	return 200, StripeGetResponse{stripe.Data, stripe.Expiration, stripe.Burn}, nil
}