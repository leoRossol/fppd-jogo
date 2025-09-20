// armadilha.go -> funcoes para as armadilhas do jogo
package main

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
