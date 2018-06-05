package main

import (
	"net/http"
	"encoding/json"
	"fmt"
	"log"
	"errors"
)

type Endpoint func(*http.Request, Context) (int, JsonResponse, error)

type Context struct {}

type JsonResponse interface {}

type ErrorJsonResponse struct {
	Code int `json:"http-code"`
	Message string `json:"error-message"`
}

func initRouting(resource string, methods map[string]Endpoint, public bool) {
	http.HandleFunc(resource, func(w http.ResponseWriter, r *http.Request) {
		var status int
		var response JsonResponse
		var err error
		if !public {
			status, err = auth(r.Header)
		}
		if !testRedisConnection() {
			establishRedisConnection(false)
			if !testRedisConnection() {
				status = 503
				err = errors.New("service temporary unavailable")
			}
		}
		if err == nil {
			if methods[r.Method] != nil {
				status, response, err = methods[r.Method](r, Context{})
			} else {
				status = 405
				err = errors.New("method not allowed")
			}
		}
		sendResponse(status, response, err, w)
	})
}

func auth(h http.Header) (int, error) {
	var status int
	var err error
	apiKey := h.Get("X-Api-Key")
	if config.RequireApiKey && (apiKey == "") {
		status = 401
		err = errors.New("header X-Api-Key is required")
	} else if (apiKey != "") && !config.isApiKeyEnabled(apiKey) {
		status = 401
		err = errors.New("key from header X-Api-Key is not enabled")
	}
	return status, err
}

func sendResponse(status int, resp JsonResponse, err error, w http.ResponseWriter) {
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