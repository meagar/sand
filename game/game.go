package game

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

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

type Game struct {
	initialized bool

	sand               []Sand
	sandImg            *ebiten.Image
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
	if g.initialized == true {
		panic("double init")
	}
	g.initialized = true

	log.Println("Init")
	g.screenWidth = w
	g.screenHeight = h

	g.gridWidth = 200
	g.gridHeight = int(float32(g.gridWidth*g.screenHeight) / float32(g.screenWidth))

	g.gridScaleF = float64(g.screenWidth) / float64(g.gridWidth)
	g.gridScaleI = int(math.Ceil(g.gridScaleF))

	g.grid = make([][]int, g.gridHeight)

	for idx := range g.grid {
		row := make([]int, g.gridWidth)
		/*for idx := range row {
			r := rand.Int()
			fmt.Println(r)
			if r > 8810822825046566661 {
				row[idx] = 1
			}
		}*/

		g.grid[idx] = row
	}

	g.cursorImg = ebiten.NewImage(int(g.cursorWidth), int(g.cursorHeight))
	g.cursorImg.Fill(color.RGBA{255, 0, 0, 255})

	g.sandImg = ebiten.NewImage(g.gridScaleI, g.gridScaleI)
	g.sandImg.Fill(color.RGBA{0, 255, 0, 255})

	g.background = ebiten.NewImage(g.screenWidth, g.screenHeight)
	g.background.Fill(color.White)

}

//
// Rendering -----------------------------------------------------------------
//

// Draw satisfies Ebiten's Game interface
func (g *Game) Draw(screen *ebiten.Image) {
	g.drawOpts.GeoM.Reset()
	g.drawOpts.ColorM.Reset()

	g.drawOpts.GeoM.Reset()

	g.drawBackground(screen)
	g.drawGrid(screen)
	g.drawCursor(screen)
	// g.drawSand(screen)

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"X: %d Y: %d Btn: %t\nFPS: %0.2f\nSand: %d",
		g.cursorX, g.cursorY, g.buttonDown, ebiten.CurrentFPS(), len(g.sand)))
}

func (g *Game) drawGrid(screen *ebiten.Image) {
	opts := ebiten.DrawImageOptions{}
	opts.GeoM.Translate(g.gridScaleF*0.5, g.gridScaleF*0.5)
	for iy, row := range g.grid {
		for ix, cell := range row {
			if cell != 0 {
				opts.Filter = ebiten.FilterLinear
				opts.GeoM.Reset()
				// str := fmt.Sprintf("x") //%d,%d", ix, iy)
				// bound, _ := font.BoundString(mplusSmallFont, str)
				// w := (bound.Max.X - bound.Min.X).Ceil()
				// h := (bound.Max.Y - bound.Min.Y).Ceil()
				// x := float64(ix) * g.gridScale
				// y := float64(iy) * g.gridScale
				// text.Draw(screen, str, mplusSmallFont, int(x), int(y), color.Black)
				opts.GeoM.Translate(math.Floor(float64(ix)*g.gridScaleF), math.Floor(float64(iy)*g.gridScaleF))
				img := g.sandImg
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

func (g *Game) drawSand(screen *ebiten.Image) {
	ops := &ebiten.DrawImageOptions{}
	for _, s := range g.sand {
		ops.GeoM.Translate(float64(s.x), float64(s.y))
		screen.DrawImage(g.sandImg, ops)
		ops.GeoM.Reset()
	}
}

//
// Game State ----------------------------------------------------------------
//

// Update satisfies Ebiten's Game interface
func (g *Game) Update() error {
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
}

func (g *Game) updateSand() {
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

	for iy, row := range g.grid {
		for ix, cell := range row {
			if cell == 0 || cell == 2 {
				// Cell that is currently empty, or a cell into which we've already moved
				continue
			}

			if iy+1 >= g.gridHeight {
				// Bottom of the grid
				continue
			}

			// if cell != 1 && cell != 3 {
			// 	log.Fatalf("cell has value of %d", cell)
			// }

			// We're sand.

			if g.grid[iy+1][ix] == 0 {
				// The cell directly below us is empty
				g.grid[iy][ix] -= 1
				g.grid[iy+1][ix] += 2
			} else if ix > 0 && g.grid[iy+1][ix-1] == 0 {
				// fall off to the left
				g.grid[iy][ix] -= 1
				g.grid[iy+1][ix-1] += 2
			} else if ix+1 < g.gridWidth && g.grid[iy+1][ix+1] == 0 {
				// fall off to the right
				g.grid[iy][ix] -= 1
				g.grid[iy+1][ix+1] += 2
			}
		}
	}

	for iy, row := range g.grid {
		for ix, cell := range row {
			if cell == 2 {
				g.grid[iy][ix] = 1
			}
		}
	}
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

// Layout satisfies Ebiten's Game interface
func (g *Game) Layout(outerWidth, outerHeight int) (int, int) {
	if !g.initialized {
		g.Init(outerWidth, outerHeight)
	}

	return g.screenWidth, g.screenHeight
}
