package main

import (
	"encoding/gob"
	"time"
)

// rpc_types.go
// ----------------
// Tipos compartilhados entre cliente e servidor via RPC.
// Este arquivo deve permanecer sincronizado entre cliente e servidor.
//
// NOTAS IMPORTANTES PARA INTEGRAÇÃO (colocar comentários somente aqui):
// - Para integrar o cliente existente (main.go):
//   1) No início do `main()` do cliente, gerar um ClientID único (por exemplo com uuid).
//   2) Instanciar um `RPCClient` (em `client_rpc.go`) apontando para a porta do servidor (p.ex. ":12345").
//   3) Chamar `rpcClient.SendCommand("REGISTER", ...)` para registrar o jogador no servidor.
//   4) Criar uma goroutine de polling que chama `rpcClient.GetState()` periodicamente (p.ex. 200-500ms)
//      e atualiza a lista local de jogadores para renderização.
//
// - Modificações necessárias nos ficheiros antigos (resumo):
//   * `main.go`  : inicializar `RPCClient`, iniciar polling goroutine e passar `rpcClient` onde necessário.
//   * `personagem.go` : ao mover o jogador (em `personagemMover` ou `personagemExecutarAcao`), chamar
//                       `rpcClient.SendCommand("UPDATE_POS", map[string]interface{}{ "x": nx, "y": ny, "lives": jogo.Pontos })`.
//   * `jogo.go` : opcionalmente, não mover a lógica de movimento para o servidor; o servidor apenas armazena
//                posições reportadas pelos clientes. Mantenha `jogoMoverElemento` e `jogoPodeMoverPara` no cliente.
//   * `interface.go` : alterar `interfaceDesenharJogo` para desenhar também os outros jogadores (iteração sobre
//                     lista retornada por `GetState()`). Use `PlayerInfo.ID` para evitar desenhar o jogador local duas vezes.
//
// Tipos RPC usados pelo servidor/cliente
type PlayerInfo struct {
	ID       string
	X, Y     int
	Lives    int
	LastSeen int64 // unix timestamp
}

type StateReply struct {
	Players    []PlayerInfo
	ServerTime int64
}

// Payloads tipados para comunicação RPC
type RegisterPayload struct {
	Name string
	X, Y int
}

type UpdatePosPayload struct {
	X, Y  int
	Lives int
}

// CommandArgs representa um comando enviado pelo cliente ao servidor
// Payload agora é interface{} para suportar structs tipados. Os tipos
// precisam ser registrados com gob para permitir serialização via net/rpc.
type CommandArgs struct {
	ClientID string
	Seq      int64
	Cmd      string
	Payload  interface{}
}

type CommandReply struct {
	Seq     int64
	Applied bool
	Message string
}

type ClientIDArgs struct {
	ClientID string
	Now      time.Time
}

func init() {
	// Registrar os tipos usados para que encoding/gob consiga codificar/decodificar
	gob.Register(RegisterPayload{})
	gob.Register(UpdatePosPayload{})
}

// Validação simples para UpdatePosPayload
// Retorna nil se tudo OK, ou um erro descrevendo o problema.
func ValidateUpdatePos(p UpdatePosPayload) error {
	if p.X < 0 || p.Y < 0 {
		return &ValidationError{"coordinates must be non-negative"}
	}
	if p.Lives < 0 {
		return &ValidationError{"lives must be non-negative"}
	}
	return nil
}

// Validação simples para RegisterPayload
func ValidateRegister(r RegisterPayload) error {
	if r.Name == "" {
		return &ValidationError{"name required"}
	}
	if r.X < 0 || r.Y < 0 {
		return &ValidationError{"coordinates must be non-negative"}
	}
	return nil
}

// ValidationError é usado nas funções de validação para testes e mensagens claras.
type ValidationError struct{
	Msg string
}

func (e *ValidationError) Error() string { return e.Msg }

// FIM rpc_types.go
