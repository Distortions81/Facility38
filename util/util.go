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

func MoveMaterialToObj(src *glob.MObj, dest *glob.MObj, dir int) {

	success := false
	dead := true

	mat := src.External[dir]
	if mat == nil {
		return
	}

	if dest != nil {
		if src.Valid && dest.Valid {
			dead = false

			if src.External[dir].Amount <= 0 {
				return
			}

			var destDir int
			if dest.OutputDir == consts.DIR_INTERNAL {
				destDir = consts.DIR_INTERNAL
			} else {
				destDir = dir
			}

			if dest.External[destDir] == nil {
				dest.External[destDir] = &glob.MatData{Type: mat.Type, TypeP: mat.TypeP, Amount: mat.Amount}
				success = true
			} else if dest.External[destDir].Type == mat.Type {
				dest.External[destDir].Amount += mat.Amount
				success = true
			} else {
				fmt.Println("Material mismatch: ", dest.External[destDir].Type, " != ", mat.Type)
			}

		}
	}

	if dead {
		src.SendTo[dir] = nil
		fmt.Println("Removed dead link.")
		return
	}

	if success {
		fmt.Println("Sent ", mat.Amount, mat.TypeP.Name, "to dest", DirToName(dir))
		src.External[dir].Amount = 0
	} else {
		fmt.Println("Failed to send ", mat.Amount, mat.TypeP.Name, "to dest", DirToName(dir))
	}

}

func MoveMaterialOut(obj *glob.MObj, dir int, mat *glob.MatData) {

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
		fmt.Println("Sent ", mat.Amount, mat.TypeP.Name, "to ext", DirToName(outDir))
		obj.Contents[dir].Amount = 0
	} else {
		fmt.Println("Failed to send ", mat.Amount, mat.TypeP.Name, "to ext", DirToName(outDir))
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
