package main

import (
	"net/http"
	"encoding/json"
	"fmt"
	"log"
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
		status, response, err := callback(r, Context{})
		sendResponse(status, response, err, w)
	})
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