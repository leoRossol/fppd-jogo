//go:build server

//build server

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

// GameServer gerencia o estado global do jogo multiplayer, incluindo:
// - Lista de jogadores ativos e suas posições
// - Sistema de deduplicação de comandos (exactly-once)
// - Limpeza automática de dados antigos
type GameServer struct {
	mu        sync.Mutex                        // Protege acesso concorrente aos maps
	players   map[string]PlayerInfo             // Mapa de jogadores ativos indexado por ClientID
	processed map[string]map[int64]CommandReply // Cache de comandos processados para deduplicação

	// Controle de TTL (Time To Live)
	processedTimestamps map[string]map[int64]time.Time // Registra quando cada comando foi processado
	config              struct {
		port         int           // Porta do servidor RPC
		ttlProcessed time.Duration // Tempo máximo para manter comandos em cache
		ttlPlayer    time.Duration // Tempo máximo sem atualização antes de remover jogador
	}
}

func NewGameServer() *GameServer {
	s := &GameServer{
		players:             make(map[string]PlayerInfo),
		processed:           make(map[string]map[int64]CommandReply),
		processedTimestamps: make(map[string]map[int64]time.Time),
	}

	// Configurações default
	s.config.port = 12345
	s.config.ttlProcessed = 30 * time.Minute
	s.config.ttlPlayer = 1 * time.Minute

	return s
}

// SendCommand processa comandos dos clientes com garantia de exactly-once:
// - REGISTER: registra novo jogador
// - UPDATE_POS: atualiza posição do jogador
// - LOGOUT: remove jogador do servidor
func (s *GameServer) SendCommand(args *CommandArgs, reply *CommandReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Inicializa estruturas de deduplicação para novo cliente
	if _, ok := s.processed[args.ClientID]; !ok {
		s.processed[args.ClientID] = make(map[int64]CommandReply)
		s.processedTimestamps[args.ClientID] = make(map[int64]time.Time)
	}

	// Sistema de deduplicação: retorna resposta em cache se comando já foi processado
	if prev, ok := s.processed[args.ClientID][args.Seq]; ok {
		*reply = prev
		fmt.Printf("[SERVER] %s Duplicate command detected for %s seq=%d - returning cached reply\n",
			time.Now().Format(time.RFC3339), args.ClientID, args.Seq)
		return nil
	}

	// Implementação simples dos comandos esperados.
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

	// Armazena o resultado e timestamp
	s.processed[args.ClientID][args.Seq] = cr
	s.processedTimestamps[args.ClientID][args.Seq] = time.Now()
	*reply = cr
	return nil
}

// GetState retorna lista de jogadores ativos para os clientes
// Usado pelo cliente para sincronizar estado do jogo
func (s *GameServer) GetState(args *ClientIDArgs, reply *StateReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Printf("[SERVER] %s Received GetState from %s at %s\n", time.Now().Format(time.RFC3339), args.ClientID, args.Now)

	// Constrói lista de jogadores ativos
	players := make([]PlayerInfo, 0, len(s.players))
	for _, p := range s.players {
		players = append(players, p)
	}
	reply.Players = players
	reply.ServerTime = time.Now().Unix()

	fmt.Printf("[SERVER] %s Replying GetState to %s with %d players\n", time.Now().Format(time.RFC3339), args.ClientID, len(players))
	return nil
}

// parseFlags configura o servidor usando flags de linha de comando ou variáveis de ambiente
// Exemplo: go run server.go --port=8080 --ttl-player=30s
func (s *GameServer) parseFlags() {
	port := flag.Int("port", s.config.port, "Port to listen on")
	ttlProcessed := flag.Duration("ttl-processed", s.config.ttlProcessed, "TTL for processed commands")
	ttlPlayer := flag.Duration("ttl-player", s.config.ttlPlayer, "TTL for inactive players")

	// Também aceita via env vars
	if portEnv := os.Getenv("GAME_PORT"); portEnv != "" {
		if p, err := strconv.Atoi(portEnv); err == nil {
			*port = p
		}
	}

	flag.Parse()

	s.config.port = *port
	s.config.ttlProcessed = *ttlProcessed
	s.config.ttlPlayer = *ttlPlayer
}

// startCleanupRoutine inicia uma goroutine que periodicamente:
// - Remove jogadores inativos (sem atualização > ttlPlayer)
// - Limpa cache de comandos antigos (processados > ttlProcessed)
func (s *GameServer) startCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			s.mu.Lock()
			now := time.Now()

			// Limpa jogadores inativos
			for id, player := range s.players {
				lastSeen := time.Unix(player.LastSeen, 0)
				if now.Sub(lastSeen) > s.config.ttlPlayer {
					fmt.Printf("[SERVER] %s Removing inactive player %s (last seen %v ago)\n",
						time.Now().Format(time.RFC3339), id, now.Sub(lastSeen))
					delete(s.players, id)
				}
			}

			// Limpa comandos processados antigos
			for clientID, seqMap := range s.processedTimestamps {
				for seq, timestamp := range seqMap {
					if now.Sub(timestamp) > s.config.ttlProcessed {
						delete(s.processed[clientID], seq)
						delete(s.processedTimestamps[clientID], seq)
					}
				}
				// Remove maps vazios
				if len(s.processed[clientID]) == 0 {
					delete(s.processed, clientID)
					delete(s.processedTimestamps, clientID)
				}
			}

			s.mu.Unlock()
		}
	}()
}

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
