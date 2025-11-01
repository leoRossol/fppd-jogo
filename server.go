//go:build server
// +build server

package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

// server.go
// ----------------
// Este arquivo contém um esqueleto funcional do servidor RPC. Abaixo há comentários
// e TODOs detalhados (destinados ao Member A) com instruções passo-a-passo sobre o que
// implantar para satisfazer todos os requisitos do enunciado.
//
// Member A (Servidor) - Responsabilidades principais (resumidas):
// 1) Implementar TTL/limpeza para `processed` e `players` (goroutine de manutenção).
// 2) Opcional: persistência simples (checkpoint em ficheiro) para `processed` se for exigido
//    que exactly-once sobreviva a reinícios do servidor.
// 3) Ampliar suporte a comandos (ex.: LOGOUT, CHANGE_LIVES) no switch de `SendCommand`.
// 4) Melhorar logs (timestamps, nível de log) e permitir configurar porta via env/flag.
// 5) Escrever testes que verifiquem retransmissão/duplicate-detection (ver `tests/`).
//
// Pontos de integração com o restante do projeto (onde o cliente fará chamadas):
// - `REGISTER` (cliente chama no startup)
// - `UPDATE_POS` (cliente chama após mover localmente)
// - `GetState` (clientes chamam periodicamente para polling)

// GameServer gerencia o estado global observado pelos clientes
type GameServer struct {
	mu        sync.Mutex
	players   map[string]PlayerInfo
	processed map[string]map[int64]CommandReply // processed[clientID][seq] = reply
	// TODO Member A: adicionar campos para TTL e persistência
	//   processedTimestamps map[string]map[int64]int64 // timestamp of when processed
}

func NewGameServer() *GameServer {
	return &GameServer{
		players:   make(map[string]PlayerInfo),
		processed: make(map[string]map[int64]CommandReply),
	}
}

// SendCommand aplica um comando recebido de um cliente garantindo exactly-once
// TODO Member A: ampliar handling de payloads tipados (ver rpc_types.go) e adicionar
// mecanismos de persistência se for requerido pela especificação do professor.
func (s *GameServer) SendCommand(args *CommandArgs, reply *CommandReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// LOG: sempre imprimir requisição recebida (para depuração conforme enunciado)
	fmt.Printf("[SERVER] %s Received SendCommand from %s seq=%d cmd=%s payload=%v\n", time.Now().Format(time.RFC3339), args.ClientID, args.Seq, args.Cmd, args.Payload)

	if _, ok := s.processed[args.ClientID]; !ok {
		s.processed[args.ClientID] = make(map[int64]CommandReply)
	}

	// Re-transmissão detectada: devolver o mesmo resultado sem reexecutar
	if prev, ok := s.processed[args.ClientID][args.Seq]; ok {
		*reply = prev
		fmt.Printf("[SERVER] %s Duplicate command detected for %s seq=%d - returning cached reply\n", time.Now().Format(time.RFC3339), args.ClientID, args.Seq)
		return nil
	}

	// Implementação simples dos comandos esperados. Exemplo: UPDATE_POS
	var cr CommandReply
	cr.Seq = args.Seq

	switch args.Cmd {
	case "UPDATE_POS":
		// Atualmente Payload é map[string]interface{}; para maior segurança, considere
		// trocar por um struct UpdatePosPayload em rpc_types.go e usar esse tipo aqui.
		x, _ := args.Payload["x"].(int)
		y, _ := args.Payload["y"].(int)
		lives, _ := args.Payload["lives"].(int)
		pi := PlayerInfo{ID: args.ClientID, X: x, Y: y, Lives: lives, LastSeen: time.Now().Unix()}
		s.players[args.ClientID] = pi
		cr.Applied = true
		cr.Message = "position-updated"
		fmt.Printf("[SERVER] %s Updated position for %s -> (%d,%d) lives=%d\n", time.Now().Format(time.RFC3339), args.ClientID, x, y, lives)
	case "REGISTER":
		// payload pode conter campos extras; registra jogador básico
		pi := PlayerInfo{ID: args.ClientID, X: 0, Y: 0, Lives: 3, LastSeen: time.Now().Unix()}
		s.players[args.ClientID] = pi
		cr.Applied = true
		cr.Message = "registered"
		fmt.Printf("[SERVER] %s Registered player %s\n", time.Now().Format(time.RFC3339), args.ClientID)
	case "LOGOUT":
		// TODO Member A: implementar remoção segura do jogador
		delete(s.players, args.ClientID)
		cr.Applied = true
		cr.Message = "logged-out"
		fmt.Printf("[SERVER] %s Player %s logged out\n", time.Now().Format(time.RFC3339), args.ClientID)
	default:
		cr.Applied = false
		cr.Message = "unknown-command"
		fmt.Printf("[SERVER] %s Unknown command %s from %s\n", time.Now().Format(time.RFC3339), args.Cmd, args.ClientID)
	}

	// Armazena o resultado para garantir exactly-once
	s.processed[args.ClientID][args.Seq] = cr
	*reply = cr
	return nil
}

// GetState retorna o estado atual observado pelo servidor (lista de jogadores)
// TODO Member A: filtrar jogadores inativos e/ou adicionar paginação se for o caso
func (s *GameServer) GetState(args *ClientIDArgs, reply *StateReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Printf("[SERVER] %s Received GetState from %s at %s\n", time.Now().Format(time.RFC3339), args.ClientID, args.Now)

	players := make([]PlayerInfo, 0, len(s.players))
	for _, p := range s.players {
		players = append(players, p)
	}
	reply.Players = players
	reply.ServerTime = time.Now().Unix()

	fmt.Printf("[SERVER] %s Replying GetState to %s with %d players\n", time.Now().Format(time.RFC3339), args.ClientID, len(players))
	return nil
}

func main() {
	gs := NewGameServer()
	rpc.Register(gs)

	addr := ":12345" // TODO: Member A - tornar configurável via flag/env
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", addr, err)
	}
	defer l.Close()

	fmt.Printf("[SERVER] RPC server listening on %s\n", addr)

	// TODO Member A: iniciar uma goroutine que periodicamente limpe `processed` e `players` inativos.
	// Exemplo:
	// go func() {
	//   for range time.Tick(1 * time.Minute) {
	//     s.mu.Lock()
	//     // remover processed antigos / players inativos
	//     s.mu.Unlock()
	//   }
	// }()

	// rpc.Accept will block and accept connections
	rpc.Accept(l)
}
