package game

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

/*
var (
	mplusSmallFont font.Face
)

func init() {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	mplusSmallFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})

	if err != nil {
		log.Fatal(err)
	}
}
*/
type Game struct {
	initialized bool
	ticks       uint64

	sandImg            *ebiten.Image
	blockImg           *ebiten.Image
	ticksSinceLastSand int

	gridWidth  int
	gridHeight int
	grid       [][]int
	gridScaleF float64 // The scaling factor between grid and screen
	gridScaleI int

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
	cursorImg    *ebiten.Image

	screenWidth  int
	screenHeight int
}

func New() *Game {
	return &Game{
		cursorWidth:  10,
		cursorHeight: 10,
	}
}

func (g *Game) Init(w, h int) {
	if g.initialized {
		panic("double init")
	}

	g.initialized = true

	log.Println("Init")
	g.screenWidth = w
	g.screenHeight = h

	g.gridWidth = 150
	g.gridHeight = int(float32(g.gridWidth*g.screenHeight) / float32(g.screenWidth))

	g.gridScaleF = float64(g.screenWidth) / float64(g.gridWidth)
	g.gridScaleI = int(math.Ceil(g.gridScaleF))

	g.grid = make([][]int, g.gridHeight)

	for idx := range g.grid {
		row := make([]int, g.gridWidth)
		// for idx := range row {
		// row[idx] = 1
		// }

		g.grid[idx] = row
	}

	width := g.gridWidth / 4
	for i := width; i < width*3; i++ {
		g.grid[g.gridHeight/2][i] = 2
	}

	g.cursorImg = ebiten.NewImage(int(g.cursorWidth), int(g.cursorHeight))
	g.cursorImg.Fill(color.RGBA{255, 0, 0, 255})

	g.sandImg = ebiten.NewImage(g.gridScaleI, g.gridScaleI)
	g.sandImg.Fill(color.RGBA{0, 255, 0, 255})

	g.blockImg = ebiten.NewImage(g.gridScaleI, g.gridScaleI)
	g.blockImg.Fill(color.RGBA{0, 0, 255, 255})

	g.background = ebiten.NewImage(g.screenWidth, g.screenHeight)
	g.background.Fill(color.White)
}

//
// Rendering -----------------------------------------------------------------
//

// Draw satisfies Ebiten's Game interface
func (g *Game) Draw(screen *ebiten.Image) {
	g.drawOpts.GeoM.Reset()
	// g.drawOpts.ColorM.Reset()

	g.drawOpts.GeoM.Reset()

	g.drawBackground(screen)
	g.drawGrid(screen)
	g.drawCursor(screen)
	// g.drawSand(screen)

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"X: %d Y: %d Btn: %t\nFPS: %0.2f",
		g.cursorX, g.cursorY, g.buttonDown, ebiten.CurrentFPS()))
}

func (g *Game) drawGrid(screen *ebiten.Image) {
	opts := ebiten.DrawImageOptions{}

	opts.GeoM.Translate(g.gridScaleF*0.5, g.gridScaleF*0.5)
	for iy, row := range g.grid {
		for ix, cell := range row {
			if cell != 0 {
				// opts.Filter = ebiten.FilterLinear
				// opts.ColorM.Reset()
				opts.GeoM.Reset()
				// opts.ColorM.RotateHue((float64(g.ticks) + float64(ix+iy)) / 10)
				// str := fmt.Sprintf("x") //%d,%d", ix, iy)
				// bound, _ := font.BoundString(mplusSmallFont, str)
				// w := (bound.Max.X - bound.Min.X).Ceil()
				// h := (bound.Max.Y - bound.Min.Y).Ceil()
				// x := float64(ix) * g.gridScale
				// y := float64(iy) * g.gridScale
				// text.Draw(screen, str, mplusSmallFont, int(x), int(y), color.Black)
				opts.GeoM.Translate(math.Floor(float64(ix)*g.gridScaleF), math.Floor(float64(iy)*g.gridScaleF))
				img := g.sandImg
				if cell == 2 {
					img = g.blockImg
				}
				screen.DrawImage(img, &opts)
			}
		}
	}
}

func (g *Game) drawBackground(screen *ebiten.Image) {
	screen.DrawImage(g.background, &ebiten.DrawImageOptions{})
}

func (g *Game) drawCursor(screen *ebiten.Image) {
	x := math.Floor(float64(g.cursorX) / g.gridScaleF)
	y := math.Floor(float64(g.cursorY) / g.gridScaleF)
	x *= g.gridScaleF
	y *= g.gridScaleF

	opts := ebiten.DrawImageOptions{}

	opts.GeoM.Translate(x, y)

	screen.DrawImage(g.cursorImg, &opts)
}

//
// Game State ----------------------------------------------------------------
//

// Update satisfies Ebiten's Game interface
func (g *Game) Update() error {
	g.ticks += 1

	// Determine deltas for cursor movement
	g.updateCursor()
	g.updateSand()

	return nil
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

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		for i, row := range g.grid {
			for j, cell := range row {
				if cell == 1 {
					g.grid[i][j] = 0
				}
			}
		}
	}
}

func (g *Game) updateSand() {
	// snow
	for i := 1; i < 10; i++ {
		g.grid[0][random(g.gridWidth)] = 1

	}
	// g.ticksSinceLastSand++

	// if g.ticksSinceLastSand >= 3 {
	// 	log.Printf("tick")
	// 	g.ticksSinceLastSand = 0
	// } else {
	// 	return
	// }

	if g.buttonDown {
		x := int(math.Floor(float64(g.cursorX) / g.gridScaleF))
		y := int(math.Floor(float64(g.cursorY) / g.gridScaleF))
		g.grid[y][x] = 1
	}

	for iy := g.gridHeight - 2; iy >= 0; iy-- {
		sx := 0
		dx := 1
		ex := g.gridWidth
		if iy%2 == 0 {
			sx = g.gridWidth - 1
			ex = -1
			dx = -1
		}
		for ix := sx; ix != ex; ix += dx {
			cell := g.grid[iy][ix]
			if cell != 1 {
				continue
			}

			if cell == 0 {
				// Cell that is currently empty, or a cell into which we've already moved
				continue
			}

			if g.grid[iy+1][ix] == 0 {
				// The cell directly below us is empty
				g.grid[iy][ix] -= 1
				g.grid[iy+1][ix] = 1
				continue
			}

			if cointoss() {
				if ix > 0 && g.grid[iy+1][ix-1] == 0 {
					// fall off to the left
					g.grid[iy][ix] -= 1
					g.grid[iy+1][ix-1] = 1
				} else if ix+1 < g.gridWidth && g.grid[iy+1][ix+1] == 0 {
					// fall off to the right
					g.grid[iy][ix] -= 1
					g.grid[iy+1][ix+1] = 1
				}
			} else {
				if ix+1 < g.gridWidth && g.grid[iy+1][ix+1] == 0 {
					// fall off to the right
					g.grid[iy][ix] -= 1
					g.grid[iy+1][ix+1] = 1
				} else if ix > 0 && g.grid[iy+1][ix-1] == 0 {
					// fall off to the left
					g.grid[iy][ix] -= 1
					g.grid[iy+1][ix-1] = 1
				}
			}
		}
	}

	// for iy, row := range g.grid {
	// 	for ix, cell := range row {
	// 		if cell == 2 {
	// 			g.grid[iy][ix] = 1
	// 		}
	// 	}
	// }
	// g.printGrid()
}

func (g *Game) printGrid() {
	fmt.Println("")
	for _, row := range g.grid {
		for _, cell := range row {
			fmt.Print(cell)
			fmt.Print(" ")
		}
		fmt.Print("\n")
	}
}

func cointoss() bool {
	return rand.Intn(2) == 0
}
func random(max int) int {
	return rand.Intn(max)
}

// Layout satisfies Ebiten's Game interface
func (g *Game) Layout(outerWidth, outerHeight int) (int, int) {
	if !g.initialized {
		g.Init(outerWidth, outerHeight)
	}

	return g.screenWidth, g.screenHeight
}
