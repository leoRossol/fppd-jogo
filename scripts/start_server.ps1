# Start server (PowerShell)
# Usage: Open a PowerShell terminal and run this script.
# It runs the server main (build tag 'server').

param(
    [int]$Port = 12345
)

# Set GAME_PORT env var for the process and run
$env:GAME_PORT = $Port.ToString()
Write-Host "Starting server on port $Port (GO run -tags server)"
# Use `go run -tags server .` so you don't need to build a binary each time
go run -tags server .
