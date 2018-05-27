package main

import (
	"net/http"
	"encoding/json"
	"errors"
)

type StripeCreateResponse struct {
	Key string `json:"key"`
	Expiration int `json:"expiration"`
}

type StripeCreateRequest struct {
	Data string `json:"data"`
	Expiration int `json:"expiration"`
	Mode string `json:"mode"`
}

func getStripeCreateRequest(r *http.Request) (StripeCreateRequest, error) {
	var sc StripeCreateRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&sc)
	defer r.Body.Close()
	// TODO Validate
	return sc, err
}

func mStripeCreate(r *http.Request, c Context) (int, JsonResponse, error) {
	req, err := getStripeCreateRequest(r)
	if err != nil {
		return 503, StripeCreateResponse{}, err
	}
	stripe, err := createStripeInRedis(req.Data, req.Expiration, req.Mode)
	if err != nil {
		return 503, StripeCreateResponse{}, err
	}
	return 201, StripeCreateResponse{stripe.Key, stripe.Expiration}, nil
}

type StripeGetResponse struct {
	Data string `json:"data"`
	Expiration int `json:"expiration"`
}

type StripeGetRequest struct {
	Key string `json:"key"`
}

func getStripeGetRequest(r *http.Request) (StripeGetRequest, error) {
	var sc StripeGetRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&sc)
	defer r.Body.Close()
	// TODO Validate
	return sc, err
}

func mStripeGet(r *http.Request, c Context) (int, JsonResponse, error) {
	req, err := getStripeGetRequest(r)
	if err != nil {
		return 503, StripeGetResponse{}, err
	}
	stripe, err := loadStripeFromRedis(req.Key)
	if err != nil {
		return 503, StripeGetResponse{}, err
	}
	if stripe.Key == "" {
		return 404, StripeGetResponse{}, errors.New("key not found")
	}
	if stripe.Key != req.Key {
		return 503, StripeGetResponse{}, nil
	}
	return 200, StripeGetResponse{stripe.Data, stripe.Expiration}, nil
}