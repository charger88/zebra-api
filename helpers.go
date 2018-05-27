package main

import (
	"time"
	"math/rand"
)

var randomInitialized = false

const randomStringUcLcD = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const randomStringLcD = "abcdefghijklmnopqrstuvwxyz0123456789"
const randomStringLc = "abcdefghijklmnopqrstuvwxyz"
const randomStringD = "0123456789"

func randomString(n int, chars string) string {
	if !randomInitialized {
		rand.Seed(time.Now().Unix())
		randomInitialized = true
	}
	letterRunes := []rune(chars)
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
