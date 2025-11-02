// main.go - Loop principal do jogo
package main

import (
	"fmt"
	"os"
	"time"
)

// INSTRUÇÕES DE INTEGRAÇÃO (Member B)
// --------------------------------------------------
// Este arquivo contém o loop principal do cliente. Para integrar o modo multiplayer
// via RPC, faça as seguintes alterações (comentadas aqui para orientar):
// 1) Gerar e persistir um ClientID único no início do programa (ex: gravar em `.clientid`).
// 2) Instanciar o RPC client: rpcClient := NewRPCClient(":12345", clientID)
// 3) Chamar rpcClient.SendCommand("REGISTER", map[string]interface{}{"name": "playerX"})
//    logo após carregar o mapa e criar o estado `jogo`.
// 4) Iniciar uma goroutine de polling que periodicamente chama rpcClient.GetState()
//    e atualiza `jogo.OtherPlayers` com o resultado. Intervalo recomendado: 200-500ms.
//    Exemplo (pseudocódigo):
//      go func() {
//          for {
//              state, err := rpcClient.GetState()
//              if err == nil { jogo.OtherPlayers = state.Players }
//              time.Sleep(300 * time.Millisecond)
//          }
//      }()
// 5) Ao processar eventos de teclado, garantir que `personagemExecutarAcao` chame
//    o rpcClient.SendCommand("UPDATE_POS", payload) logo após mover localmente.
// 6) Certificar-se de que todos os logs de RPC sejam visíveis no terminal para depuração.
// --------------------------------------------------

func main() {
	// Inicializa a interface (termbox)
	interfaceIniciar()
	defer interfaceFinalizar()

	// Usa "mapa.txt" como arquivo padrão ou lê o primeiro argumento
	mapaFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapaFile = os.Args[1]
	}

	for {
		canalMonstro := make(chan MonstroMsg)
		canalArmadilha := make(chan ArmadilhaMsg)
		canalMoeda := make(chan Moeda)
		canalMoedaColetada := make(chan MoedaColetadaMsg)
		canalTeclado := make(chan EventoTeclado)
		done := make(chan struct{}) //canal pra cancelar routines antigas

		// Inicializa o jogo
		jogo := jogoNovo()
		if err := jogoCarregarMapa(mapaFile, &jogo); err != nil {
			panic(err)
		}

		// Integração RPC (Member B) — gerar/persistir ClientID, registrar e iniciar polling
		rpcAddr := os.Getenv("RPC_ADDR")
		if rpcAddr == "" {
			rpcAddr = "127.0.0.1:12345"
		}
		clientIDFile := os.Getenv("CLIENTID_FILE")
		if clientIDFile == "" {
			clientIDFile = ".clientid"
		}
		clientID, err := LoadOrCreateClientID(clientIDFile)
		if err != nil {
			fmt.Printf("[CLIENT] erro ao obter ClientID: %v\n", err)
			clientID = "unknown-client"
		}
		rpcClient := NewRPCClient(rpcAddr, clientID)
		// Tenta registrar (não bloqueante)
		go func() {
			_, _ = rpcClient.SendCommand("REGISTER", map[string]interface{}{"name": clientID})
		}()

		// Polling para obter estado remoto e popular jogo.OtherPlayers
		go func() {
			for {
				state, err := rpcClient.GetState()
				if err == nil {
					jogo.OtherPlayers = state.Players
				}
				time.Sleep(300 * time.Millisecond)
			}
		}()

		jogo.Pontos = -1

		// Inicia a goroutine para ler eventos do teclado
		go interfaceLerEventoTeclado(canalTeclado)

		//cria o monstro
		monstro := &Monstro{X: 69, Y: 15}
		go monstroLoop(monstro, &jogo, canalMonstro, done)

		//cria as armadilhas
		armadilhas := []*Armadilha{
			{X: 6, Y: 14, Ativa: true, ID: 1},
			{X: 10, Y: 7, Ativa: true, ID: 1},
			{X: 20, Y: 5, Ativa: true, ID: 1},
			{X: 30, Y: 10, Ativa: true, ID: 1},
			{X: 40, Y: 15, Ativa: true, ID: 1},
			{X: 38, Y: 5, Ativa: true, ID: 1},
			{X: 60, Y: 8, Ativa: true, ID: 1},
			{X: 25, Y: 18, Ativa: true, ID: 1},
			{X: 11, Y: 19, Ativa: true, ID: 1},
			{X: 35, Y: 25, Ativa: true, ID: 1},
			{X: 51, Y: 4, Ativa: true, ID: 1},
			{X: 69, Y: 16, Ativa: true, ID: 1},
			{X: 46, Y: 11, Ativa: true, ID: 1},
			{X: 51, Y: 25, Ativa: true, ID: 1},
			{X: 3, Y: 3, Ativa: true, ID: 1},
			{X: 13, Y: 28, Ativa: true, ID: 1},
			{X: 45, Y: 20, Ativa: true, ID: 1},
			{X: 65, Y: 23, Ativa: true, ID: 1},
			{X: 74, Y: 26, Ativa: true, ID: 1},
			{X: 72, Y: 10, Ativa: true, ID: 10},
		}
		for _, a := range armadilhas {
			go armadilhaLoop(a, &jogo, canalArmadilha, done)
		}

		moeda := &Moeda{X: 6, Y: 10}
		go moedaLoop(moeda, &jogo, canalMoeda, canalMoedaColetada, done)

		// Desenha o estado inicial do jogo
		interfaceDesenharJogo(&jogo, armadilhas, moeda)

		//nova logica de jogo
		rodando := true
		for rodando {
			select {
			case msg := <-canalMonstro:
				jogo.MonstroX = msg.X
				jogo.MonstroY = msg.Y
				if msg.Encostou {
					jogo.StatusMsg = "O MONSTRO TE PEGOU, VOCE MORREU"
					interfaceDesenharJogo(&jogo, armadilhas, moeda)
					time.Sleep(2 * time.Second)

					// Exibe quantas moedas foram coletadas
					jogo.StatusMsg = "GAME OVER! Você coletou " + fmt.Sprintf("%d", jogo.Pontos) + " moedas antes de morrer. Pressione qualquer tecla para continuar..."
					interfaceDesenharJogo(&jogo, armadilhas, moeda)

					// Espera o jogador pressionar uma tecla para continuar
					<-canalTeclado
					rodando = false
				}
			case <-canalArmadilha:
				jogo.StatusMsg = "CAIU EM UMA ARMADILHA, VOCE MORREU"
				interfaceDesenharJogo(&jogo, armadilhas, moeda)
				time.Sleep(2 * time.Second)

				// Exibe quantas moedas foram coletadas
				jogo.StatusMsg = "GAME OVER! Você coletou " + fmt.Sprintf("%d", jogo.Pontos) + " moedas antes de morrer. Pressione qualquer tecla para continuar..."
				interfaceDesenharJogo(&jogo, armadilhas, moeda)

				// Espera o jogador pressionar uma tecla para continuar
				<-canalTeclado
				rodando = false
			case novaMoeda := <-canalMoeda:
				moeda.X = novaMoeda.X
				moeda.Y = novaMoeda.Y
			case msgMoedaColetada := <-canalMoedaColetada:
				if msgMoedaColetada.Coletada {
					// Feature de mudar a posi das armadilhas quando coletar moedas
					moverTodasArmadilhas(armadilhas, &jogo)
					jogo.StatusMsg = "Moeda coletada! Novas armadilhas foram posicionadas!"
				}
			case evento := <-canalTeclado:
				if continuar := personagemExecutarAcao(evento, &jogo); !continuar {
					return
				}
			case <-time.After(50 * time.Millisecond):
				// para atualizar a tela periodicamente
				interfaceDesenharJogo(&jogo, armadilhas, moeda)
			}
		}
		close(done)
	}
}
