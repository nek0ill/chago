<#
.SYNOPSIS
Starts the monitoring stack (Prometheus + Grafana)

.DESCRIPTION
Launches Docker containers for monitoring infrastructure with checks
for Docker availability and port conflicts.
#>

# Check Docker is installed and running
try {
    docker version | Out-Null
} catch {
    Write-Error "Docker is not available. Please install Docker Desktop for Windows."
    exit 1
}

# Check required ports
$portsInUse = Get-NetTCPConnection -State Listen | 
    Where-Object { $_.LocalPort -in @(3000,9090) } | 
    Select-Object -ExpandProperty LocalPort

if ($portsInUse) {
    Write-Error "Required ports already in use: $($portsInUse -join ',')"
    exit 1
}

try {
    Write-Host "Starting monitoring stack..."
    docker-compose -f monitoring/docker-compose.yml up -d
    
    Write-Host ""
    Write-Host "Monitoring services:"
    Write-Host "Grafana:   http://localhost:3000 (admin/admin)"
    Write-Host "Prometheus: http://localhost:9090"
} catch {
    Write-Error "Failed to start monitoring stack: $_"
    exit 1
}
