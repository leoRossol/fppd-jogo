# Jogo de Terminal em Go (single-player + RPC multiplayer)

Este projeto é um jogo de terminal em Go que agora suporta modo multiplayer via RPC.
O cliente mantém toda a lógica do jogo e o servidor centraliza o estado dos jogadores (posições, vidas).

Principais pontos:
- Cliente: interface, lógica de movimentação, polling periódico para `GetState` no servidor.
- Servidor: mantém lista de jogadores e deduplicação exactly-once por ClientID+Seq; não contém lógica de movimentação nem mapa.

## Controles (single-player / cliente)

| Tecla | Ação |
|-------|------|
| W | Mover para cima |
| A | Mover para esquerda |
| S | Mover para baixo |
| D | Mover para direita |
| E | Interagir |
| ESC | Sair do jogo |

## Como rodar (modo RPC multiplayer)

Esta seção mostra passos organizados para rodar o servidor e o(s) cliente(s) tanto em modo de desenvolvimento (go run) quanto compilando binários.

Requisitos
- Go 1.18+ instalado
- Sistema: instruções abaixo usam PowerShell no Windows; em Bash (Linux/macOS) substitua as declarações de ambiente por `export VAR=valor`.

Passo rápido (modo dev)
- Abra um terminal para o servidor e rode:

```powershell
# inicia o servidor (usa o main com build tag `server`)
go run -tags server .
```

- Abra outro(s) terminal(is) para os clientes e rode em cada um:

```powershell
# inicia o cliente (interface + lógica local)
go run .
```

Build (gerar binários)
- Compilar servidor (gera `gameserver.exe` no Windows):

```powershell
go build -tags server -o gameserver.exe .
.\\gameserver.exe
```

- Compilar cliente (`jogo.exe`):

```powershell
go build -o jogo.exe .
.\\jogo.exe
```

Executando com variáveis de ambiente úteis
- Definir porta do servidor (ex.: 8080):

```powershell
$env:GAME_PORT = "8080"
go run -tags server .
```

- Forçar endereço do servidor ou ClientID no cliente:

```powershell
$env:SERVER_ADDR = "127.0.0.1:12345"
$env:CLIENT_ID = "playerA"
go run .
```

Rodar múltiplos clientes localmente
- Abra múltiplos terminais e execute `go run .` em cada um, ou use o script PowerShell `scripts/start_clients.ps1`:

```powershell
# inicia 2 clientes em novos terminais (PowerShell Core assumed for Start-Process pwsh)
.\\scripts\\start_clients.ps1 -Count 2
```

Polling / intervalos
- O cliente faz polling de `GetState` periodicamente (padrão 300ms). Para ajustar o intervalo, defina `POLL_MS` em milissegundos:

```powershell
$env:POLL_MS = "500"
go run .
```

Logs e depuração
- O servidor e o cliente imprimem informações relevantes no terminal para depuração (requisições recebidas, respostas, erros de RPC e retries).

Testes automatizados
- Para executar a suíte de testes:

```powershell
go test ./...
```

Scripts de ajuda
- Scripts adicionados em `scripts/`:
	- `start_server.ps1` — inicia o servidor (aceita parâmetro `-Port`).
	- `start_clients.ps1` — abre múltiplas instâncias do cliente em terminais novos.

Notas finais
- O servidor NÃO guarda o mapa nem a lógica de movimentação — isso continua sendo responsabilidade do cliente.
- A persistência de `Seq` no cliente é atômica (escreve em arquivo temporário e renomeia), garantindo resiliência contra crashes durante a escrita.

## Estado atual em relação aos requisitos do trabalho

Resumo curto:
- O servidor gerencia a sessão e o estado dos jogadores (posições, vidas). ✔
- O servidor não mantém o mapa nem a lógica de movimentação (fica no cliente). ✔
- Comunicação sempre iniciada pelos clientes; servidor apenas responde. ✔
- Cliente possui goroutine de polling para `GetState`. ✔
- Chamadas RPC têm retries/backoff implementados no cliente. ✔
- Exactly-once (deduplicação por ClientID+Seq) implementado no servidor com TTL e limpeza. ✔

Notas/pequenas recomendações: Persistência de Seq agora realizada de forma atômica no cliente; recomenda-se adicionar testes de falhas de rede.

## Estrutura do projeto (resumida)

- `server_main.go` — main do servidor (build tag `server`).
- `server.go` — implementação do `GameServer`, deduplicação e `GetState`.
- `client_rpc.go` — cliente RPC com retries e persistência de Seq.
- `main.go` — cliente/jogo com loop principal e integração RPC.
- `jogo.go`, `personagem.go`, `interface.go` — lógica do jogo local e UI.
- `rpc_types.go` — tipos compartilhados (PlayerInfo, CommandArgs, etc.).
- `server_rpc_test.go` — teste que valida exactly-once e GetState via RPC.

---

Se quiser, eu adiciono um passo-a-passo mais enxuto para apresentação (1 slide) ou crio scripts `start_server.ps1` / `start_clients.ps1` para automatizar demos locais — quer que eu crie isso agora?


