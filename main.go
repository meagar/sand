package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/meagar/sand/game"
)

func main() {
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Life")
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)

	g := game.New()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
