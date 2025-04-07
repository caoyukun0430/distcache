#!/bin/bash

# 设置默认值
MYSQL_HOST=${MYSQL_HOST:-"127.0.0.1"}
MYSQL_PORT=${MYSQL_PORT:-"3306"}
MYSQL_USER=${MYSQL_USER:-"root"}
MYSQL_PASSWORD=${MYSQL_PASSWORD:-"1234"}
MYSQL_DATABASE=${MYSQL_DATABASE:-"distcache"}

echo "Trying to connect to MySQL at ${MYSQL_HOST}:${MYSQL_PORT} with user ${MYSQL_USER}"
until mysql -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u "${MYSQL_USER}" -p"${MYSQL_PASSWORD}" -e "SELECT 1" >/dev/null 2>&1; do
    echo "MySQL is not ready yet... retrying in 2 seconds"
    sleep 2
done
echo "MySQL is ready!"

echo "Starting etcd cluster..."
goreman -f pkg/etcd/cluster/Procfile start &
sleep 5

echo "Starting ggcache services..."
go run main.go -port 9999 &
sleep 3

go run main.go -port 10000 -metricsPort 2223 &
sleep 3

go run main.go -port 10001 -metricsPort 2224 &
sleep 3

echo "Starting client tests..."
cd test/grpc/
go run grpc_client.go &

wait
