from confluent_kafka import Consumer, KafkaError
from datetime import datetime
import json
import re

# Kafka Configuration
KAFKA_BROKER = "kafka:9092"
KAFKA_TOPIC = "logs"
GROUP_ID = "log-consumer"

# Define the Kafka consumer configuration
consumer_config = {
    'bootstrap.servers': KAFKA_BROKER,
    'group.id': GROUP_ID,
    'auto.offset.reset': 'latest',  # Change to 'latest' to get only new logs
    'fetch.min.bytes': 1,         # Fetch even 1 byte ASAP
    'fetch.wait.max.ms': 10       # Wait max 10ms for data to arrive
}

# Calculate Time Difference
def time_diff(timestamp):
    """
    Calculate the time difference between the log timestamp and the current time in seconds.
    """
    log_time = datetime.strptime(timestamp, "%Y-%m-%dT%H:%M:%SZ")
    now = datetime.utcnow()

    time_diff = now - log_time
    return time_diff.total_seconds()

def consume_logs():
    """Consume logs from Kafka topic in real-time."""
    consumer = Consumer(consumer_config)
    consumer.subscribe([KAFKA_TOPIC])
    num_logs = 0

    try:
        print(f"Listening for logs on topic: {KAFKA_TOPIC}...\n")
        while True:
            msg = consumer.poll(timeout=1.0)
            
            if msg is None:
                continue
            if msg.error():
                if msg.error().code() == KafkaError._PARTITION_EOF:
                    continue
                else:
                    print(f"Kafka error: {msg.error()}")
                    continue
            
            log_message = msg.value().decode('utf-8').strip()
            log_data = json.loads(log_message)

            time_difference = time_diff(log_data["timestamp"])
            num_logs += 1
            print(f"✅Log {num_logs}: {time_difference} seconds")


    except KeyboardInterrupt:
        print("\nStopping log consumer...")
    finally:
        consumer.close()

if __name__ == "__main__":
    consume_logs()