package util

import (
	"GameTest/glob"
	"math"
)

func Distance(xa, ya, xb, yb int) float64 {
	x := math.Abs(float64(xa - xb))
	y := math.Abs(float64(ya - yb))
	return math.Sqrt(x*x + y*y)
}

func MidPoint(x1, y1, x2, y2 int) (int, int) {
	return (x1 + x2) / 2, (y1 + y2) / 2
}

func GetObj(pos glob.Position, chunk *glob.MapChunk) *glob.MObj {
	obj := chunk.MObj[pos]

	return &obj
}

//Automatically converts position to chunk format
func GetChunk(pos glob.Position) glob.MapChunk {
	chunk := glob.WD.WorldMap[glob.Position{X: int(pos.X / glob.ChunkSize), Y: int(pos.Y / glob.ChunkSize)}]
	return chunk
}

func PosToChunkPos(pos glob.Position) glob.Position {
	return glob.Position{X: int(pos.X / glob.ChunkSize), Y: int(pos.Y / glob.ChunkSize)}
}

func FloatXYToPosition(x float64, y float64) glob.Position {

	return glob.Position{X: int(x), Y: int(y)}
}
