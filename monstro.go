// monstro.go -> funcoes para a movimentacao e etc do monstro
package main

// estrutura que representa o monstro no jogo
type Monstro struct {
	X, Y int //posicao do bicho
}

// move o monstro em direcao ao player
func monstroMover(monstro *Monstro, alvoX, alvoY int, jogo *Jogo) {

	dx, dy := 0, 0
	if monstro.X < alvoX {
		dx = 1
	} else if monstro.X > alvoX {
		dx = -1
	}
	if monstro.Y < alvoY {
		dy = 1
	} else if monstro.Y > alvoY {
		dy = -1
	}
	monstro.X += dx
	monstro.Y += dy
}

// verifica se o monstro encostou no player
func monstroEncostou(monstro *Monstro, jogo *Jogo) bool {
	return monstro.X == jogo.PosX && monstro.Y == jogo.PosY
}
