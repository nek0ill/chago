<#
.SYNOPSIS
Starts the encrypted chat server with automatic key generation

.DESCRIPTION
This script starts the chat server, generating an encryption key if none is provided.
It validates the port is available before starting.

.PARAMETER Key
Optional encryption key (32 byte hex string). If not provided, one will be generated.

.PARAMETER Port
Optional server port (default: 8080)

.EXAMPLE
.\start_server.ps1 -Key "test_key_1234567890abcdef1234567890abcdef" -Port 9090
#>

param(
    [string]$Key,
    [int]$Port = 8080
)

function Generate-EncryptionKey {
    $rng = [System.Security.Cryptography.RNGCryptoServiceProvider]::new()
    $bytes = [byte[]]::new(32)
    $rng.GetBytes($bytes)
    return [System.BitConverter]::ToString($bytes).Replace('-','').ToLower()
}

# Generate key if not provided
if (-not $Key) {
    $Key = Generate-EncryptionKey
    Write-Host "Generated encryption key: $Key"
}

# Check if port is available
try {
    $listener = [System.Net.Sockets.TcpListener]::new($Port)
    $listener.Start()
    $listener.Stop()
} catch {
    Write-Error "Port $Port is not available"
    exit 1
}

# Start server process
try {
    Write-Host "Starting server on port $Port..."
    Start-Process -NoNewWindow -FilePath ".\chago.exe" -ArgumentList "server", "--port", $Port, "--key", $Key, "--metrics-port", 2112

    # Verify metrics endpoint comes up
    $attempts = 0
    do {
        Start-Sleep -Seconds 1
        $response = try { Invoke-WebRequest "http://localhost:2112/metrics" -UseBasicParsing } catch { $null }
        $attempts++
    } while ($null -eq $response -and $attempts -lt 10)

    if ($null -eq $response) {
        Write-Warning "Metrics endpoint did not start properly"
    } else {
        Write-Host "Server started successfully"
    }
} catch {
    Write-Error "Failed to start server: $_"
    exit 1
}
