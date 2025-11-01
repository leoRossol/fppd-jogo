Equipe (5 pessoas) — separação detalhada e paralelizável

Este ficheiro lista tarefas claras, por membro, até a finalização do trabalho. Foi pensado para que os cinco membros trabalhem em paralelo com dependências mínimas.

Regras gerais
- Cada alteração de código deve conter comentários TODO explicando o porquê e o que foi alterado.
- Commits pequenos e claros; um pull request por membro quando terminar a sua parte.
- Testes manuais rápidos sempre que uma peça ficar pronta (ver critérios de aceitação abaixo).

1) Member A — Servidor (core)
Prioridade: alta — entrega necessária antes do teste de integração completo.
Objetivos:
- Finalizar `server.go` (lógica principal já esqueleto criada).
- Implementar deduplicação exactly-once (já existe in-memory) com:
  - TTL / limpeza periódica de entradas antigas em `processed`.
  - Remoção de `players` inativos (LastSeen TTL).
  - Tornar porta configurável via flag/env.
- (Opcional) Persistência simples para `processed` (checkpoint em ficheiro) se o grupo decidir que a garantia deve sobreviver a reinícios.
Tarefas concretas (arquivos a editar):
- `server.go`: implementar goroutine de limpeza, adicionar flags (porta, ttlProcessed, ttlPlayer).
- `rpc_types.go`: (se escolher payloads tipados) ajustar tipos usados pelo servidor.
Critérios de aceitação:
- `server.go` compila (`go run -tags server server.go`) e responde a `Register`/`UpdatePos`.
- Após N minutos (configurado) um player inativo deixa de aparecer em `GetState`.

2) Member B — Cliente / Integração (core)
Prioridade: alta — integração com o jogo local para poder testar em paralelo.
Objetivos:
- Integrar `client_rpc.go` no cliente em `main.go`.
- Gerar e persistir `ClientID` (ficheiro `.clientid`) para manter identidade entre execuções.
- Implementar goroutine de polling que chama `rpcClient.GetState()` (intervalo configurável) e popula `jogo.OtherPlayers`.
- Enviar `UPDATE_POS` via `rpcClient.SendCommand` sempre que o jogador se mover (após movimentação local).
Tarefas concretas (arquivos a editar):
- `main.go`: gerar/persistir ClientID, instanciar `RPCClient`, chamar `REGISTER`, criar polling goroutine.
- `personagem.go`: após `personagemMover` enviar `UPDATE_POS` (payload tipado ou map).
- `interface.go`: desenhar `jogo.OtherPlayers` sem duplicar o jogador local (usar `LocalClientID`).
Critérios de aceitação:
- Cliente registra-se no servidor e aparece em `GetState` de outro cliente.
- Quando o jogador se move, a posição é enviada e refletida em `GetState` de outros clientes.

3) Member C — Tipos e testes unitários
Prioridade: média — segurança e robustez da serialização.
Objetivos:
- Definir payloads tipados em `rpc_types.go` (por exemplo `UpdatePosPayload{X,Y,Lives int}`) e substituir `map[string]interface{}`.
- Escrever testes unitários para:
  - Serialização/Deserialização entre cliente/servidor (gob/json conforme `net/rpc`).
  - Funções utilitárias (ex.: validação de payload).
Tarefas concretas (arquivos a editar/criar):
- `rpc_types.go`: adicionar structs de payload e comentários.
- `rpc_types_test.go`: testes unitários.
Critérios de aceitação:
- `go test ./...` passa nos testes criados por Member C.

4) Member D — Testes de integração e cenários
Prioridade: média — validação e cenários reais.
Objetivos:
- Criar rotinas/scripts manuais e automáticas para executar cenários: registro, movimento, retransmissão (timeout), e remoção por inatividade.
- Produzir logs determinísticos que permitam verificar exactly-once.
Tarefas concretas:
- Scripts PowerShell para iniciar servidor + 2 clientes em janelas diferentes (p.ex. `start-process`), coletar saídas em `tests/logs/`.
- Scripts que forçam falhas (p.ex. matar processo do servidor, pausar rede local) para testar retransmissão.
- Documentar os comandos e passos para reproduzir manualmente.
Critérios de aceitação:
- Cenários A–D (veja abaixo) passam quando executados pelo script de Member E ou manualmente.

5) Member E — Automação, harness e documentação de execução
Prioridade: média — permite validação rápida e entrega organizada.
Objetivos:
- Criar scripts de automação e um pequeno harness que:
  - Inicie o servidor (`go run -tags server server.go`).
  - Inicie dois clientes (`go run main.go`) em janelas separadas e redirecione logs para `tests/logs/`.
  - Execute testes automatizados (simular timeout e reenviar seq) e valide condições (ex.: posição sincronizada).
- Gerar relatório curto (texto) com resultados dos testes e onde encontrar logs.
Tarefas concretas:
- `scripts/start_env.ps1`: iniciar servidor + 2 clientes e coletar logs.
- `scripts/run_tests.ps1`: executar cenários automáticos (timeout, resend) e produzir PASS/FAIL.
- Organizar logs em `tests/logs/YYYYMMDD_HHMMSS/`.
- Atualizar `README.md` com instruções de build/exec e exemplos de uso dos scripts.
Critérios de aceitação:
- Os scripts conseguem rodar os cenários básicos de forma automática e geram um relatório PASS/FAIL.

Dependências e paralelismo
- Member A e B são críticos: ambos podem trabalhar em paralelo sobre contratos estáveis (tipos RPC). Member C deve definir tipos cedo para evitar retrabalho.
- Member D e E podem trabalhar paralelamente com Member A/B: criar scripts, executar e reportar bugs.

Cenários de teste principais (A–D) — PASS/FAIL
- A: Registro e Polling — cliente A registra, cliente B vê A no `GetState`.
- B: Movimento reportado — A move, B vê a nova posição via `GetState`.
- C: Retransmissão/exactly-once — simular perda de resposta e reenviar mesmo seq; servidor aplica apenas uma vez.
- D: Remoção por inatividade — client desaparece após TTL configurado.

Checklist final antes de entregar
1. Integração completa (Member B) funcionando localmente com 2 clientes.
2. Servidor (Member A) com TTL/limpeza funcionando.
3. Tipos tipados e testes unitários (Member C) passando.
4. Scripts de integração e simulação (Member D) prontos.
5. Harness e documentação (Member E) prontos e testados.