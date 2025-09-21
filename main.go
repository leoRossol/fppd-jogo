// main.go - Loop principal do jogo
package main

import (
	"os"
	"time"
)

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
		canalTeclado := make(chan EventoTeclado)
		done := make(chan struct{}) //canal pra cancelar routines antigas

		// Inicializa o jogo
		jogo := jogoNovo()
		if err := jogoCarregarMapa(mapaFile, &jogo); err != nil {
			panic(err)
		}

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
		go moedaLoop(moeda, &jogo, canalMoeda, done)

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
					rodando = false
				}
			case <-canalArmadilha:
				jogo.StatusMsg = "CAIU EM UMA ARMADILHA, VOCE MORREU"
				interfaceDesenharJogo(&jogo, armadilhas, moeda)
				time.Sleep(2 * time.Second)
				rodando = false
			case novaMoeda := <-canalMoeda:
				moeda.X = novaMoeda.X
				moeda.Y = novaMoeda.Y
				// Removemos o increment do contador daqui, pois agora ele acontece diretamente no moedaLoop
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
