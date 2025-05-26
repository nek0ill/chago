#!/bin/bash

echo "Starting monitoring stack..."
docker-compose -f monitoring/docker-compose.yml up -d

echo -e "\nAccess:"
echo "Grafana:   http://localhost:3000 (admin:admin)"
echo "Prometheus: http://localhost:9090"
