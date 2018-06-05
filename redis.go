package main

import (
	"log"
	"github.com/mediocregopher/radix.v2/redis"
)

var redisClient *redis.Client

func establishRedisConnection(fatal bool) {
	var err error
	redisClient, err = redis.Dial("tcp", config.Redis.Host + ":" + config.Redis.Port)
	if err != nil {
		if fatal {
			log.Fatal("Redis connection problem: " + err.Error())
		} else {
			log.Print("Redis connection problem: " + err.Error())
		}
	}
	if !testRedisConnection() {
		if fatal {
			log.Fatal("Redis test connection problem: " + err.Error())
		} else {
			log.Print("Redis test connection problem: " + err.Error())
		}
	}
}

func testRedisConnection() bool {
	err := redisClient.Cmd("SET", "TEST", 1, "EX", 1).Err
	if err != nil {
		return false
	}
	return true
}
