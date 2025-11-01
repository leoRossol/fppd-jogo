package main

import (
	"fmt"
	"net/rpc"
	"sync"
	"time"
)

// client_rpc.go
// ----------------
// Este arquivo fornece um cliente RPC com retries e geração de Seq.
// Abaixo há instruções detalhadas (destinadas ao Member B) sobre como integrar
// este cliente ao projeto existente (`main.go`, `personagem.go`, `interface.go`).
//
// Member B (Cliente / Integração) - Passo-a-passo recomendado:
// 1) Persistir/gerar ClientID no cliente
//    - No início do `main()` (arquivo `main.go`) gere um ClientID único (UUID).
//    - Persistir este ClientID num ficheiro oculto local (ex: `.clientid`) para que
//      o mesmo ID seja reutilizado entre execuções (importante para exactly-once).
//
// 2) Instanciar RPCClient
//    - rpcClient := NewRPCClient(":12345", clientID)
//    - Depois de criar rpcClient, chamar rpcClient.SendCommand("REGISTER", map[string]interface{}{"name": "playerX"})
//      para que o servidor registre o jogador.
//
// 3) Polling (goroutine)
//    - Criar uma goroutine no `main.go` que a cada 200-500ms chama rpcClient.GetState()
//      e atualiza uma nova lista em `Jogo`, por exemplo `jogo.OtherPlayers []PlayerInfo`.
//    - Exemplo (pseudocódigo):
//      go func() {
//          for {
//              state, err := rpcClient.GetState()
//              if err == nil { jogo.OtherPlayers = state.Players }
//              time.Sleep(300 * time.Millisecond)
//          }
//      }()
//
// 4) Envio de comandos ao mover o personagem
//    - Em `personagemMover` (ou logo após mover localmente), enviar atualização de posição:
//      payload := map[string]interface{}{"x": nx, "y": ny, "lives": jogo.Pontos}
//      rpcClient.SendCommand("UPDATE_POS", payload)
//    - Use sempre o mesmo ClientID e não reinicie Seq ao reiniciar o cliente (se possível persistir lastSeq).
//
// 5) Renderização dos outros jogadores
//    - Em `interfaceDesenharJogo` (arquivo `interface.go`) desenhar `jogo.OtherPlayers` sobre o mapa.
//    - Evitar desenhar o jogador local duas vezes: compare `PlayerInfo.ID` com o `ClientID` local.
//
// 6) Logging e erros
//    - Todos os pedidos e respostas RPC já são logados aqui (prints). Certifique-se de que
//      `main.go` redirecione/registre esses logs ou que o terminal do cliente fique visível
//      para depuração conforme o enunciado.
//
// 7) Persistência de Seq (opcional, recomendado)
//    - Para maior robustez, guarde `lastSeq` em disco sempre que incrementar, ou reenvie comandos
//      idempotentes com o mesmo seq quando for necessário.
//
// NOTAS sobre tipos de payload
// - Atualmente `SendCommand` usa `map[string]interface{}` para payloads. Isso funciona com `net/rpc`/gob,
//   mas é mais robusto criar structs concretos (ex.: `UpdatePosPayload`) em `rpc_types.go`.

// RPCClient encapsula chamadas RPC com retries e geração de seq
type RPCClient struct {
	addr     string
	mu       sync.Mutex
	client   *rpc.Client
	ClientID string
	Seq      int64
}

func NewRPCClient(addr, clientID string) *RPCClient {
	return &RPCClient{addr: addr, ClientID: clientID}
}

func (r *RPCClient) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.client != nil {
		return nil
	}
	var err error
	// tentativas simples com backoff
	backoff := 100 * time.Millisecond
	for i := 0; i < 5; i++ {
		r.client, err = rpc.Dial("tcp", r.addr)
		if err == nil {
			return nil
		}
		time.Sleep(backoff)
		backoff *= 2
	}
	return err
}

// SendCommand envia um comando para o servidor com retries; garante que o mesmo seq seja transmitido nas retransmissões
func (r *RPCClient) SendCommand(cmd string, payload map[string]interface{}) (CommandReply, error) {
	// prepara args
	r.mu.Lock()
	r.Seq++
	seq := r.Seq
	args := CommandArgs{ClientID: r.ClientID, Seq: seq, Cmd: cmd, Payload: payload}
	r.mu.Unlock()

	var reply CommandReply
	// conectar se necessário
	if err := r.connect(); err != nil {
		return reply, err
	}

	// retries com backoff; como usamos seq, reexecução é tolerante (server detecta duplicados)
	backoff := 100 * time.Millisecond
	for i := 0; i < 5; i++ {
		fmt.Printf("[CLIENT] Sending SendCommand to %s seq=%d cmd=%s\n", r.addr, seq, cmd)
		callErr := r.client.Call("GameServer.SendCommand", &args, &reply)
		if callErr == nil {
			fmt.Printf("[CLIENT] Got reply for seq=%d: %+v\n", seq, reply)
			return reply, nil
		}
		fmt.Printf("[CLIENT] SendCommand error: %v - retrying...\n", callErr)
		time.Sleep(backoff)
		backoff *= 2
		// reconectar antes da próxima tentativa
		r.client = nil
		if err := r.connect(); err != nil {
			return reply, err
		}
	}
	return reply, fmt.Errorf("SendCommand failed after retries")
}

// GetState solicita o estado atual do servidor (polling)
func (r *RPCClient) GetState() (StateReply, error) {
	var reply StateReply
	if err := r.connect(); err != nil {
		return reply, err
	}

	args := ClientIDArgs{ClientID: r.ClientID, Now: time.Now()}
	backoff := 100 * time.Millisecond
	for i := 0; i < 5; i++ {
		fmt.Printf("[CLIENT] Requesting GetState from %s\n", r.addr)
		callErr := r.client.Call("GameServer.GetState", &args, &reply)
		if callErr == nil {
			fmt.Printf("[CLIENT] Received state with %d players\n", len(reply.Players))
			return reply, nil
		}
		fmt.Printf("[CLIENT] GetState error: %v - retrying...\n", callErr)
		time.Sleep(backoff)
		backoff *= 2
		r.client = nil
		if err := r.connect(); err != nil {
			return reply, err
		}
	}
	return reply, fmt.Errorf("GetState failed after retries")
}
