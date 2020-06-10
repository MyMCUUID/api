package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"time"
)

var client *redis.Client

func SetupRedis(){
	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"), // no password set
		DB:       0,  // use default DB
	})
}

func StoreData(ctx context.Context, username string, uuid string, response string) error {
	client.Set(ctx, fmt.Sprintf("player:%s", username), response, time.Hour * 24)
	client.Set(ctx, fmt.Sprintf("uuid:%s", uuid), response, time.Hour * 24)
	return nil
}

func HasDataFromUsername(ctx context.Context, username string) (bool, error) {
	value, err := client.Exists(ctx, fmt.Sprintf("player:%s", username)).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, nil
	}
	if value == 0 {
		return false, nil
	}
	return true, nil
}

func HasDataFromUUID(ctx context.Context, uuid string) (bool, error) {
	value, err := client.Exists(ctx, fmt.Sprintf("uuid:%s", uuid)).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, nil
	}
	if value == 0 {
		return false, nil
	}
	return true, nil
}

func GetDataFromUsername(ctx context.Context, username string) (*string, error) {
	value, err := client.Get(ctx, fmt.Sprintf("player:%s", username)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("does not exist")
	} else if err != nil {
		return nil, err
	}
	return &value, nil
}

func GetDataFromUUID(ctx context.Context, uuid string) (*string, error) {
	value, err := client.Get(ctx, fmt.Sprintf("uuid:%s", uuid)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("does not exist")
	} else if err != nil {
		return nil, err
	}
	return &value, nil
}
