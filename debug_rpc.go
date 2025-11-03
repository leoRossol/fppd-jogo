//go:build ignore
// +build ignore

// debug_rpc.go - util simples para testar conexão e GetState
package main

import (
	"fmt"
	"net/rpc"
	"time"
)

type ClientIDArgs struct {
	ClientID string
	Now      time.Time
}

type Player struct {
	ID       string
	X        int
	Y        int
	Lives    int
	LastSeen int64
}

type StateReply struct {
	Players    []Player
	ServerTime int64
}

func main() {
	addr := "127.0.0.1:12345"
	fmt.Printf("Tentando conectar ao servidor em %s...\n", addr)
	
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("ERRO: Não conseguiu conectar ao servidor: %v\n", err)
		fmt.Println("\nSolução: Inicie o servidor primeiro com:")
		fmt.Println("  go run -tags server .")
		return
	}
	defer client.Close()
	
	fmt.Println("✓ Conectado ao servidor!")
	
	// Testar GetState
	args := ClientIDArgs{ClientID: "debug-checker", Now: time.Now()}
	var reply StateReply
	
	err = client.Call("GameServer.GetState", &args, &reply)
	if err != nil {
		fmt.Printf("ERRO ao chamar GetState: %v\n", err)
		return
	}
	
	fmt.Printf("\n✓ GetState funcionou!\n")
	fmt.Printf("  Jogadores conectados: %d\n", len(reply.Players))
	fmt.Printf("  Timestamp do servidor: %d\n", reply.ServerTime)
	
	if len(reply.Players) == 0 {
		fmt.Println("\n⚠ Nenhum jogador registrado ainda.")
		fmt.Println("  Os clientes devem chamar REGISTER ao iniciar.")
	} else {
		fmt.Println("\n  Lista de jogadores:")
		for i, p := range reply.Players {
			fmt.Printf("    %d. ID=%s X=%d Y=%d Lives=%d LastSeen=%d\n", 
				i+1, p.ID, p.X, p.Y, p.Lives, p.LastSeen)
		}
	}
}
