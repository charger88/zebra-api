package main

import (
	"log"
	"github.com/mediocregopher/radix.v2/redis"
)

var redisClient *redis.Client

func establishRedisConnection(fatal bool) {
	var err error
	connectionString := config.Redis.Host + ":" + config.Redis.Port
	redisClient, err = redis.Dial("tcp", connectionString)
	if config.Redis.Password != "" {
		res := redisClient.Cmd("AUTH", config.Redis.Password)
		if  res.Err != nil {
			redisClient.Close()
			log.Fatal("Redis connection problem: " + res.Err.Error())
		}
	}
	if err != nil {
		if fatal {
			log.Fatal("Redis connection problem: " + err.Error())
		} else {
			log.Print("Redis connection problem: " + err.Error())
		}
	}
	err = testRedisConnection()
	if err != nil {
		if fatal {
			log.Fatal("Redis test connection problem: " + err.Error())
		} else {
			log.Print("Redis test connection problem: " + err.Error())
		}
	}
}

func testRedisConnection() error {
	err := redisClient.Cmd("SET", "TEST", 1, "EX", 1).Err
	if err != nil {
		return err
	}
	return nil
}
