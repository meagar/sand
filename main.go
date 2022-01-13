package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	initialized bool

	drawOpts ebiten.DrawImageOptions

	background *ebiten.Image

	// Mouse state/graphics
	cursorAbsX   int
	cursorAbsY   int
	cursorX      int
	cursorY      int
	cursorWidth  int
	cursorHeight int
	buttonDown   bool
	cursor       *ebiten.Image

	screenWidth  int
	screenHeight int
}

type Sand struct {
	x int
	y int
}

func (g *Game) Init(w, h int) {
	g.screenWidth = w
	g.screenHeight = h

	g.cursor = ebiten.NewImage(int(g.cursorWidth), int(g.cursorHeight))
	g.cursor.Fill(color.RGBA{255, 0, 0, 255})

	g.background = ebiten.NewImage(g.screenWidth, g.screenHeight)
	g.background.Fill(color.White)
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawOpts.GeoM.Reset()
	g.drawOpts.ColorM.Reset()

	g.drawOpts.GeoM.Reset()

	g.drawBackground(screen)
	g.drawCursor(screen)

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"X: %d Y: %d Btn: %t\nFPS: %0.2f",
		g.cursorX, g.cursorY, g.buttonDown, ebiten.CurrentFPS()))
}

func (g *Game) drawBackground(screen *ebiten.Image) {
	screen.DrawImage(g.background, &ebiten.DrawImageOptions{})
}

func (g *Game) drawCursor(screen *ebiten.Image) {
	g.drawOpts.GeoM.Translate(float64(g.cursorX-(g.cursorWidth/2)), float64(g.cursorY-(g.cursorHeight/2)))
	screen.DrawImage(g.cursor, &g.drawOpts)

}

func (g *Game) updateCursor() {
	x, y := ebiten.CursorPosition()
	dx, dy := g.cursorAbsX-x, g.cursorAbsY-y
	g.cursorAbsX, g.cursorAbsY = x, y

	g.cursorX -= dx
	g.cursorY -= dy

	if g.cursorX-g.cursorWidth < 0 {
		g.cursorX = g.cursorWidth
	} else if g.cursorX+g.cursorWidth > g.screenWidth {
		g.cursorX = g.screenWidth - g.cursorWidth
	}

	if g.cursorY-g.cursorHeight < 0 {
		g.cursorY = g.cursorHeight
	} else if g.cursorY+g.cursorHeight > g.screenHeight {
		g.cursorY = g.screenHeight - g.cursorHeight
	}

	g.buttonDown = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

func (g *Game) Update() error {
	// Determine deltas for cursor movement
	g.updateCursor()

	return nil
}

func (g *Game) Layout(outerWidth, outerHeight int) (int, int) {
	outerWidth, outerHeight = 640, 480
	if !g.initialized {
		g.Init(outerWidth, outerHeight)
	}

	return outerWidth, outerHeight
}

func main() {
	ebiten.SetWindowSize(1024, 768)
	ebiten.SetWindowTitle("Life")
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)

	g := Game{
		cursorWidth:  10,
		cursorHeight: 10,
	}

	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}
