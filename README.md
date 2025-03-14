# Vector-Kafka Log Aggregation

![GitHub repo size](https://img.shields.io/github/repo-size/TheSpecial2015/vector-kafka-log-aggregation)
![GitHub last commit](https://img.shields.io/github/last-commit/TheSpecial2015/vector-kafka-log-aggregation)

This project demonstrates a lightweight log aggregation pipeline using [Vector](https://vector.dev/), [Apache Kafka](https://kafka.apache.org/), and [ZooKeeper](https://zookeeper.apache.org/). Logs are generated, parsed, and sent to a Kafka topic for downstream processing or analysis. The entire setup runs in Docker containers orchestrated by Docker Compose, making it easy to spin up for development or testing.

## Features

- **Log Generation**: Custom log generator producing structured logs at a configurable rate.
- **Log Parsing**: Vector parses logs into timestamp, level, and message fields using Grok patterns.
- **Kafka Integration**: Logs are aggregated into a Kafka topic with Snappy compression.
- **ZooKeeper**: Manages Kafka’s coordination in a traditional setup.
- **Dockerized**: Fully containerized with `wurstmeister/kafka` and `wurstmeister/zookeeper`.

## Architecture

### Workflow Diagram

<img src="https://github.com/TheSpecial2015/vector-kafka-log-aggregation/blob/ansible/diagram-v2.png" alt="architecture diagram" width="1074" height="709" />

### Dashboard

<img src="https://github.com/TheSpecial2015/vector-kafka-log-aggregation/blob/ansible/logs-dashboard.png" alt="architecture diagram" width="635" height="335" />

1. The log generator writes logs to `/logs/output.log`.
2. Vector reads, parses, and forwards them as JSON to Kafka’s `logs` topic.
3. Kafka stores the messages, coordinated by ZooKeeper.
4. Consumers (not included) can read from the `logs` topic.

### The Pipeline Latency :

<img src="https://github.com/TheSpecial2015/vector-kafka-log-aggregation/blob/main/latency.png" alt="latency metric" width="300" height="150" />

### The Pipeline Ressource Consumption :

<img src="https://github.com/TheSpecial2015/vector-kafka-log-aggregation/blob/ansible/resource-stats.png" alt="ressource metric" width="750" height="200" />

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- [Git](https://git-scm.com/downloads)
- Basic familiarity with Kafka and Vector

## Getting Started

### Clone the Repository

```bash
git clone https://github.com/TheSpecial2015/vector-kafka-log-aggregation.git
cd vector-kafka-log-aggregation
```
