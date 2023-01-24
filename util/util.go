package util

import (
	"GameTest/consts"
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

func GetObj(pos *glob.Position, chunk *glob.MapChunk) *glob.WObject {
	if chunk != nil {
		o := chunk.WObject[*pos]
		return o
	} else {
		return nil
	}
}

// Automatically converts position to chunk format
func GetChunk(pos *glob.Position) *glob.MapChunk {
	glob.WorldMapLock.Lock()
	chunk := glob.WorldMap[glob.Position{X: pos.X / consts.ChunkSize, Y: pos.Y / consts.ChunkSize}]
	glob.WorldMapLock.Unlock()
	return chunk
}

func PosToChunkPos(pos *glob.Position) glob.Position {
	return glob.Position{X: pos.X / consts.ChunkSize, Y: pos.Y / consts.ChunkSize}
}

func FloatXYToPosition(x float64, y float64) glob.Position {

	return glob.Position{X: int(x), Y: int(y)}
}

func GetNeighborObj(src *glob.WObject, pos glob.Position, dir int) *glob.WObject {

	switch dir {
	case consts.DIR_NORTH:
		pos.Y--
	case consts.DIR_EAST:
		pos.X++
	case consts.DIR_SOUTH:
		pos.Y++
	case consts.DIR_WEST:
		pos.X--
	}

	////fmt.Println("Finding neighbor:", pos, DirToName(dir))

	chunk := GetChunk(&pos)
	obj := GetObj(&pos, chunk)
	return obj
}

func DirToName(dir int) string {
	switch dir {
	case consts.DIR_NORTH:
		return "North"
	case consts.DIR_EAST:
		return "East"
	case consts.DIR_SOUTH:
		return "South"
	case consts.DIR_WEST:
		return "West"
	}

	return "Error"
}

func ReverseDirection(dir int) int {
	switch dir {
	case consts.DIR_NORTH:
		return consts.DIR_SOUTH
	case consts.DIR_EAST:
		return consts.DIR_WEST
	case consts.DIR_SOUTH:
		return consts.DIR_NORTH
	case consts.DIR_WEST:
		return consts.DIR_EAST
	}

	return consts.DIR_NONE
}
