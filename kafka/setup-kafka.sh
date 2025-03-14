#!/bin/bash

# Wait for Kafka to be ready
until kafka-topics.sh --list --bootstrap-server kafka:9092 2>/dev/null; do
  echo "Waiting for Kafka to be ready..."
  sleep 2
done

# Create topics with 3 partitions and 1 replica
for topic in "c1-datanode1-logs" "c1-datanode2-logs" "c1-datanode3-logs" "c1-datanode4-logs" "c1-datanode5-logs"; do
  kafka-topics.sh --create \
    --topic "$topic" \
    --partitions 3 \
    --replication-factor 1 \
    --bootstrap-server kafka:9092 \
    --if-not-exists
done