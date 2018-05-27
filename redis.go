package main

import (
	"log"
	"github.com/mediocregopher/radix.v2/redis"
)

var redisClient *redis.Client

func establishRedisConnection() {
	var err error
	redisClient, err = redis.Dial("tcp", config.Redis.Host + ":" + config.Redis.Port)
	if err != nil {
		log.Fatal("Redis connection problem: " + err.Error())
	}
	testRedisConnection()
}

// TODO Replace log.Fatal to something more smart
func testRedisConnection() bool {
	value := 1
	key := "PING:" + randomString(32, randomStringLcD)
	err := redisClient.Cmd("SET", key, value, "EX", 60).Err
	if err != nil {
		log.Fatal("Redis connection problem: " + err.Error())
		return false
	}
	resp := redisClient.Cmd("GET", key)
	if resp.Err != nil {
		log.Fatal("Redis connection problem: " + err.Error())
		return false
	}
	lvalue, err := resp.Int()
	if err != nil {
		log.Fatal("Redis test problem: " + err.Error())
		return false
	}
	if lvalue != value {
		log.Fatal("Redis test problem: wrong test value")
		return false
	}
	return true
}
