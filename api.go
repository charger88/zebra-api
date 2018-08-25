package main

import (
	"net/http"
	"encoding/json"
	"fmt"
	"log"
	"errors"
	"strings"
	"github.com/mediocregopher/radix.v2/redis"
)

type Endpoint func(*http.Request, Context) (int, JsonResponse, error)

type Context struct {
	redisClient *redis.Client
}

type JsonResponse interface {}

type ErrorJsonResponse struct {
	Code int `json:"http-code"`
	Message string `json:"error-message"`
}

var routeResources = map[string][]string{}

func initRouting(resource string, methods map[string]Endpoint, public bool) {
	var methodsList []string
	for m := range methods {
		methodsList = append(methodsList, m)
	}
	methodsList = append(methodsList, "OPTIONS")
	routeResources[resource] = methodsList
	http.HandleFunc(resource, func(w http.ResponseWriter, r *http.Request) {
		var status int
		var response JsonResponse
		var err error
		if !public && (r.Method != http.MethodOptions) {
			status, err = auth(r)
		}
		var redisClient *redis.Client
		var redisRequired = (resource == "/stripe") && (r.Method != "OPTIONS")
		if (err == nil) && redisRequired {
			redisClient, err = testRedisConnectionAndGetClient(false)
			if err != nil {
				establishRedisConnection(false)
				redisClient, err = testRedisConnectionAndGetClient(false)
				if err != nil {
					status = 503
					log.Print(err)
					err = errors.New("service temporary unavailable")
				}
			}
		}
		if err == nil {
			if methods[r.Method] != nil || r.Method == http.MethodOptions {
				if methods[r.Method] != nil {
					status, response, err = methods[r.Method](r, Context{redisClient})
				} else {
					status = 200
				}
				var allowedMethods []string
				for k := range methods {
					allowedMethods = append(allowedMethods, k)
				}
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ","))
			} else {
				status = 405
				err = errors.New("method not allowed")
			}
			if redisRequired {
				redisClient.Close()
			}
		}
		sendResponse(status, response, err, w)
	})
}

func auth(r *http.Request) (int, error) {
	var status int
	var err error
	requireKey := config.RequireApiKey && (!config.RequireApiKeyForPostOnly || config.RequireApiKeyForPostOnly && (r.Method != http.MethodGet))
	apiKey := r.Header.Get("X-Api-Key")
	if requireKey && (apiKey == "") {
		status = 401
		err = errors.New("header X-Api-Key is required")
	} else if (apiKey != "") && !config.isApiKeyEnabled(apiKey) {
		status = 401
		err = errors.New("key from header X-Api-Key is not enabled")
	}
	return status, err
}

func sendResponse(status int, resp JsonResponse, err error, w http.ResponseWriter) {
	allowedHeaders := []string{"Content-type"}
	if config.RequireApiKey {
		allowedHeaders = append(allowedHeaders, "X-Api-Key")
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ","))
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(status)
	if status >= 400 {
		var message string
		if status >= 500 {
			if err != nil {
				log.Print("Runtime error: " + err.Error())
			} else {
				log.Print("Runtime error: unknown")
			}
			message = "Internal Server Error"
		} else {
			if err != nil {
				message =  err.Error()
			} else {
				message = "Unknown error"
			}
		}
		resp = ErrorJsonResponse{status, message}
	}
	response, _ := json.Marshal(resp)
	fmt.Fprint(w, string(response))
}