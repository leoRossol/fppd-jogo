# Start one or more clients (PowerShell)
# Usage: start_clients.ps1 -Count 2

param(
    [int]$Count = 1
)

for ($i = 1; $i -le $Count; $i++) {
    Write-Host "Starting client $i"
    Start-Process -FilePath pwsh -ArgumentList '-NoExit','-Command','go run .'
}

Write-Host "Started $Count client(s) in new terminal(s)."