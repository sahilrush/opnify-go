package services

import (
	
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

var (
	client     *redis.Client
	subscriber *redis.Client
)

func ConnectRedis() error {
	client = redis.NewClient(&redis.Options{
		Addr:     " localhost:6379",
		Password: "",
		DB:       0,
	})

	subscriber = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if err := client.Ping(ctx).Err(); err != nil {

		return fmt.Errorf("failed to connect to redis client: %W", err)

	}

	if err := subscriber.Ping(ctx).Err(); err != nil {

		return fmt.Errorf("failed to connect to Redis subscriber: %w", err)
	}
	log.Println("Connected to Redis")
	return nil
}

// QueuePush pushes data to a Redis queue (list)

func QueuePush(queueName string, data string) error {
	if client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	err := client.LPush(ctx, queueName, data).Err()
	if err != nil {
		return fmt.Errorf("failed to push to queue: &w", err)
	}

	log.Printf("Data pushed to queue %s: %s\n", queueName, data)
	return nil
}

// Subscribe listens for messages on a Redis channel

func Subscribe(channelName string) {
	if subscriber == nil {
		log.Fatal("Redis subscriber is not initialized")
	}
	pubsub := subscriber.Subscribe(ctx, channelName)

	_, err := pubsub.Receive(ctx)
	if err != nil {
		log.Fatalf("failed to subscribe to channel %s: %v", channelName, err)
	}
	log.Printf("Subscribed to channel: %s\n", channelName)

	//listen for messages
	ch := pubsub.Channel()
	for msg := range ch {
		log.Printf("Received message from channel %s: %s\n", msg.Channel, msg.Payload)
	}

}

//close redis closes the redis client

func CloseRedis() {
	if client != nil {
		client.Close()
	}
	if subscriber != nil {
		subscriber.Close()
	}
	log.Println("Redis connections closed")

}
