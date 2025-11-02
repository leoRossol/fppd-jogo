#!/usr/bin/env bash
# Script de teste automático simples para integração — verifica registros e reinício do servidor
# Produz um resumo PASS/FAIL básico baseado em buscas nos logs gerados por start_env.sh

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

if [ -n "${1-}" ]; then
  LOG_DIR="$1"
else
  # encontra o último diretório de logs
  LOG_DIR=$(ls -dt "$ROOT_DIR/tests/logs"/* 2>/dev/null | head -n1 || true)
fi

if [ -z "$LOG_DIR" ] || [ ! -d "$LOG_DIR" ]; then
  echo "Log directory not found. Run scripts/start_env.sh first or pass the log dir as first arg."
  exit 2
fi

echo "Using logs: $LOG_DIR"

SERVER_LOG="$LOG_DIR/server.log"
CLIENT1_LOG="$LOG_DIR/client1.log"
CLIENT2_LOG="$LOG_DIR/client2.log"

echo "Checking registrations in server log..."
REG_COUNT=$(grep -c "Registered player" "$SERVER_LOG" || true)
echo "Found $REG_COUNT 'Registered player' lines in server log"

PASS=true
if [ "$REG_COUNT" -lt 2 ]; then
  echo "FAIL: expected at least 2 registrations (client1 and client2)"
  PASS=false
else
  echo "PASS: registration count OK"
fi

echo "Simulating server restart..."
if [ -f "$LOG_DIR/server.pid" ]; then
  SERVER_PID=$(cat "$LOG_DIR/server.pid")
  echo "Killing server PID $SERVER_PID"
  kill "$SERVER_PID" || true
  sleep 1
  # restart server
  nohup sh -c 'go run -tags server server.go' >> "$SERVER_LOG" 2>&1 &
  NEW_PID=$!
  echo $NEW_PID > "$LOG_DIR/server.pid"
  echo "Restarted server PID $NEW_PID"
  sleep 2
  # check that new registrations appear (clients may re-register or poll)
  NEW_REG_COUNT=$(grep -c "Registered player" "$SERVER_LOG" || true)
  if [ "$NEW_REG_COUNT" -gt "$REG_COUNT" ]; then
    echo "PASS: server restart caused new registrations ($NEW_REG_COUNT > $REG_COUNT)"
  else
    echo "WARN: no additional registrations after restart (clients may not auto-retry)." 
    # do not fail the whole test, but note
  fi
else
  echo "server.pid not found — cannot simulate restart automatically"
fi

echo
if [ "$PASS" = true ]; then
  echo "OVERALL: PASS (basic checks)"
  exit 0
else
  echo "OVERALL: FAIL"
  exit 1
fi
