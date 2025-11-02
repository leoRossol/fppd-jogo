# Jogo de Terminal em Go

Este projeto é um pequeno jogo desenvolvido em Go que roda no terminal usando a biblioteca [termbox-go](https://github.com/nsf/termbox-go). O jogador controla um personagem que pode se mover por um mapa carregado de um arquivo de texto.

## Como funciona

- O mapa é carregado de um arquivo `.txt` contendo caracteres que representam diferentes elementos do jogo.
- O personagem se move com as teclas **W**, **A**, **S**, **D**.
- Pressione **E** para interagir com o ambiente.
- Pressione **ESC** para sair do jogo.

### Controles

| Tecla | Ação              |
|-------|-------------------|
| W     | Mover para cima   |
| A     | Mover para esquerda |
| S     | Mover para baixo  |
| D     | Mover para direita |
| E     | Interagir         |
| ESC   | Sair do jogo      |

## Como rodar o servidor

1. Compile o server.go junto com o rpc_types.go

```bash
go build -tags server -o gameserver.exe server.go rpc_types.go
```

2. Rode 
```bash
.\gameserver.exe
```

### Rodando o cliente

Variáveis de ambiente (opcionais):
- `SERVER_ADDR` endereço do servidor (padrão: 127.0.0.1:12345)
- `CLIENT_ID` força um ID de cliente específico (senão usa/persiste `.clientid`)
- `CLIENT_ID_FILE` caminho do arquivo para persistir o ID (padrão: `.clientid`)
- `POLL_MS` intervalo de polling do estado em milissegundos (padrão: 300)
- `DEBUG_PANEL` se `1`, exibe um painel de logs no rodapé da tela
- `CLIENT_LOG` caminho de arquivo para logs do cliente (não polui a tela)
- `DEBUG` se `1`, envia logs para stderr (pode sobrepor a tela; use só para depuração rápida)

Exemplo (dois clientes no mesmo diretório):
```bash
# Terminal 1 (servidor)
go run -tags server ./server.go ./rpc_types.go

# Terminal 2 (cliente A)
export SERVER_ADDR=127.0.0.1:12345
export CLIENT_ID=A1
export DEBUG_PANEL=1        # painel de logs no rodapé (opcional)
# export CLIENT_LOG=clientA.log  # alternativa: logs em arquivo
go run .

# Terminal 3 (cliente B)
export SERVER_ADDR=127.0.0.1:12345
export CLIENT_ID=B2
export DEBUG_PANEL=1
go run .
```

## Como compilar

1. Instale o Go e clone este repositório.
2. Inicialize um novo módulo "jogo":

```bash
go mod init jogo
go get -u github.com/nsf/termbox-go
```

3. Compile o programa:

Linux:

```bash
go build -o jogo
```

Windows:

```bash
go build -o jogo.exe
```

Também é possivel compilar o projeto usando o comando `make` no Linux ou o script `build.bat` no Windows.

## Como executar

1. Certifique-se de ter o arquivo `mapa.txt` com um mapa válido.
2. Execute o programa no termimal:

```bash
./jogo
```

## Estrutura do projeto

- main.go — Ponto de entrada e loop principal
- interface.go — Entrada, saída e renderização com termbox
- jogo.go — Estruturas e lógica do estado do jogo
- personagem.go — Ações do jogador


