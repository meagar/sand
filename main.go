package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	drawOpts ebiten.DrawImageOptions

	cursorX    int
	cursorY    int
	buttonDown bool
}

var in *ebiten.Image
var out *ebiten.Image
var mouseIn bool

var screenWidth int
var screenHeight int

type Sand struct {
	x int
	y int
}

func (g *Game) Init() {
	in = ebiten.NewImage(screenWidth, screenHeight)
	in.Fill(color.Gray{90})

	out = ebiten.NewImage(screenWidth, screenHeight)
	out.Fill(color.Gray{50})
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawOpts.GeoM.Reset()
	g.drawOpts.ColorM.Reset()

	if in == nil {
		g.Init()
	}

	if mouseIn {
		screen.DrawImage(in, &g.drawOpts)
	} else {
		screen.DrawImage(out, &g.drawOpts)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("X: %d Y: %d Btn: %t", g.cursorX, g.cursorY, g.buttonDown))
}

func (g *Game) Update() error {
	g.cursorX, g.cursorY = ebiten.CursorPosition()

	mouseIn = g.cursorX > 0 && g.cursorX < screenWidth && g.cursorY > 0 && g.cursorY < screenHeight
	g.buttonDown = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	return nil
}

func (*Game) Layout(outerWidth, outerHeight int) (int, int) {
	screenWidth = outerWidth
	screenHeight = outerHeight

	return outerWidth, outerHeight
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Life")

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
