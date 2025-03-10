package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/gorilla/websocket"
)

// Kafka Configuration
const (
	KafkaTopic = "logs"
	GroupID    = "log-consumer"
)

// LogData represents the structure of our log messages
type LogEntry struct {
    Host            string `json:"host"`
    IPAddr          string `json:"ipaddr"`
    Date            string `json:"date"`
    Time            string `json:"time"`
    PID             string `json:"pid"`
    Thread          string `json:"thread"`
    Level           string `json:"level"`
    Message         string `json:"message"`
}

type Consumer struct {
	messages chan LogEntry
}


// WebSocket upgrader
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // Adjust this for production
    },
}

// Store connected clients
type wsServer struct {
    clients    map[*websocket.Conn]bool
    broadcast  chan LogEntry
    register   chan *websocket.Conn
    unregister chan *websocket.Conn
    mutex      sync.Mutex
}

func newWSServer() *wsServer {
    return &wsServer{
        clients:    make(map[*websocket.Conn]bool),
        broadcast:  make(chan LogEntry),
        register:   make(chan *websocket.Conn),
        unregister: make(chan *websocket.Conn),
    }
}

// Run the WebSocket server
func (s *wsServer) run() {
    for {
        select {
        case client := <-s.register:
            s.mutex.Lock()
            s.clients[client] = true
            s.mutex.Unlock()
        case client := <-s.unregister:
            s.mutex.Lock()
            if _, ok := s.clients[client]; ok {
                delete(s.clients, client)
                client.Close()
            }
            s.mutex.Unlock()
        case message := <-s.broadcast:
            s.mutex.Lock()
            for client := range s.clients {
                err := client.WriteJSON(message)
                if err != nil {
                    log.Printf("WebSocket error: %v", err)
                    client.Close()
                    delete(s.clients, client)
                }
            }
            s.mutex.Unlock()
        }
    }
}

// WebSocket handler
func (s *wsServer) handleWS(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Upgrade error: %v", err)
        return
    }
    defer conn.Close()

    s.register <- conn

    // Keep connection alive
    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            s.unregister <- conn
            break
        }
    }
}

func main() {

	// ** Get Kafka broker from environment or use default
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka:9092"
	}
	fmt.Printf("Connecting to Kafka broker: %s\n", kafkaBroker)
	
	// ** Setup Sarama config
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest // equivalent to 'latest'
	config.Consumer.Fetch.Min = 1                         // Fetch even 1 byte ASAP
	config.Consumer.Fetch.Default = 1000000               // 1MB
	config.Consumer.MaxWaitTime = 10 * time.Millisecond   // Wait max 10ms for data to arrive
	config.Version = sarama.V2_0_0_0                      // Kafka version
	
	// ** Add connection retry
	config.Net.DialTimeout = 10 * time.Second
	config.Net.ReadTimeout = 10 * time.Second
	config.Net.WriteTimeout = 10 * time.Second
	config.Metadata.Retry.Max = 5
	config.Metadata.Retry.Backoff = 5 * time.Second

	// ** Create consumer with stats tracking
	consumer := Consumer{
		messages: make(chan LogEntry),
	}

	// ** Setup context that will be used to cancel the consumer when we're done
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ** Create consumer group client with retry logic
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

	// ** Track errors
	go func() {
		for err := range client.Errors() {
			fmt.Printf("Kafka error: %v\n", err)
		}
	}()

	// ** Handle SIGINT and SIGTERM
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// ** Consume logs in a separate goroutine
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

    // Set up WebSocket server
    ws := newWSServer()
    go ws.run()
    
    // Connect WebSocket broadcast to Kafka messages
    go func() {
        for msg := range consumer.messages {
            ws.broadcast <- msg
        }
    }()

    // Set up HTTP handler
    http.HandleFunc("/ws", ws.handleWS)
    
    // Start HTTP server
    log.Println("Starting WebSocket server on :8080")
    err = http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }

	// ** Wait for termination signal
	<-signals
	fmt.Println("\nStopping log consumer...")
	cancel()

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

    for message := range claim.Messages() {
        logMessage := string(message.Value)
        
        var logEntry LogEntry
        if err := json.Unmarshal([]byte(logMessage), &logEntry); err != nil {
            fmt.Printf("Error unmarshaling JSON: %v\n", err)
            continue
        }
        
        // Send to WebSocket clients
        consumer.messages <- logEntry
        
        // Mark the message as processed
        session.MarkMessage(message, "")
    }
    return nil
}