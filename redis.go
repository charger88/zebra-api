package main

import (
	"log"
	"github.com/mediocregopher/radix.v2/pool"
	"time"
	"errors"
	"github.com/mediocregopher/radix.v2/redis"
)

var redisPool *pool.Pool

func establishRedisConnection(fatal bool) {
	var err error
	connectionString := config.RedisHost + ":" + config.RedisPort
	redisPool, err = pool.New("tcp", connectionString, config.RedisPool)
	if err != nil {
		if fatal {
			log.Fatal("Redis connection problem: " + err.Error())
		} else {
			log.Print("Redis connection problem: " + err.Error())
		}
	}
	if config.RedisPassword != "" {
		res := redisPool.Cmd("AUTH", config.RedisPassword)
		if  res.Err != nil {
			log.Fatal("Redis connection problem: " + res.Err.Error())
		}
	}
	_, err = testRedisConnectionAndGetClient(true)
	if err != nil {
		if fatal {
			log.Fatal("Redis test connection problem: " + err.Error())
		} else {
			log.Print("Redis test connection problem: " + err.Error())
		}
	}
	redisPool.Cmd("SELECT", config.RedisDatabase)
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
				_, err := testRedisConnectionAndGetClient(true)
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

func testRedisConnectionAndGetClient(closeClient bool) (*redis.Client, error) {
	var err error
	var redisClient *redis.Client
	if redisPool.Avail() < 1 {
		return redisClient, errors.New("redis pool was reached")
	}
	redisClient, err = redisPool.Get()
	if err != nil {
		return redisClient, err
	}
	res := redisClient.Cmd("PING")
	if closeClient {
		redisClient.Close()
	}
	if res.Err != nil {
		return redisClient, res.Err
	}
	resStr, err := res.Str()
	if err != nil {
		return redisClient, err
	}
	if resStr != "PONG" {
		return redisClient, errors.New("incorrect test response")
	}
	return redisClient, nil
}
