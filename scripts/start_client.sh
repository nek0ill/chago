#!/bin/bash

if [ -z "$1" ]; then
  echo "Usage: $0 <encryption_key> [server_host:port]"
  exit 1
fi

SERVER=${2:-"localhost:8080"}
echo "Connecting to $SERVER"
./chago client --server $SERVER --key $1
