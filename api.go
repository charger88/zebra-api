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

func initRouting(pattern string, callback Endpoint) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		var status int
		var response JsonResponse
		var err error
		status, err = auth(r.Header)
		if err == nil {
			status, response, err = callback(r, Context{})
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

// TODO func initRoutingREST

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