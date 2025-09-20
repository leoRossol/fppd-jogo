// jogo.go - Funções para manipular os elementos do jogo, como carregar o mapa e mover o personagem
package main

import (
	"bufio"
	"math/rand"
	"os"
	"time"
)

// Elemento representa qualquer objeto do mapa (parede, personagem, vegetação, etc)
type Elemento struct {
	simbolo  rune
	cor      Cor
	corFundo Cor
	tangivel bool // Indica se o elemento bloqueia passagem
}

// Jogo contém o estado atual do jogo
type Jogo struct {
	Mapa               [][]Elemento // grade 2D representando o mapa
	PosX, PosY         int          // posição atual do personagem
	UltimoVisitado     Elemento     // elemento que estava na posição do personagem antes de mover
	StatusMsg          string       // mensagem para a barra de status
	MonstroX, MonstroY int          //posicao atual do monstro
	Pontos             int          //moedas coletadas
}

// mensagem do monstro
type MonstroMsg struct {
	X, Y     int
	Encostou bool
}

// canal para a comunicacao do monstro
var canalMonstro = make(chan MonstroMsg)

// armadilhas
type ArmadilhaMsg struct {
	ID      int
	X, Y    int
	Ativada bool
}

// canal armadilhas
var canalArmadilha = make(chan ArmadilhaMsg)

// Elementos visuais do jogo
var (
	Personagem    = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Inimigo       = Elemento{'☠', CorPadrao, CorPadrao, true}
	MonstroElem   = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede        = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao     = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio         = Elemento{' ', CorPadrao, CorPadrao, false}
	ArmadilhaElem = Elemento{'Δ', CorVermelho, CorPadrao, false}
	MoedaElem     = Elemento{'$', CorAmarelo, CorPadrao, false}
)

// Cria e retorna uma nova instância do jogo
func jogoNovo() Jogo {
	// O ultimo elemento visitado é inicializado como vazio
	// pois o jogo começa com o personagem em uma posição vazia
	return Jogo{UltimoVisitado: Vazio}
}

// Lê um arquivo texto linha por linha e constrói o mapa do jogo
func jogoCarregarMapa(nome string, jogo *Jogo) error {
	arq, err := os.Open(nome)
	if err != nil {
		return err
	}
	defer arq.Close()

	scanner := bufio.NewScanner(arq)
	y := 0
	for scanner.Scan() {
		linha := scanner.Text()
		var linhaElems []Elemento
		for x, ch := range linha {
			e := Vazio
			switch ch {
			case Parede.simbolo:
				e = Parede
			case Inimigo.simbolo:
				e = Inimigo
			case Vegetacao.simbolo:
				e = Vegetacao
			case Personagem.simbolo:
				jogo.PosX, jogo.PosY = x, y // registra a posição inicial do personagem
			}
			linhaElems = append(linhaElems, e)
		}
		jogo.Mapa = append(jogo.Mapa, linhaElems)
		y++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// Verifica se o personagem pode se mover para a posição (x, y)
func jogoPodeMoverPara(jogo *Jogo, x, y int) bool {
	// Verifica se a coordenada Y está dentro dos limites verticais do mapa
	if y < 0 || y >= len(jogo.Mapa) {
		return false
	}

	// Verifica se a coordenada X está dentro dos limites horizontais do mapa
	if x < 0 || x >= len(jogo.Mapa[y]) {
		return false
	}

	// Verifica se o elemento de destino é tangível (bloqueia passagem)
	if jogo.Mapa[y][x].tangivel {
		return false
	}

	// Pode mover para a posição
	return true
}

// Move um elemento para a nova posição
func jogoMoverElemento(jogo *Jogo, x, y, dx, dy int) {
	nx, ny := x+dx, y+dy

	// Obtem elemento atual na posição
	elemento := jogo.Mapa[y][x] // guarda o conteúdo atual da posição

	jogo.Mapa[y][x] = jogo.UltimoVisitado   // restaura o conteúdo anterior
	jogo.UltimoVisitado = jogo.Mapa[ny][nx] // guarda o conteúdo atual da nova posição
	jogo.Mapa[ny][nx] = elemento            // move o elemento
}

// controla a monstro
func monstroLoop(monstro *Monstro, jogo *Jogo) {
	for {
		//move em direcao ao player
		monstroMover(monstro, jogo.PosX, jogo.PosY, jogo)
		//verifica se encostou no player
		encostou := monstroEncostou(monstro, jogo)
		//envia mensagem ao controlador do jogo
		canalMonstro <- MonstroMsg{X: monstro.X, Y: monstro.Y, Encostou: encostou}

		//MOVIMENTO ADICIONAL PARA (GAMBIARRA)
		monstroMover(monstro, jogo.PosX, jogo.PosY, jogo)
		//verifica se encostou no player
		encostou = monstroEncostou(monstro, jogo)
		//envia mensagem ao controlador do jogo
		canalMonstro <- MonstroMsg{X: monstro.X, Y: monstro.Y, Encostou: encostou}

		//delay para o monstro ir devagar
		time.Sleep(1000 * time.Millisecond)
	}
}

// controla as armadilhas
func armadilhaLoop(armadilha *Armadilha, jogo *Jogo) {
	for {
		if armadilhaAtivada(armadilha, jogo) {
			canalArmadilha <- ArmadilhaMsg{
				ID:      armadilha.ID,
				X:       armadilha.X,
				Y:       armadilha.Y,
				Ativada: true,
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func moedaColetada(moeda *Moeda, jogo *Jogo) bool {
	return moeda.X == jogo.PosX && moeda.Y == jogo.PosY
}

func moedaLoop(moeda *Moeda, jogo *Jogo) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		//player tem que coletar a moeda
		for !moedaColetada(moeda, jogo) {
			time.Sleep(50 * time.Millisecond)
		}

		//apos coletada, escolhe nova posição aleatória válida
		for {
			nx := r.Intn(len(jogo.Mapa[0]))
			ny := r.Intn(len(jogo.Mapa))
			if jogoPodeMoverPara(jogo, nx, ny) && (nx != jogo.PosX || ny != jogo.PosY) {
				moeda.X = nx
				moeda.Y = ny
				break
			}
		}
		//notifica o loop principal sobre a nova posição
		canalMoeda <- *moeda
	}
}

// canal dessa merda
var canalMoeda = make(chan Moeda)
