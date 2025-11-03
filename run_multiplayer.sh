#!/bin/bash
# Script para iniciar servidor e 2 clientes facilmente
# Uso: ./run_multiplayer.sh

echo "🎮 Iniciando ambiente multiplayer..."
echo ""
echo "Instruções:"
echo "  - Terminal 1: Servidor (este terminal)"
echo "  - Abra mais 2 terminais e execute:"
echo "    Terminal 2: RPC_ADDR='127.0.0.1:12345' CLIENTID_FILE='clientA.clientid' go run ."
echo "    Terminal 3: RPC_ADDR='127.0.0.1:12345' CLIENTID_FILE='clientB.clientid' go run ."
echo ""
echo "Iniciando servidor..."
go run -tags server .
