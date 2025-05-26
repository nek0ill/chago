# Encrypted Chat Application

Secure end-to-end encrypted chat application with built-in monitoring using Prometheus and Grafana.

## Windows Quick Start

1. Install prerequisites:
   ```powershell
   # Install Go
   winget install GoLang.Go
   
   # Install Docker Desktop
   winget install Docker.DockerDesktop
   ```

2. Build and run:
   ```powershell
   # 1. Build application
   go build -o encrypted-chat.exe
   
   # 2. Start server (PowerShell)
   .\scripts\start_server.ps1
   
   # 3. Launch monitoring
   .\scripts\start_monitoring.ps1
   
   # 4. Connect client
   .\scripts\start_client.ps1 -Key "YOUR_ENCRYPTION_KEY"
   ```

## Cross-Platform Support

For Linux/macOS, use the .sh scripts in the scripts/ directory.

## Windows Firewall Setup

If seeing connection issues:
```powershell
# Allow inbound connections
New-NetFirewallRule -DisplayName "Encrypted Chat" -Direction Inbound -LocalPort 8080,2112 -Protocol TCP -Action Allow

# Verify rules
Get-NetFirewallRule | Where-Object { $_.DisplayName -like "*Encrypted Chat*" }
```

# 2. Start server (generate key if needed)
./encrypted-chat server --port 8080 --key $(openssl rand -hex 32)

# 3. Start monitoring stack
docker-compose -f monitoring/docker-compose.yml up -d

# 4. Connect client
./encrypted-chat client --server localhost:8080 --key YOUR_KEY
```

## Features

- End-to-end AES-256 encryption
- Secure key exchange
- Real-time messaging
- Built-in monitoring
- Horizontal scalability

## Monitoring

### Metrics Collected:
- `chat_messages_sent_total`: Total messages sent
- `chat_messages_received_total`: Total messages received  
- `chat_active_connections`: Current active connections

### Access Dashboards:
- Grafana: http://localhost:3000 (admin/admin)
- Prometheus: http://localhost:9090

## CI/CD Pipeline

1. Build stage:
   - Compiles Go application
   - Runs unit tests
   - Tests metrics endpoint

2. Build Monitoring:
   - Validates monitoring config
   - Builds Docker images

3. Deployment:
   - Pushes images to registry  
   - Deploys to Kubernetes

## Security Best Practices

1. Key Management:
   - Generate strong keys (`openssl rand -hex 32`)
   - Rotate keys frequently
   - Never hardcode keys

2. Monitoring:
   - Protect Grafana/Prometheus with auth
   - Use network policies to restrict access
   - Monitor for failed decryption attempts

3. Infrastructure:
   - Use TLS for all traffic
   - Regularly update dependencies
   - Implement rate limiting

## Testing

```bash
# Verify metrics endpoint
curl http://localhost:2112/metrics

# Load testing  
./scripts/load_test.sh

# Encryption test
./scripts/test_encryption.sh
```

## Helper Scripts

- `start_server.sh`: Starts server with generated key
- `start_client.sh`: Connects to server
- `start_monitoring.sh`: Launches monitoring stack
