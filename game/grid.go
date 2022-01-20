package game

import "log"

type Tile struct {
	Type TileType
	Dx   int8
	Dy   int8
}

type TileType int

const Blank TileType = 0
const Wall TileType = 1
const Sand TileType = 2

type Grid [][]Tile

func (t TileType) String() string {
	switch t {
	case Blank:
		return "blank"
	case Wall:
		return "wall"
	case Sand:
		return "sand"
	}

	log.Fatalf("Invalid tile type: %d", t)
	return ""
}

func (g *Grid) Reset() {
	for i := 0; i < len(*g); i++ {
		for j := 0; j < len((*g)[0]); j++ {
			(*g)[i][j] = Tile{}
		}
	}
}
