package util

import (
	"GameTest/consts"
	"GameTest/glob"
	"fmt"
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
	if chunk != nil {
		o := chunk.MObj[pos]
		return o
	} else {
		return nil
	}
}

//Automatically converts position to chunk format
func GetChunk(pos glob.Position) *glob.MapChunk {
	chunk := glob.WorldMap[glob.Position{X: int(pos.X / consts.ChunkSize), Y: int(pos.Y / consts.ChunkSize)}]
	return chunk
}

func PosToChunkPos(pos glob.Position) glob.Position {
	return glob.Position{X: int(pos.X / consts.ChunkSize), Y: int(pos.Y / consts.ChunkSize)}
}

func FloatXYToPosition(x float64, y float64) glob.Position {

	return glob.Position{X: int(x), Y: int(y)}
}

func GetNeighborObj(pos glob.Position, dir int) *glob.MObj {
	switch dir {
	case consts.DIR_NORTH:
		pos.Y++
	case consts.DIR_EAST:
		pos.X++
	case consts.DIR_SOUTH:
		pos.Y++
	case consts.DIR_WEST:
		pos.X--
	}

	fmt.Println("Finding neighbor:", pos, dir)

	chunk := GetChunk(pos)
	fmt.Println("chunk: ", chunk)
	return GetObj(pos, chunk)
}

func MoveMaterialToObj(src *glob.MObj, dest *glob.MObj, dir int) {

	if src == nil || dest == nil {
		return
	}
	for _, mat := range src.External[dir] {
		if mat == nil {
			continue
		}
		if dest.External[dir][mat.Type] == nil {
			dest.External[dir][mat.Type] = &glob.MatData{Type: mat.Type, TypeP: mat.TypeP, Amount: mat.Amount}
		} else {
			dest.External[dir][mat.Type].Amount += mat.Amount
		}

		src.External[dir][mat.Type].Amount = 0
	}

}

func MoveMaterialOut(obj *glob.MObj) {

	if obj == nil {
		return
	}
	for _, mat := range obj.Contains {
		if mat == nil {
			continue
		}
		if obj.External[obj.OutputDir][mat.Type] == nil {
			obj.External[obj.OutputDir][mat.Type] = &glob.MatData{Type: mat.Type, TypeP: mat.TypeP, Amount: mat.Amount}
		} else {
			obj.External[obj.OutputDir][mat.Type].Amount += mat.Amount
		}
		obj.Contains[mat.Type].Amount = 0

	}

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
	case consts.DIR_INTERNAL:
		return "Internal"
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
	case consts.DIR_INTERNAL:
		return consts.DIR_INTERNAL
	}

	return -1
}
