package main

import (
	"net/http"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"golang.org/x/crypto/bcrypt"
	"math"
)

type StripeCreateResponse struct {
	Key string `json:"key"`
	Expiration int `json:"expiration"`
	OwnerKey string `json:"owner-key"`
}

type StripeCreateRequest struct {
	Data string `json:"data"`
	Burn bool `json:"burn"`
	Expiration int `json:"expiration"`
	Mode string `json:"mode"`
	Password string `json:"password"`
	EncryptedWithClientSidePassword bool `json:"encrypted-with-client-side-password"`
}

func getStripeCreateRequest(r *http.Request) (StripeCreateRequest, error, int) {
	var sc StripeCreateRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&sc)
	defer r.Body.Close()
	if sc.Data == "" {
		return sc, errors.New("data: empty data"), 422
	}
	if len(sc.Data) > config.MaxTextLength {
		fmt.Sprintf("try again in %d seconds", config.AllowedSharesPeriod)
		over := int(math.Ceil((float64(len(sc.Data)) / float64(config.MaxTextLength) - 1) * 100))
		return sc, errors.New(fmt.Sprintf("data: text is %d%% longer than allowed", over)), 422
	}
	if !validateExpiration(sc.Expiration) {
		return sc, errors.New("expiration: validation failed"), 422
	}
	if config.PasswordPolicy == "disabled" {
		if sc.Password != "" {
			return sc, errors.New("password: field is not allowed"), 422
		}
	} else if !validatePassword(sc.Password, config.PasswordPolicy == "required") {
		return sc, errors.New("password: format validation failed"), 422
	}
	if config.EncryptionPasswordPolicy == "disabled" {
		if sc.EncryptedWithClientSidePassword {
			return sc, errors.New("client-side encryption password: not allowed"), 422
		}
	} else if !sc.EncryptedWithClientSidePassword && config.EncryptionPasswordPolicy == "required" {
		return sc, errors.New("client-side encryption password: not provided"), 422
	}
	if sc.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(sc.Password), bcrypt.DefaultCost)
		if err != nil {
			extendedLog(r, "password hashing error: " + err.Error())
		}
		sc.Password = string(hash)
	}
	return sc, err, 400
}

func mStripeCreate(r *http.Request, c Context) (int, JsonResponse, error) {
	req, err, errStatus := getStripeCreateRequest(r)
	if err != nil {
		extendedLog(r, "can't parse request: " + err.Error())
		return errStatus, StripeCreateResponse{}, err
	}
	_, rateLimitStatus := rateLimit(c.redisClient, "mStripeCreate", getIp(r), config.AllowedSharesNumberInPeriod, config.AllowedSharesPeriod)
	if !rateLimitStatus {
		extendedLog(r, "rate limit violation")
		return 429, StripeCreateResponse{}, errors.New( fmt.Sprintf("try again in %d seconds", config.AllowedSharesPeriod))
	}
	stripe, err := createStripeInRedis(c.redisClient, req.Data, req.Expiration, req.Mode, req.Password, req.Burn, req.EncryptedWithClientSidePassword, 0)
	if err != nil {
		extendedLog(r, "stripe was not saved in Redis: " + err.Error())
		return 503, StripeCreateResponse{}, err
	}
	extendedLog(r, "stripe " + stripe.Key + " was created")
	return 201, StripeCreateResponse{stripe.Key, stripe.Expiration, stripe.OwnerKey}, nil
}

type StripeGetResponse struct {
	Key string `json:"key"`
	Data string `json:"data"`
	Expiration int `json:"expiration"`
	Burn bool`json:"burn"`
	EncryptedWithClientSidePassword bool `json:"encrypted-with-client-side-password"`
}

type StripeGetRequest struct {
	Key string `json:"key"`
	Password string `json:"password"`
	CheckKey string `json:"check-key"`
}

func getStripeGetRequest(r *http.Request) (StripeGetRequest, error, int) {
	var err error
	var sc StripeGetRequest
	sc.Key = r.URL.Query().Get("key")
	if !validateKey(sc.Key) {
		return sc, errors.New("key: format validation failed"), 422
	}
	sc.Password = r.URL.Query().Get("password")
	if sc.Password != "" {
		if !validatePassword(sc.Password, true) {
			return sc, errors.New("password: format validation failed"), 422
		}
	}
	sc.CheckKey = r.URL.Query().Get("check-key")
	if sc.CheckKey != "" {
		if !validateKey(sc.CheckKey) {
			return sc, errors.New("check-key: format validation failed"), 422
		}
	}
	return sc, err, 400
}

func mStripeGet(r *http.Request, c Context) (int, JsonResponse, error) {
	req, err, errStatus := getStripeGetRequest(r)
	if err != nil {
		extendedLog(r, "can't parse request: " + err.Error())
		return errStatus, StripeGetResponse{}, err
	}
	rateLimitKey := ""
	rateLimitStatus := false
	if req.CheckKey != "" {
		resp := c.redisClient.Cmd("GET", config.RedisKeyPrefix + "CHECK:" + req.CheckKey)
		if resp.Err == nil {
			dat, err := resp.Str()
			if (err == nil) && (dat == req.Key) {
				err = deleteRedisKey(c.redisClient, "CHECK:" + req.CheckKey)
				if err == nil {
					rateLimitStatus = true
				}
			}
		}
	}
	if !rateLimitStatus {
		rateLimitKey, rateLimitStatus = rateLimit(c.redisClient, "mStripeGet", getIp(r), config.AllowedBadAttempts, 60)
	}
	if !rateLimitStatus {
		extendedLog(r, "rate limit violation")
		return 429, StripeGetResponse{}, errors.New( fmt.Sprintf("try again in %d seconds", 60))
	}
	stripe, err := loadStripeFromRedis(c.redisClient, req.Key)
	if err != nil {
		extendedLog(r, "can't load stripe " + req.Key + " from redis: " + err.Error())
		return 503, StripeGetResponse{}, err
	}
	if stripe.Key == "" {
		extendedLog(r, "stripe not found")
		return 404, StripeGetResponse{}, errors.New("key not found")
	}
	if stripe.Key != req.Key {
		extendedLog(r, "incorrect key for stripe " + stripe.Key)
		return 503, StripeGetResponse{}, nil
	}
	if (stripe.Password != "") || (req.Password != "") {
		err = bcrypt.CompareHashAndPassword([]byte(stripe.Password), []byte(req.Password))
		if err != nil {
			extendedLog(r, "incorrect password for stripe "+stripe.Key+": "+err.Error())
			return 403, StripeGetResponse{}, errors.New("incorrect password")
		}
	}
	if stripe.Burn {
		resp := c.redisClient.Cmd("SET", config.RedisKeyPrefix + "BURN:" + stripe.Key, 1, "EX", 3600, "NX")
		if resp.Err != nil {
			return 503, StripeGetResponse{}, resp.Err
		}
		val, _ := resp.Str()
		if val != "OK" {
			return 404, StripeGetResponse{}, errors.New("key not found")
		}
	}
	if stripe.Burn {
		deleteRedisKey(c.redisClient, "STRIPE:" + stripe.Key)
		deleteRedisKey(c.redisClient, "BURN:" + stripe.Key)
		stripe.Expiration = 0
	}
	if rateLimitKey != "" {
		deleteRedisKey(c.redisClient, rateLimitKey)
	}
	extendedLog(r, "stripe " + stripe.Key + " was retrieved")
	return 200, StripeGetResponse{stripe.Key, stripe.Data, stripe.Expiration, stripe.Burn, stripe.EncryptedWithSpecialPassword}, nil
}

type StripeDeleteResponse struct {
	Success bool `json:"success"`
	CheckKey string `json:"check-key"`
}

type StripeDeleteRequest struct {
	Key string `json:"key"`
	OwnerKey string `json:"owner-key"`
}

func getStripeDeleteRequest(r *http.Request) (StripeDeleteRequest, error, int) {
	var err error
	var sc StripeDeleteRequest
	sc.Key = r.URL.Query().Get("key")
	if !validateKey(sc.Key) {
		return sc, errors.New("key: format validation failed"), 422
	}
	sc.OwnerKey = r.URL.Query().Get("owner-key")
	if !validateKey(sc.Key) {
		return sc, errors.New("owner key: format validation failed"), 422
	}
	return sc, err, 400
}

func mStripeDelete(r *http.Request, c Context) (int, JsonResponse, error) {
	req, err, errStatus := getStripeDeleteRequest(r)
	if err != nil {
		extendedLog(r, "can't parse request: " + err.Error())
		return errStatus, StripeGetResponse{}, err
	}
	rateLimitKey, rateLimitStatus := rateLimit(c.redisClient, "mStripeDelete", getIp(r), 10, 600)
	if !rateLimitStatus {
		extendedLog(r, "rate limit violation")
		return 429, StripeDeleteResponse{}, errors.New( fmt.Sprintf("try again in %d seconds", 600))
	}
	stripe, err := loadStripeFromRedis(c.redisClient, req.Key)
	if err != nil {
		extendedLog(r, "can't load stripe " + req.Key + " from redis: " + err.Error())
		return 503, StripeDeleteResponse{}, err
	}
	if stripe.Key == "" {
		extendedLog(r, "stripe not found " + stripe.Key)
		return 404, StripeDeleteResponse{}, errors.New("key not found")
	}
	if stripe.Key != req.Key {
		extendedLog(r, "incorrect key for stripe " + stripe.Key)
		return 503, StripeDeleteResponse{}, nil
	}
	if stripe.OwnerKey != req.OwnerKey {
		extendedLog(r, "incorrect owner key for stripe " + stripe.Key)
		return 403, StripeDeleteResponse{}, errors.New("incorrect owner key")
	}
	if stripe.Burn {
		resp := c.redisClient.Cmd("SET", config.RedisKeyPrefix + "BURN:" + stripe.Key, 1, "EX", 3600, "NX")
		if resp.Err != nil {
			return 503, StripeDeleteResponse{}, resp.Err
		}
		val, _ := resp.Str()
		if val != "OK" {
			return 404, StripeDeleteResponse{}, errors.New("key not found")
		}
	}
	deleteRedisKey(c.redisClient, "STRIPE:" + stripe.Key)
	deleteRedisKey(c.redisClient, rateLimitKey)
	extendedLog(r, "stripe " + stripe.Key + " was deleted")
	checkKey := randomString(32, randomStringUcLcD)
	resp := c.redisClient.Cmd("SET", config.RedisKeyPrefix + "CHECK:" + checkKey, stripe.Key, "EX", 30, "NX")
	if resp.Err != nil {
		return 200, StripeDeleteResponse{true, ""}, nil
	}
	return 200, StripeDeleteResponse{true, checkKey}, nil
}

func validateKey(key string) bool {
	re := regexp.MustCompile("^([A-Za-z0-9]+)$")
	return re.MatchString(key)
}

func validatePassword(password string, required bool) bool {
	if required && (password == "") {
		return false
	}
	if password != "" {
		re := regexp.MustCompile("^([A-Za-z0-9]+)$")
		return re.MatchString(password)
	} else {
		return true
	}
}

func validateExpiration(expiration int) bool {
	return (expiration >= 10) && (expiration <= config.MaxExpirationTime)
}