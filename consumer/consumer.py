from confluent_kafka import Consumer, KafkaError
import re

# Kafka Configuration
KAFKA_BROKER = "kafka:9092"
KAFKA_TOPIC = "logs"
GROUP_ID = "log-consumer"

# Define the Kafka consumer configuration
consumer_config = {
    'bootstrap.servers': KAFKA_BROKER,
    'group.id': GROUP_ID,
    'auto.offset.reset': 'earliest'  # Change to 'latest' to get only new logs
}

# Regex pattern to parse log entries
LOG_PATTERN = re.compile(r'(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z) \[(\w+)\] (\S+) (\S+) (\S+)')

def parse_log(log_message):
    """Parse a log message using regex."""
    match = LOG_PATTERN.search(log_message)
    if match:
        timestamp, level, first_string, second_string, third_string = match.groups()
        return {
            "timestamp": timestamp,
            "level": level,
            "first_string": first_string,
            "second_string": second_string,
            "third_string": third_string
        }
    return None

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
            num_logs += 1
            print(f"✅ Parsed Log {num_logs}: {log_message}")


    except KeyboardInterrupt:
        print("\nStopping log consumer...")
    finally:
        consumer.close()

if __name__ == "__main__":
    consume_logs()