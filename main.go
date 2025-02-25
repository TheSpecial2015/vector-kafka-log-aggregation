package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

// Options represents the configuration for log generation
type Options struct {
	PayloadGen        string
	Distribution      string
	PayloadSize       int
	MessagesPerSecond int
	TotalMessages     int
	OutputFormat      string
	OutputFile        string
	UseLogSamples     string
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     string      `json:"level"`
	Message   string      `json:"message"`
	Metadata  interface{} `json:"metadata,omitempty"`
	Source    string      `json:"source"`
}

func main() {
	opt := Options{}
	flag.StringVar(&opt.OutputFile, "file", "", "The file to output (default: STDOUT)")
	flag.StringVar(&opt.OutputFormat, "output-format", "default", "The output format: default, crio (mimic CRIO output)")
	flag.StringVar(&opt.PayloadGen, "payload-gen", "constant", "Payload generator [enum]: constant(default), fixed")
	flag.StringVar(&opt.Distribution, "distribution", "fixed", "Payload distribution [enum] (default = fixed)")
	flag.IntVar(&opt.PayloadSize, "payload_size", 100, "Payload length [int] (default = 100)")
	flag.IntVar(&opt.MessagesPerSecond, "msgpersec", 1, "Number of messages per second (default = 1)")
	flag.IntVar(&opt.TotalMessages, "totMessages", 1, "Total number of messages (only applicable for 'fixed' payload-gen")
	flag.StringVar(&opt.UseLogSamples, "use_log_samples", "false", "Use log samples or not [enum] (default = false)")
	flag.Parse()

	// Define if we should run infinitely or for a fixed number of messages
	totalMessages := int64(opt.TotalMessages)
	if opt.PayloadGen == "constant" {
		totalMessages = 0 // Run indefinitely
	}

	// Setup output destination
	var output *os.File
	var err error
	if opt.OutputFile != "" {
		output, err = os.Create(opt.OutputFile)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer output.Close()
	} else {
		output = os.Stdout
	}

	// Determine log source type
	source := "synthetic"
	if opt.UseLogSamples == "true" {
		source = "application"
	}

	// Generate logs
	generateLogs(output, totalMessages, int64(opt.MessagesPerSecond), opt.PayloadSize, opt.OutputFormat, source)
}

// generateLogs creates log entries at the specified rate
func generateLogs(output *os.File, totalMessages, msgPerSec int64, payloadSize int, format, source string) {
	logInterval := time.Duration(1000000/msgPerSec) * time.Microsecond
	ticker := time.NewTicker(logInterval)
	defer ticker.Stop()

	rand.Seed(time.Now().UnixNano())
	logLevels := []string{"info", "warning", "error", "debug"}
	count := int64(0)

	for range ticker.C {
		logEntry := createLogEntry(payloadSize, logLevels, format, source)
		fmt.Fprintln(output, logEntry)

		count++
		if totalMessages > 0 && count >= totalMessages {
			break
		}
	}
}

// createLogEntry generates a single log entry
func createLogEntry(payloadSize int, logLevels []string, format, source string) string {
	timestamp := time.Now().Format(time.RFC3339)
	level := logLevels[rand.Intn(len(logLevels))]
	message := generateRandomString(payloadSize)

	switch format {
	case "crio":
		// Mimic CRIO format
		return fmt.Sprintf("%s stdout F %s", timestamp, message)
	case "json":
		entry := LogEntry{
			Timestamp: timestamp,
			Level:     level,
			Message:   message,
			Source:    source,
			Metadata: map[string]interface{}{
				"hostname":  "log-generator",
				"pid":       os.Getpid(),
				"requestID": generateRandomString(8),
			},
		}
		jsonData, _ := json.Marshal(entry)
		return string(jsonData)
	default:
		// Simple text format
		return fmt.Sprintf("%s [%s] %s", timestamp, level, message)
	}
}

// generateRandomString creates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return strings.TrimSpace(string(b))
}