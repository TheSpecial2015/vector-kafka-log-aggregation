package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

// Kafka Configuration
const (
	KafkaTopic = "logs"
	GroupID    = "log-consumer"
)

// LogData represents the structure of our log messages
type LogData struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Level     string `json:"level"`
	Source    string `json:"source"`
}

// TimeStats tracks statistics for time differences
type TimeStats struct {
	Count    int
	Min      float64
	Max      float64
	Sum      float64
	Average  float64
}

// Initialize TimeStats with starting values
func NewTimeStats() *TimeStats {
	return &TimeStats{
		Count:    0,
		Min:      math.MaxFloat64,
		Max:      -1,
		Sum:      0,
		Average:  0,
	}
}

// Update stats with a new time difference value
func (stats *TimeStats) Update(value float64) {
	stats.Count++
	stats.Sum += value
	
	// Update min value
	if value < stats.Min {
		stats.Min = value
	}
	
	// Update max value
	if value > stats.Max {
		stats.Max = value
	}
	
	// Calculate new average
	stats.Average = stats.Sum / float64(stats.Count)
}


// Calculate Time Difference
func timeDiff(timestamp string) float64 {
	// Parse the timestamp from the log
	logTime, err := time.Parse("2006-01-02T15:04:05Z", timestamp)
	if err != nil {
		fmt.Printf("Error parsing timestamp: %v\n", err)
		return 0
	}
	
	// Get current time in UTC
	now := time.Now().UTC()
	
	// Calculate difference in seconds
	return now.Sub(logTime).Seconds()
}

// Function to print statistics
func printStats(stats *TimeStats) {
	fmt.Printf("\n===== Time Difference Statistics =====\n")
	fmt.Printf("Messages Processed: %d\n", stats.Count)
	fmt.Printf("Minimum Delay: %.2f seconds\n", stats.Min)
	fmt.Printf("Maximum Delay: %.2f seconds\n", stats.Max)
	fmt.Printf("Average Delay: %.2f seconds\n", stats.Average)
	fmt.Printf("=====================================\n\n")
}

func main() {
	// Get Kafka broker from environment or use default
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka:9092"
	}
	
	fmt.Printf("Connecting to Kafka broker: %s\n", kafkaBroker)
	
	// Setup Sarama config
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest // equivalent to 'latest'
	config.Consumer.Fetch.Min = 1                         // Fetch even 1 byte ASAP
	config.Consumer.Fetch.Default = 1000000               // 1MB
	config.Consumer.MaxWaitTime = 10 * time.Millisecond   // Wait max 10ms for data to arrive
	config.Version = sarama.V2_0_0_0                      // Kafka version
	
	// Add connection retry
	config.Net.DialTimeout = 10 * time.Second
	config.Net.ReadTimeout = 10 * time.Second
	config.Net.WriteTimeout = 10 * time.Second
	config.Metadata.Retry.Max = 5
	config.Metadata.Retry.Backoff = 5 * time.Second

	// Create consumer with stats tracking
	timeStats := NewTimeStats()
	consumer := Consumer{
		stats: timeStats,
	}

	// Setup context that will be used to cancel the consumer when we're done
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create consumer group client with retry logic
	var client sarama.ConsumerGroup
	var err error
	
	for retries := 0; retries < 10; retries++ {
		fmt.Printf("Attempting to connect to Kafka (attempt %d)...\n", retries+1)
		client, err = sarama.NewConsumerGroup([]string{kafkaBroker}, GroupID, config)
		if err == nil {
			break
		}
		fmt.Printf("Failed to connect: %v. Retrying in 5 seconds...\n", err)
		time.Sleep(5 * time.Second)
	}
	
	if err != nil {
		fmt.Printf("Failed to create consumer group client after retries: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Successfully connected to Kafka!")
	
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Printf("Error closing client: %v\n", err)
		}
	}()

	// Track errors
	go func() {
		for err := range client.Errors() {
			fmt.Printf("Kafka error: %v\n", err)
		}
	}()

	// Handle SIGINT and SIGTERM
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Create a ticker to print stats periodically
	statsTicker := time.NewTicker(30 * time.Second)
	go func() {
		for range statsTicker.C {
			if timeStats.Count > 0 {
				printStats(timeStats)
			}
		}
	}()
	defer statsTicker.Stop()

	// Consume logs in a separate goroutine
	go func() {
		for {
			// Consume from topic
			fmt.Printf("Listening for logs on topic: %s...\n\n", KafkaTopic)
			if err := client.Consume(ctx, strings.Split(KafkaTopic, ","), &consumer); err != nil {
				fmt.Printf("Error from consumer: %v\n", err)
			}
			// Check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
		}
	}()

	// Wait for termination signal
	<-signals
	fmt.Println("\nStopping log consumer...")
	cancel()
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	numLogs int
	stats   *TimeStats
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must be run in a consumer group session
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	fmt.Println("Starting to consume messages...")
	// Loop over messages in the claim
	for message := range claim.Messages() {
		logMessage := string(message.Value)
		
		var logData LogData
		if err := json.Unmarshal([]byte(logMessage), &logData); err != nil {
			fmt.Printf("Error unmarshaling JSON: %v\n", err)
			continue
		}
		
		timeDifference := timeDiff(logData.Timestamp)
		consumer.numLogs++
		
		// Update statistics with the new time difference
		if (consumer.numLogs > 150) {
			consumer.stats.Update(timeDifference)
		}
		
		fmt.Printf("✅Log %d: %.2f seconds | Level: %s | Message: %s | Min: %.2f | Max: %.2f | Avg: %.2f\n", 
			consumer.numLogs, timeDifference, logData.Level, logData.Message,
			consumer.stats.Min, consumer.stats.Max, consumer.stats.Average)
		
		// Mark the message as processed
		session.MarkMessage(message, "")
	}
	return nil
}