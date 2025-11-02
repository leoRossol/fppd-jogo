// personagem.go - Funções para movimentação e ações do personagem
package main

import "fmt"

// personagem.go
// --------------------------------------------------
// Comentários/TODOs para Member B:
// - Depois de integrar o RPC client, enviar um comando UPDATE_POS ao servidor sempre
//   que o jogador se mover localmente. Não mova a lógica de movimentação para o servidor;
//   o cliente continua sendo autoridade sobre seu próprio movimento.
// - Exemplo de payload:
//     payload := map[string]interface{}{"x": nx, "y": ny, "lives": jogo.Pontos}
//     rpcClient.SendCommand("UPDATE_POS", payload)
// - Armazene/recupere o rpcClient a partir de uma variável global acessível aqui ou passe-o
//   como parâmetro (ex: personagemMover(tecla, jogo, rpcClient)).
// --------------------------------------------------

// Atualiza a posição do personagem com base na tecla pressionada (WASD)
func personagemMover(tecla rune, jogo *Jogo) {
	dx, dy := 0, 0
	switch tecla {
	case 'w':
		dy = -1 // Move para cima
	case 'a':
		dx = -1 // Move para a esquerda
	case 's':
		dy = 1 // Move para baixo
	case 'd':
		dx = 1 // Move para a direita
	}

	nx, ny := jogo.PosX+dx, jogo.PosY+dy
	// Verifica se o movimento é permitido e realiza a movimentação
	if jogoPodeMoverPara(jogo, nx, ny) {
		jogoMoverElemento(jogo, jogo.PosX, jogo.PosY, dx, dy)
		jogo.PosX, jogo.PosY = nx, ny

		// === B) reportar posicao ao servidor
		if rpcClient != nil {
			payload := map[string]interface{}{"x": jogo.PosX, "y": jogo.PosY, "lives": jogo.Pontos}
			go rpcClient.SendCommand("UPDATE_POS", payload)
		}
	}
}

// Define o que ocorre quando o jogador pressiona a tecla de interação
// Neste exemplo, apenas exibe uma mensagem de status
// Você pode expandir essa função para incluir lógica de interação com objetos
func personagemInteragir(jogo *Jogo) {
	// Atualmente apenas exibe uma mensagem de status
	jogo.StatusMsg = fmt.Sprintf("Interagindo em (%d, %d)", jogo.PosX, jogo.PosY)
}

// Processa o evento do teclado e executa a ação correspondente
func personagemExecutarAcao(ev EventoTeclado, jogo *Jogo) bool {
	switch ev.Tipo {
	case "sair":
		// Retorna false para indicar que o jogo deve terminar
		return false
	case "interagir":
		// Executa a ação de interação
		personagemInteragir(jogo)
	case "mover":
		// Move o personagem com base na tecla
		personagemMover(ev.Tecla, jogo)
	}
	// TODO Member B: Logo após mover (no caso "mover"), enviar UPDATE_POS via rpcClient.
	// Se o rpcClient estiver disponível globalmente, chame aqui:
	//   payload := map[string]interface{}{ "x": jogo.PosX, "y": jogo.PosY, "lives": jogo.Pontos }
	//   rpcClient.SendCommand("UPDATE_POS", payload)

	return true // Continua o jogo
}
