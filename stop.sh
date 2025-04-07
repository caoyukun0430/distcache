#!/bin/bash

echo "Stopping all running services..."

echo "Stopping Go servers..."
pkill -f "go run main.go"

echo "Stopping etcd cluster..."
pkill -f goreman

echo "Stopping clients..."
pkill -f "go run.*grpc_client.go"


echo "All services have been stopped."
