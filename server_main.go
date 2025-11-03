//go:build server

package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
)

// Arquivo com build tag 'server' que contém a função main para executar o servidor
func main() {
	// Inicializa e configura o servidor
	gs := NewGameServer()
	gs.parseFlags()
	rpc.Register(gs)

	// Inicia servidor RPC
	addr := fmt.Sprintf(":%d", gs.config.port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", addr, err)
	}
	defer l.Close()

	fmt.Printf("[SERVER] RPC server listening on %s (ttlProcessed=%v, ttlPlayer=%v)\n",
		addr, gs.config.ttlProcessed, gs.config.ttlPlayer)

	// Inicia limpeza automática em background
	gs.startCleanupRoutine()

	// Aceita conexões RPC até o servidor ser encerrado
	rpc.Accept(l)
}
