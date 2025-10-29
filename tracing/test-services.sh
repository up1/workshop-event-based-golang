#!/bin/bash

echo "Testing Order Service and Report Service with distributed tracing..."

# Test creating an order
echo "Creating a test order..."
curl -X POST http://localhost:8080/order \
  -H "Content-Type: application/json" \
  -d '{
    "total_price": 1000,
    "customer_id": 1,
    "product_id": 1
  }'

echo -e "\n\nWaiting 2 seconds for message processing..."
sleep 2

# Check reports
echo "Checking generated reports..."
curl -X GET http://localhost:8081/reports

echo -e "\n\nCreating another order..."
curl -X POST http://localhost:8080/order \
  -H "Content-Type: application/json" \
  -d '{
    "total_price": 2500,
    "customer_id": 2,
    "product_id": 3
  }'

echo -e "\n\nWaiting 2 seconds for message processing..."
sleep 2

# Check reports again
echo "Checking updated reports..."
curl -X GET http://localhost:8081/reports

echo -e "\n\nTest completed!"
echo "Check Jaeger UI at: http://localhost:16686"
echo "Check RabbitMQ UI at: http://localhost:15672 (guest/guest)"