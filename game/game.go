package game

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"

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

	// Sprites
	sandImg    *ebiten.Image
	blockImg   *ebiten.Image
	background *ebiten.Image

	// ticksSinceLastSand int

	// Game state
	snowing    bool
	mode       TileType
	gridWidth  int
	gridHeight int
	grid       Grid
	nextGrid   Grid
	gridScaleF float64 // The scaling factor between grid and screen
	gridScaleI int
	gravity    float64

	drawOpts ebiten.DrawImageOptions

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
		gravity:      1,
		mode:         Sand,
	}
}

func makeGrid(width, height int) Grid {
	grid := make([][]Tile, height)

	for idx := range grid {
		row := make([]Tile, width)
		// for idx := range row {
		// row[idx] = 1
		// }

		grid[idx] = row
	}

	return grid
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

	g.grid = makeGrid(g.gridWidth, g.gridHeight)
	g.nextGrid = makeGrid(g.gridWidth, g.gridHeight)

	width := g.gridWidth / 4
	for i := width; i < width*3; i++ {
		g.grid[g.gridHeight/2][i].Type = Wall
	}

	g.cursorImg = ebiten.NewImage(int(g.cursorWidth), int(g.cursorHeight))
	g.cursorImg.Fill(color.RGBA{255, 0, 0, 255})

	g.sandImg = ebiten.NewImage(g.gridScaleI, g.gridScaleI)
	g.sandImg.Fill(color.RGBA{255, 255, 255, 255})

	g.blockImg = ebiten.NewImage(g.gridScaleI, g.gridScaleI)
	g.blockImg.Fill(color.RGBA{0, 0, 255, 255})

	g.background = ebiten.NewImage(g.screenWidth, g.screenHeight)
	g.background.Fill(color.Black)
}

func (g *Game) resetGrid() {
	for i, row := range g.grid {
		for j := range row {
			g.grid[i][j] = Tile{
				Type: Blank,
			}
		}
	}

	width := g.gridWidth / 4
	for i := width; i < width*3; i++ {
		g.grid[g.gridHeight/2][i].Type = Wall
	}
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
		"Brush (1, 2, 3): %s\nGravity (g): %d\nSnow (s): %t\nReset: R\nFPS: %0.2f",
		g.mode, int(g.gravity), g.snowing, ebiten.CurrentFPS()))
}

func (g *Game) drawGrid(screen *ebiten.Image) {
	opts := ebiten.DrawImageOptions{}

	opts.GeoM.Translate(g.gridScaleF*0.5, g.gridScaleF*0.5)
	for iy, row := range g.grid {
		for ix, cell := range row {
			if cell.Type != Blank {
				// opts.Filter = ebiten.FilterLinear
				// opts.ColorM.Reset()
				// opts.ColorM.RotateHue((float64(g.ticks) + float64(ix+iy)) / 10)
				// str := fmt.Sprintf("x") //%d,%d", ix, iy)
				// bound, _ := font.BoundString(mplusSmallFont, str)
				// w := (bound.Max.X - bound.Min.X).Ceil()
				// h := (bound.Max.Y - bound.Min.Y).Ceil()
				// x := float64(ix) * g.gridScale
				// y := float64(iy) * g.gridScale
				// text.Draw(screen, str, mplusSmallFont, int(x), int(y), color.Black)
				opts.GeoM.Reset()
				opts.GeoM.Translate(math.Floor(float64(ix)*g.gridScaleF), math.Floor(float64(iy)*g.gridScaleF))
				img := g.sandImg
				if cell.Type == Wall {
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
	g.updateGrid()

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

	// Input handling

	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		os.Exit(1)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.resetGrid()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.gravity = -g.gravity
	}

	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.mode = Sand
	}

	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.mode = Wall
	}

	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.mode = Blank
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.snowing = !g.snowing
	}
}

func (g *Game) updateGrid() {
	g.nextGrid.Reset()

	// snow
	if g.snowing {
		for i := 1; i < 10; i++ {
			g.grid[0][random(g.gridWidth)].Type = Sand
		}
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
		g.grid[y][x].Type = g.mode
	}

	sy := 0
	ey := g.gridHeight
	dy := 1
	if g.gravity > 0 {
		sy = g.gridHeight - 1
		ey = -1
		dy = -1
	}

	for iy := sy; iy != ey; iy += dy {
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

			if cell.Type == Wall {
				g.nextGrid[iy][ix].Type = Wall
				continue
			}

			if cell.Type == Blank {
				continue
			}

			nextY := iy + int(g.gravity)

			if nextY == g.gridHeight || nextY < 0 {
				// We're blocked, stay in place
				g.nextGrid[iy][ix] = g.grid[iy][ix]
				continue
			}

			if g.nextGrid[nextY][ix].Type == Blank {
				// The cell directly below us is empty
				// g.nextGrid[iy][ix].Type = Blank
				g.nextGrid[nextY][ix].Type = Sand
				continue
			}

			if cointoss() {
				if ix > 0 && g.nextGrid[nextY][ix-1].Type == Blank {
					// fall off to the left
					// g.grid[iy][ix].Type = Blank
					g.nextGrid[nextY][ix-1].Type = Sand
					continue
				} else if ix+1 < g.gridWidth && g.nextGrid[nextY][ix+1].Type == Blank {
					// fall off to the right
					// g.grid[iy][ix].Type = Blank
					g.nextGrid[nextY][ix+1].Type = Sand
					continue
				}
			} else {
				if ix+1 < g.gridWidth && g.nextGrid[nextY][ix+1].Type == Blank {
					// fall off to the right
					// g.grid[iy][ix].Type = Blank
					g.nextGrid[nextY][ix+1].Type = Sand
					continue
				} else if ix > 0 && g.nextGrid[nextY][ix-1].Type == Blank {
					// fall off to the left
					// g.nextGrid[iy][ix].Type = Sand
					g.nextGrid[nextY][ix-1].Type = Sand
					continue
				}
			}

			// We were unable to move this sand, copy to the same position
			g.nextGrid[iy][ix] = g.grid[iy][ix]
		}
	}

	g.grid, g.nextGrid = g.nextGrid, g.grid

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
