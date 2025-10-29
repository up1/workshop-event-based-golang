#!/bin/bash

echo "Setting up RabbitMQ for Watermill..."

# Wait for RabbitMQ to be ready
echo "Waiting for RabbitMQ to be ready..."
until docker exec rabbitmq rabbitmqctl status >/dev/null 2>&1; do
    echo "RabbitMQ is unavailable - sleeping"
    sleep 2
done

echo "RabbitMQ is ready!"

# Create exchange
echo "Creating exchange 'orders'..."
docker exec rabbitmq rabbitmqadmin declare exchange name=orders type=fanout

# Create queue for service2
echo "Creating queue 'report'..."
docker exec rabbitmq rabbitmqadmin declare queue name=report durable=true

# Bind queue to exchange
echo "Binding queue 'report' to exchange 'orders'..."
docker exec rabbitmq rabbitmqadmin declare binding source=orders destination=report destination_type=queue

echo "RabbitMQ setup complete!"

# Show current setup
echo "Current exchanges:"
docker exec rabbitmq rabbitmqadmin list exchanges

echo "Current queues:"
docker exec rabbitmq rabbitmqadmin list queues

echo "Current bindings:"
docker exec rabbitmq rabbitmqadmin list bindings