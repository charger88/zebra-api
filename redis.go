package main

import (
	"log"
	"github.com/mediocregopher/radix.v2/redis"
	"time"
	"errors"
)

var redisClient *redis.Client

func establishRedisConnection(fatal bool) {
	var err error
	connectionString := config.RedisHost + ":" + config.RedisPort
	redisClient, err = redis.Dial("tcp", connectionString)
	if config.RedisPassword != "" {
		res := redisClient.Cmd("AUTH", config.RedisPassword)
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
	redisClient.Cmd("SELECT", config.RedisDatabase)
	if err != nil {
		if fatal {
			log.Fatal("Redis connection problem: " + err.Error())
		} else {
			log.Print("Redis connection problem: " + err.Error())
		}
	}
}

func retestRedisConnection() {
	ticker := time.NewTicker(time.Duration(60) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <- ticker.C:
				err := testRedisConnection()
				if err != nil {
					log.Print("Redis test connection problem: " + err.Error())
				} else {
					log.Print("Redis test connection: OK")
				}
			case <- quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func testRedisConnection() error {
	res := redisClient.Cmd("PING")
	if res.Err != nil {
		return res.Err
	}
	resStr, err := res.Str()
	if err != nil {
		return err
	}
	if resStr != "PONG" {
		return errors.New("incorrect test response")
	}
	return nil
}
