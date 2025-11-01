// interface.go - Interface gráfica do jogo usando termbox
// O código abaixo implementa a interface gráfica do jogo usando a biblioteca termbox-go.
// A biblioteca termbox-go é uma biblioteca de interface de terminal que permite desenhar
// elementos na tela, capturar eventos do teclado e gerenciar a aparência do terminal.

package main

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

// Define um tipo Cor para encapsuladar as cores do termbox
type Cor = termbox.Attribute

// Definições de cores utilizadas no jogo
const (
	CorPadrao      Cor = termbox.ColorDefault
	CorCinzaEscuro     = termbox.ColorDarkGray
	CorVermelho        = termbox.ColorRed
	CorVerde           = termbox.ColorGreen
	CorAmarelo         = termbox.ColorYellow
	CorParede          = termbox.ColorBlack | termbox.AttrBold | termbox.AttrDim
	CorFundoParede     = termbox.ColorDarkGray
	CorTexto           = termbox.ColorDarkGray
	CorMagenta         = termbox.ColorLightMagenta
)

// EventoTeclado representa uma ação detectada do teclado (como mover, sair ou interagir)
type EventoTeclado struct {
	Tipo  string // "sair", "interagir", "mover"
	Tecla rune   // Tecla pressionada, usada no caso de movimento
}

// Inicializa a interface gráfica usando termbox
func interfaceIniciar() {
	if err := termbox.Init(); err != nil {
		panic(err)
	}
}

// Encerra o uso da interface termbox
func interfaceFinalizar() {
	termbox.Close()
}

// Lê um evento do teclado e o traduz para um EventoTeclado
func interfaceLerEventoTeclado(canal chan EventoTeclado) {
	for {
		ev := termbox.PollEvent()
		if ev.Type != termbox.EventKey {
			continue
		}

		var evento EventoTeclado
		if ev.Key == termbox.KeyEsc {
			evento = EventoTeclado{Tipo: "sair"}
		} else if ev.Ch == 'e' {
			evento = EventoTeclado{Tipo: "interagir"}
		} else {
			evento = EventoTeclado{Tipo: "mover", Tecla: ev.Ch}
		}
		canal <- evento
	}
}

// Renderiza todo o estado atual do jogo na tela
func interfaceDesenharJogo(jogo *Jogo, armadilhas []*Armadilha, moeda *Moeda) {
	interfaceLimparTela()

	// Desenha todos os elementos do mapa
	for y, linha := range jogo.Mapa {
		for x, elem := range linha {
			interfaceDesenharElemento(x, y, elem)
		}
	}

	//desenhar armadilhas sobre o mapa
	for _, a := range armadilhas {
		if a.Ativa {
			interfaceDesenharElemento(a.X, a.Y, ArmadilhaElem)
		}
	}

	//desenha a moeda sobre o mapa
	interfaceDesenharElemento(moeda.X, moeda.Y, MoedaElem)

	//desenha o monstro sobre o mapa
	interfaceDesenharElemento(jogo.MonstroX, jogo.MonstroY, MonstroElem)

	// Desenha o personagem sobre o mapa
	interfaceDesenharElemento(jogo.PosX, jogo.PosY, Personagem)

	// TODO Member B: desenhar outros jogadores reportados pelo servidor
	// - `jogo.OtherPlayers` deve ser preenchido pela goroutine de polling que chama rpcClient.GetState()
	// - Evite desenhar o jogador local novamente: compare PlayerInfo.ID com o ClientID local
	//   (é preciso que o ClientID local esteja disponível para a comparação; pode ser
	//    armazenado em uma variável global `LocalClientID` ou passado para a função).
	// Exemplo comentado (não ativar sem preparar ClientID/global):
	// for _, p := range jogo.OtherPlayers {
	//     if p.ID == LocalClientID { continue }
	//     // desenhar um símbolo simples para outros jogadores, por exemplo '☺' com cor diferente
	//     interfaceDesenharElemento(p.X, p.Y, Elemento{simbolo: '☺', cor: CorAmarelo, corFundo: CorPadrao, tangivel: true})
	// }

	// Desenha a barra de status
	interfaceDesenharBarraDeStatus(jogo)

	// Força a atualização do terminal
	interfaceAtualizarTela()
}

// Limpa a tela do terminal
func interfaceLimparTela() {
	termbox.Clear(CorPadrao, CorPadrao)
}

// Força a atualização da tela do terminal com os dados desenhados
func interfaceAtualizarTela() {
	termbox.Flush()
}

// Desenha um elemento na posição (x, y)
func interfaceDesenharElemento(x, y int, elem Elemento) {
	termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
}

// Exibe uma barra de status com informações úteis ao jogador
func interfaceDesenharBarraDeStatus(jogo *Jogo) {
	// Linha de status dinâmica
	status := fmt.Sprintf("%s | Moedas: %d", jogo.StatusMsg, jogo.Pontos)
	for i, c := range status {
		termbox.SetCell(i, len(jogo.Mapa)+1, c, CorTexto, CorPadrao)
	}

	// Instruções fixas
	msg := "Use WASD para mover e E para interagir. ESC para sair."
	for i, c := range msg {
		termbox.SetCell(i, len(jogo.Mapa)+3, c, CorTexto, CorPadrao)
	}
}
