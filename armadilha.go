// armadilha.go -> funcoes para as armadilhas do jogo
package main

import (
	"math/rand"
	"time"
)

// representacao da armadilha
type Armadilha struct {
	X, Y  int  //posicao
	Ativa bool //se esta ativa ou nao
	ID    int  //identificador
}

// verifica se o player pisou na armadilha
func armadilhaAtivada(armadilha *Armadilha, jogo *Jogo) bool {
	return armadilha.Ativa && armadilha.X == jogo.PosX && armadilha.Y == jogo.PosY
}

// move todas as armadilhas para novas posições aleatórias
func moverTodasArmadilhas(armadilhas []*Armadilha, jogo *Jogo) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	for _, armadilha := range armadilhas {
		// For para achar um lufar para as
		for tentativas := 0; tentativas < 50; tentativas++ { // checkup para evitar problemas de posicionamento repetido
			nx := r.Intn(len(jogo.Mapa[0]))
			ny := r.Intn(len(jogo.Mapa))
			
			// Verifica se a posição é válida e o player nao esta nela
			if jogoPodeMoverPara(jogo, nx, ny) && (nx != jogo.PosX || ny != jogo.PosY) {
				// Verifica se não tem uma armadilha na posi atual
				posicaoLivre := true
				for _, outraArmadilha := range armadilhas {
					if outraArmadilha != armadilha && outraArmadilha.X == nx && outraArmadilha.Y == ny {
						posicaoLivre = false
						break
					}
				}
				
				if posicaoLivre {
					armadilha.X = nx
					armadilha.Y = ny
					break
				}
			}
		}
	}
}
