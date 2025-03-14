package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
)

// Kafka Configuration
const (
    GroupID = "log-consumer"
)

// Redis Configuration
const (
    RedisKey = "log_entries" // Redis list key to store messages
)

// LogData represents the structure of our log messages
type LogEntry struct {
    Host      string `json:"host"`
    IPAddr    string `json:"ipaddr"`
    Timestamp string `json:"timestamp"`
    PID       string `json:"pid"`
    Thread    string `json:"thread"`
    Level     string `json:"level"`
    Message   string `json:"message"`
}

type Consumer struct {
    messages chan LogEntry
    redis    *redis.Client
}

func main() {
    // Kafka setup
    kafkaBroker := os.Getenv("KAFKA_BROKER")
    if kafkaBroker == "" {
        kafkaBroker = "kafka:9092"
    }
    fmt.Printf("Connecting to Kafka broker: %s\n", kafkaBroker)

    // Redis setup
    redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        redisAddr = "localhost:6379"
    }
    redisPassword := os.Getenv("REDIS_PASSWORD") // Get password from environment variable
    fmt.Printf("Connecting to Redis at: %s\n", redisAddr)

    redisClient := redis.NewClient(&redis.Options{
        Addr:     redisAddr,
        Password: redisPassword, // Use the password here
        DB:       0,            // use default DB
    })

    // Test Redis connection
    if err := redisClient.Ping(context.Background()).Err(); err != nil {
        fmt.Printf("Failed to connect to Redis: %v\n", err)
        os.Exit(1)
    }

    topics := []string{
        "c1-datanode1-logs",
        "c1-datanode2-logs",
        "c1-datanode3-logs",
        "c1-datanode4-logs",
        "c1-datanode5-logs",
    }

    // Setup Sarama config
    config := sarama.NewConfig()
    config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
    config.Consumer.Offsets.Initial = sarama.OffsetNewest
    config.Consumer.Fetch.Min = 1
    config.Consumer.Fetch.Default = 1000000
    config.Consumer.MaxWaitTime = 10 * time.Millisecond
    config.Version = sarama.V2_0_0_0
    config.Net.DialTimeout = 10 * time.Second
    config.Net.ReadTimeout = 10 * time.Second
    config.Net.WriteTimeout = 10 * time.Second
    config.Metadata.Retry.Max = 5
    config.Metadata.Retry.Backoff = 5 * time.Second

    // Create channels and consumer
    mergedMessages := make(chan LogEntry, 1000)
    consumer := Consumer{
        messages: mergedMessages,
        redis:    redisClient,
    }

    // Setup context
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Create consumer group
    client, err := sarama.NewConsumerGroup([]string{kafkaBroker}, GroupID, config)
    if err != nil {
        fmt.Printf("Failed to create consumer group client: %v\n", err)
        os.Exit(1)
    }
    defer client.Close()

    // Handle signals
    signals := make(chan os.Signal, 1)
    signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

    // WaitGroup to keep track of all goroutines
    var wg sync.WaitGroup

    // Start consuming from all topics
    wg.Add(1)
    go func() {
        defer wg.Done()
        for {
            if err := client.Consume(ctx, topics, &consumer); err != nil {
                fmt.Printf("Error from consumer: %v\n", err)
            }
            if ctx.Err() != nil {
                return
            }
        }
    }()

    // Push messages to Redis
    wg.Add(1)
    go func() {
        defer wg.Done()
        for msg := range mergedMessages {
            // Convert LogEntry to JSON
            jsonMsg, err := json.Marshal(msg)
            if err != nil {
                fmt.Printf("Error marshaling message to JSON: %v\n", err)
                continue
            }

            // Push to Redis list
            if err := redisClient.LPush(ctx, RedisKey, jsonMsg).Err(); err != nil {
                fmt.Printf("Error pushing to Redis: %v\n", err)
                continue
            }
            
            // Optional: Print for debugging
            fmt.Printf("Pushed to Redis - Host: %s, Level: %s\n", msg.Host, msg.Level)
        }
    }()

    fmt.Println("Successfully connected to Kafka and Redis!")
    fmt.Printf("Listening for logs on topics: %s...\n", strings.Join(topics, ", "))

    // Wait for termination signal
    <-signals
    fmt.Println("\nStopping consumer...")
    cancel()
    close(mergedMessages)
    redisClient.Close()
    wg.Wait()
}

// Setup is run at the beginning of a new session
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
    return nil
}

// Cleanup is run at the end of a session
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
    return nil
}

// ConsumeClaim processes messages from a partition
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
    fmt.Printf("Starting to consume messages from topic %s partition %d\n",
        claim.Topic(), claim.Partition())

    for message := range claim.Messages() {
        var logEntry LogEntry
        if err := json.Unmarshal(message.Value, &logEntry); err != nil {
            fmt.Printf("Error unmarshaling JSON: %v\n", err)
            continue
        }

        // Send to merged channel
        consumer.messages <- logEntry

        // Mark message as processed
        session.MarkMessage(message, "")
    }
    return nil
}