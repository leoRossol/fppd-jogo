### O que foi implementado

- `client_rpc.go`: cliente RPC com reconexão, retries e `Seq` monotônico.
- `main.go`: geração/persistência do `ClientID`, instância do `RPCClient`, `REGISTER` no início do round e goroutine de polling `GetState` que preenche `jogo.OtherPlayers`.
- `personagem.go`: envio de `UPDATE_POS` após cada movimento local.
- `interface.go`: renderização dos outros jogadores (não duplica o jogador local).
- `debuglog.go` + `ui_debugpanel.go`: logs silenciosos por padrão; painel de debug opcional no rodapé (`DEBUG_PANEL=1`).
- `server.go`: isolado por build tag `//go:build server`.


- Ao abrir dois clientes com `CLIENT_ID` distintos:
  - A barra de status mostra “Jogadores Online: 2”.
  - Movimentar um cliente reflete a posição no outro em ~300ms.
  - O painel/arquivo de logs registra `REGISTER`, `UPDATE_POS` e `GetState` funcionando.