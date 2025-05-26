#!/bin/bash

# Generate encryption key if not provided
KEY=${1:-$(openssl rand -hex 32)}
PORT=${2:-8080}

echo "Starting server on port $PORT with key: $KEY"
./encrypted-chat server --port $PORT --key $KEY
