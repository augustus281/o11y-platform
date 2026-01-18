#!/bin/bash

API_GATEWAY_URL="http://localhost:8080"
echo "Starting load test against $API_GATEWAY_URL"

# Port forward API gateway
kubectl port-forward -n microservices svc/api-gateway 8080:80 &
PORT_FORWARD_PID=$!
sleep 5

# Function to make requests
make_requests() {
  local endpoint=$1
  local count=$2
  
  for i in $(seq 1 $count); do
    curl -s "$API_GATEWAY_URL$endpoint" > /dev/null &
    sleep 0.1
  done
  wait
}

# Generate various types of requests
echo "Generating user requests..."
make_requests "/api/users" 20

echo "Generating order requests..."
make_requests "/api/orders" 15

echo "Generating inventory requests..."
make_requests "/api/inventory" 25

echo "Generating aggregate requests..."
for user_id in 1 2; do
  make_requests "/api/user/$user_id/orders" 10
done

echo "Creating new orders..."
for i in {1..5}; do
  curl -s -X POST "$API_GATEWAY_URL/api/orders" \
    -H "Content-Type: application/json" \
    -d '{"user_id": 1, "items": [{"product_id": 101, "quantity": 1}]}' > /dev/null &
  sleep 0.2
done
wait

echo "Load test completed!"
kill $PORT_FORWARD_PID 2>/dev/null
