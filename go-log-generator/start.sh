#!/bin/bash

sleep 20

# Start Vector in the background
/usr/local/bin/vector --config /etc/vector/vector.toml &
VECTOR_PID=$!

# Give Vector a moment to start up
sleep 2

# Start generating logs
/app/log-generator --file /logs/output.log --output-format default --payload-gen constant --payload_size 200 --msgpersec 5 &
LOG_GEN_PID=$!

# Keep the container running and handle termination
trap "kill $VECTOR_PID $LOG_GEN_PID; exit" SIGTERM SIGINT
wait