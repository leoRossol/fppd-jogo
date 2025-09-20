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

	// Inicializa o jogo
	jogo := jogoNovo()
	if err := jogoCarregarMapa(mapaFile, &jogo); err != nil {
		panic(err)
	}

	//cria o monstro
	monstro := &Monstro{X: 1, Y: 1}
	go monstroLoop(monstro, &jogo)

	//cria as armadilhas
	armadilhas := []*Armadilha{
		{X: 6, Y: 14, Ativa: true, ID: 1},
		{X: 10, Y: 7, Ativa: true, ID: 2},
		{X: 20, Y: 5, Ativa: true, ID: 3},
		{X: 30, Y: 10, Ativa: true, ID: 4},
		{X: 40, Y: 15, Ativa: true, ID: 5},
		{X: 50, Y: 20, Ativa: true, ID: 6},
		{X: 60, Y: 8, Ativa: true, ID: 7},
		{X: 25, Y: 18, Ativa: true, ID: 8},
		{X: 15, Y: 22, Ativa: true, ID: 9},
		{X: 35, Y: 25, Ativa: true, ID: 10},
	}
	for _, a := range armadilhas {
		go armadilhaLoop(a, &jogo)
	}

	moeda := &Moeda{X: 2, Y: 2}
	go moedaLoop(moeda, &jogo)

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
			jogo.Pontos++
		case <-time.After(50 * time.Millisecond):
			//para não travar o loop
		}

		evento := interfaceLerEventoTeclado()
		if continuar := personagemExecutarAcao(evento, &jogo); !continuar {
			rodando = false
		}
		interfaceDesenharJogo(&jogo, armadilhas, moeda)
	}
}
