package services

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient     *redis.Client
	redisSubscriber *redis.Client
	ctx             = context.Background()
	queueName       = "apiToEngine"
)

type QueueData struct {
	ID       string `json:"_id"`
	Endpoint string `json:"endpoint"`
	Req      struct {
		Body   interface{}       `json:"body"`
		Params map[string]string `json:"params"`
	} `json:"req"`
}

func init() {
	// Initialize Redis clients
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	redisSubscriber = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

// ForwardReq returns a handler function for the given endpoint
func ForwardReq(endpoint string) gin.HandlerFunc {
	return func(c *gin.Context) {
		payload := QueueData{
			ID:       uuid.NewString(),
			Endpoint: endpoint,
		}

		// Bind request body
		if err := c.ShouldBindJSON(&payload.Req.Body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request payload",
			})
			return
		}

		// Bind URL parameters
		payload.Req.Params = make(map[string]string)
		for _, param := range c.Params {
			payload.Req.Params[param.Key] = param.Value
		}

		// Convert payload to JSON
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to process request payload",
			})
			return
		}

		// Subscribe and forward to queue
		go func() {
			pubsub := redisSubscriber.Subscribe(ctx, payload.ID)
			defer pubsub.Close()

			_, err := pubsub.Receive(ctx)
			if err != nil {
				log.Printf("Subscription error: %v\n", err)
				return
			}

			ch := pubsub.Channel()
			for msg := range ch {
				var response struct {
					StatusCode int         `json:"statusCode"`
					Data       interface{} `json:"data"`
				}
				if err := json.Unmarshal([]byte(msg.Payload), &response); err != nil {
					log.Printf("Failed to unmarshal message: %v\n", err)
					return
				}

				// Send the response back to the client
				c.JSON(response.StatusCode, response.Data)
				break
			}
		}()

		// Push the payload to the queue
		if err := redisClient.LPush(ctx, queueName, payloadJSON).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to forward request",
			})
			return
		}
	}
}
