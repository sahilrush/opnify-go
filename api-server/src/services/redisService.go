package services

import (
        "context",
        "github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func RedisClient() *redis.Client{
  return redis.NewClient(&redis.Options){
    Addr:"localhost:6379",
    DB:0,
  }
}  


