#!/usr/bin/env bash
# Script para iniciar ambiente de integração simples (macOS / Linux)
# Cria pasta de logs timestamped e inicia servidor + 2 clientes em background.
# Observação: este script assume que `go` está no PATH e que executar `go run` no repo inicia
# o servidor e os clientes corretamente. Pode ser necessário ajustar caminhos/flags conforme a implementação.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_DIR="$ROOT_DIR/tests/logs/$TIMESTAMP"
mkdir -p "$LOG_DIR"

echo "Logs: $LOG_DIR"

cd "$ROOT_DIR"

echo "Starting server..."
# Inicie o servidor compilando o pacote inteiro (inclui arquivos semânticos e arquivos com tags)
# usar 'go run -tags server .' em vez de 'go run -tags server server.go'
nohup sh -c 'go run -tags server .' > "$LOG_DIR/server.log" 2>&1 &
SERVER_PID=$!
echo $SERVER_PID > "$LOG_DIR/server.pid"
echo "Server PID: $SERVER_PID"

sleep 1

echo "Starting client 1..."
# Start client 1 in background with its own clientid file
# run the whole package so all files (client_rpc.go, jogo.go, etc.) are compiled
nohup env CLIENTID_FILE="$LOG_DIR/client1.clientid" RPC_ADDR="127.0.0.1:12345" sh -c 'go run .' > "$LOG_DIR/client1.log" 2>&1 &
CLIENT1_PID=$!
echo $CLIENT1_PID > "$LOG_DIR/client1.pid"
echo "Client1 PID: $CLIENT1_PID"

sleep 0.5

echo "Starting client 2..."
nohup env CLIENTID_FILE="$LOG_DIR/client2.clientid" RPC_ADDR="127.0.0.1:12345" sh -c 'go run .' > "$LOG_DIR/client2.log" 2>&1 &
CLIENT2_PID=$!
echo $CLIENT2_PID > "$LOG_DIR/client2.pid"
echo "Client2 PID: $CLIENT2_PID"

echo
echo "Environment started. PIDs:" 
echo "  server:  $SERVER_PID (logs: $LOG_DIR/server.log)"
echo "  client1: $CLIENT1_PID (logs: $LOG_DIR/client1.log)"
echo "  client2: $CLIENT2_PID (logs: $LOG_DIR/client2.log)"
echo
echo "To stop all processes:"
echo "  kill $SERVER_PID $CLIENT1_PID $CLIENT2_PID"
echo
echo "Log directory: $LOG_DIR"
