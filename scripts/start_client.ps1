<#
.SYNOPSIS
Connects to an encrypted chat server

.DESCRIPTION
Establishes a connection to a running chat server instance with input validation
and connection error handling.

.PARAMETER Key
Required encryption key (must match server key)

.PARAMETER Server
Optional server address (default: localhost:8080)

.EXAMPLE
.\start_client.ps1 -Key "test_key_1234567890abcdef1234567890abcdef" -Server "192.168.1.100:9090"
#>

param(
    [Parameter(Mandatory=$true)]
    [string]$Key,

    [string]$Server = "localhost:8080"
)

# Validate server address format
if ($Server -notmatch "^.+:\d+$") {
    Write-Error "Server address must be in format host:port"
    exit 1
}

try {
    Write-Host "Connecting to server at $Server..."
    Start-Process -NoNewWindow -FilePath ".\chago.exe" -ArgumentList "client", "--server", $Server, "--key", $Key
} catch {
    Write-Error "Failed to start client: $_"
    exit 1
}
