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
	o := chunk.MObj[pos]

	return o
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
		pos.Y--
	case consts.DIR_EAST:
		pos.X++
	case consts.DIR_SOUTH:
		pos.Y++
	case consts.DIR_WEST:
		pos.X--
	}

	chunk := GetChunk(PosToChunkPos(pos))
	obj := GetObj(pos, chunk)

	return obj
}

func MoveMaterialExt(src *glob.MObj, dest *glob.MObj, dir int, mat *glob.MatData) {

	success := false
	dead := true

	if dest != nil && mat != nil {
		if src.Valid && dest.Valid {
			dead = false
			if dest.External[dir] == nil {
				dest.External[dir] = &glob.MatData{Type: mat.Type, TypeP: mat.TypeP, Amount: mat.Amount}
				success = true
			} else if dest.External[dir].Type == mat.Type {
				dest.External[dir].Amount += mat.Amount
				success = true
			} else {
				fmt.Println("Material mismatch")
			}
		}
	}

	if dead {
		src.SendTo[dir] = nil
		fmt.Println("Removed dead link.")
	}

	if success {
		src.External[dir].Amount = 0
	}

}

func MoveMaterialInt(obj *glob.MObj, dir int, mat *glob.MatData) {

	success := false

	var outDir int

	if obj.OutputDir != consts.DIR_INTERNAL {
		outDir = obj.OutputDir
	} else {
		outDir = dir
	}

	if obj != nil && mat != nil {
		if obj.External[outDir] == nil {
			obj.External[outDir] = &glob.MatData{Type: mat.Type, TypeP: mat.TypeP, Amount: mat.Amount}
			success = true
		} else if obj.External[outDir].Type == mat.Type {
			obj.External[outDir].Amount += mat.Amount
			success = true
		}

	}

	if success {
		fmt.Println("Sent ", mat.Amount, mat.TypeP.Name, "to ext", outDir)
		obj.Contents[dir].Amount = 0
	}

}
