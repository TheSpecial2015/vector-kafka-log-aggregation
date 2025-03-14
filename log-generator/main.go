package main

import (
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
    MessagesPerSecond int
    TotalMessages     int
    OutputFile        string
    NodeName          string // ~New: Node identifier
}

// Templates—~your 30 HDFS errors
var templates = []string{
    "E1|Adding an already existing block %s",
    "E2|Verification succeeded for %s",
    "E3|Served block %s to %s",
    "E4|Got exception while serving %s to %s",
    "E5|Receiving block %s src: %s dest: %s",
    "E6|Received block %s src: %s dest: %s of size %d",
    "E7|writeBlock %s received exception %s",
    "E8|PacketResponder %s for block %s Interrupted",
    "E9|Received block %s of size %d from %s",
    "E10|PacketResponder %s Exception %s",
    "E11|PacketResponder %s for block %s terminating",
    "E12|%s:Exception writing block %s to mirror %s",
    "E13|Receiving empty packet for block %s",
    "E14|Exception in receiveBlock for block %s",
    "E15|Changing block file offset of block %s from %d to %d meta file offset to %d",
    "E16|%s:Transmitted block %s to %s",
    "E17|%s:Failed to transfer %s to %s got %s",
    "E18|Starting thread to transfer block %s to %s",
    "E19|Reopen Block %s",
    "E20|Unexpected error trying to delete block %s BlockInfo not found in volumeMap",
    "E21|Deleting block %s file %s",
    "E22|BLOCK* NameSystem allocateBlock: %s",
    "E23|BLOCK* NameSystem delete: %s is added to invalidSet of %s",
    "E24|BLOCK* Removing block %s from neededReplications as it does not belong to any file",
    "E25|BLOCK* ask %s to replicate %s to %s",
    "E26|BLOCK* NameSystem addStoredBlock: blockMap updated: %s is added to %s size %d",
    "E27|BLOCK* NameSystem addStoredBlock: Redundant addStoredBlock request received for %s on %s size %d",
    "E28|BLOCK* NameSystem addStoredBlock: addStoredBlock request received for %s on %s size %d But it does not belong to any file",
    "E29|PendingReplicationMonitor timed out block %s",
    "E30|BLOCK* ask %s to delete %s",
}

func main() {

    opt := Options{}
    flag.StringVar(&opt.OutputFile, "file", "", "The file to output (default: STDOUT)")
    flag.StringVar(&opt.PayloadGen, "payload-gen", "constant", "Payload generator: constant, fixed")
    flag.StringVar(&opt.Distribution, "distribution", "fixed", "Payload distribution")
    flag.IntVar(&opt.MessagesPerSecond, "msgpersec", 5, "Messages per second (default = 5)")
    flag.IntVar(&opt.TotalMessages, "totMessages", 100, "Total messages (fixed mode)")
    flag.StringVar(&opt.NodeName, "node", "datanode1", "Node name (e.g., datanode1)") // ~New flag
    flag.Parse()

    // Infinite or fixed?
    totalMessages := int64(opt.TotalMessages)
    if opt.PayloadGen == "constant" {
        totalMessages = 0 // ~Infinite
    }

    // Output setup
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

    // Generate logs
    generateLogs(output, totalMessages, int64(opt.MessagesPerSecond), opt.NodeName)
}

// generateLogs creates log entries
func generateLogs(output *os.File, totalMessages, msgPerSec int64, nodeName string) {
    logInterval := time.Duration(1000000/msgPerSec) * time.Microsecond
    ticker := time.NewTicker(logInterval)
    defer ticker.Stop()

    logLevels := []string{"INFO", "WARN", "ERROR"}
    count := int64(0)

    for range ticker.C {
        logEntry := createLogEntry(logLevels, nodeName)
        fmt.Fprintln(output, logEntry)

        count++
        if totalMessages > 0 && count >= totalMessages {
            break
        }
    }
}

// createLogEntry generates a log entry from templates
func createLogEntry(logLevels []string, nodeName string) string {
    timestamp := time.Now().UTC().Format(time.RFC3339Nano)
    level := logLevels[rand.Intn(len(logLevels))]
    template := templates[rand.Intn(len(templates))]
    parts := strings.SplitN(template, "|", 2)
    message := generateMessage(parts[1], nodeName)
	pid := fmt.Sprintf("%d", rand.Intn(1000))

    component := "org.apache.hadoop.hdfs.server.datanode.DataNode"
    if strings.Contains(message, "NameSystem") {
        component = "org.apache.hadoop.hdfs.server.namenode.FSNamesystem"
    }

    return timestamp + ", " + pid + " " + level + " " + component + ": " + message
}

// generateMessage fills template wildcards
func generateMessage(template, nodeName string) string {
    blockID := fmt.Sprintf("blk_%d", rand.Int63n(1000000))
    size := rand.Intn(1048576) + 1024 // ~1KB-1MB
    otherNode := fmt.Sprintf("datanode%d", rand.Intn(5)+1)
    errorMsg := randomError()

    switch {
    case strings.Contains(template, "Adding an already existing block"):
        return fmt.Sprintf("Adding an already existing block %s", blockID)
    case strings.Contains(template, "Verification succeeded"):
        return fmt.Sprintf("Verification succeeded for %s", blockID)
    case strings.Contains(template, "Served block"):
        return fmt.Sprintf("Served block %s to /10.0.0.%d", blockID, rand.Intn(256))
    case strings.Contains(template, "Got exception while serving"):
        return fmt.Sprintf("Got exception while serving %s to /10.0.0.%d", blockID, rand.Intn(256))
    case strings.Contains(template, "Receiving block"):
        return fmt.Sprintf("Receiving block %s src: /10.0.0.%d dest: %s", blockID, rand.Intn(256), nodeName)
    case strings.Contains(template, "Received block") && strings.Contains(template, "size"):
        return fmt.Sprintf("Received block %s src: /10.0.0.%d dest: %s of size %d", blockID, rand.Intn(256), nodeName, size)
    case strings.Contains(template, "writeBlock"):
        return fmt.Sprintf("writeBlock %s received exception %s", blockID, errorMsg)
    case strings.Contains(template, "PacketResponder") && strings.Contains(template, "Interrupted"):
        return fmt.Sprintf("PacketResponder %s for block %s Interrupted", nodeName, blockID)
    case strings.Contains(template, "Received block") && strings.Contains(template, "from"):
        return fmt.Sprintf("Received block %s of size %d from /10.0.0.%d", blockID, size, rand.Intn(256))
    case strings.Contains(template, "PacketResponder") && strings.Contains(template, "Exception"):
        return fmt.Sprintf("PacketResponder %s Exception %s", nodeName, errorMsg)
    case strings.Contains(template, "PacketResponder") && strings.Contains(template, "terminating"):
        return fmt.Sprintf("PacketResponder %s for block %s terminating", nodeName, blockID)
    case strings.Contains(template, "Exception writing block"):
        return fmt.Sprintf("%s:Exception writing block %s to mirror %s", nodeName, blockID, otherNode)
    case strings.Contains(template, "Receiving empty packet"):
        return fmt.Sprintf("Receiving empty packet for block %s", blockID)
    case strings.Contains(template, "Exception in receiveBlock"):
        return fmt.Sprintf("Exception in receiveBlock for block %s", blockID)
    case strings.Contains(template, "Changing block file offset"):
        return fmt.Sprintf("Changing block file offset of block %s from %d to %d meta file offset to %d", blockID, size, size+1024, size+1024)
    case strings.Contains(template, "Transmitted block"):
        return fmt.Sprintf("%s:Transmitted block %s to %s", nodeName, blockID, otherNode)
    case strings.Contains(template, "Failed to transfer"):
        return fmt.Sprintf("%s:Failed to transfer %s to %s got %s", nodeName, blockID, otherNode, errorMsg)
    case strings.Contains(template, "Starting thread to transfer"):
        return fmt.Sprintf("Starting thread to transfer block %s to %s", blockID, otherNode)
    case strings.Contains(template, "Reopen Block"):
        return fmt.Sprintf("Reopen Block %s", blockID)
    case strings.Contains(template, "Unexpected error trying to delete"):
        return fmt.Sprintf("Unexpected error trying to delete block %s BlockInfo not found in volumeMap", blockID)
    case strings.Contains(template, "Deleting block"):
        return fmt.Sprintf("Deleting block %s file /hdfs/data/%s", blockID, blockID)
    case strings.Contains(template, "allocateBlock"):
        return fmt.Sprintf("BLOCK* NameSystem allocateBlock: %s", blockID)
    case strings.Contains(template, "delete:"):
        return fmt.Sprintf("BLOCK* NameSystem delete: %s is added to invalidSet of %s", blockID, nodeName)
    case strings.Contains(template, "Removing block"):
        return fmt.Sprintf("BLOCK* Removing block %s from neededReplications as it does not belong to any file", blockID)
    case strings.Contains(template, "to replicate"):
        return fmt.Sprintf("BLOCK* ask %s to replicate %s to %s", nodeName, blockID, otherNode)
    case strings.Contains(template, "blockMap updated"):
        return fmt.Sprintf("BLOCK* NameSystem addStoredBlock: blockMap updated: %s is added to %s size %d", blockID, nodeName, size)
    case strings.Contains(template, "Redundant addStoredBlock"):
        return fmt.Sprintf("BLOCK* NameSystem addStoredBlock: Redundant addStoredBlock request received for %s on %s size %d", blockID, nodeName, size)
    case strings.Contains(template, "But it does not belong"):
        return fmt.Sprintf("BLOCK* NameSystem addStoredBlock: addStoredBlock request received for %s on %s size %d But it does not belong to any file", blockID, nodeName, size)
    case strings.Contains(template, "PendingReplicationMonitor"):
        return fmt.Sprintf("PendingReplicationMonitor timed out block %s", blockID)
    case strings.Contains(template, "to delete"):
        return fmt.Sprintf("BLOCK* ask %s to delete %s", nodeName, blockID)
    default:
        return template // ~Fallback
    }
}

// randomError generates a fake error message
func randomError() string {
    errors := []string{"java.io.IOException", "TimeoutException", "EOFException"}
    return errors[rand.Intn(len(errors))]
}