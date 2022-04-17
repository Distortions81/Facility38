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

func GetObj(pos *glob.Position, chunk *glob.MapChunk) *glob.MObj {
	if chunk != nil {
		o := chunk.MObj[*pos]
		return o
	} else {
		return nil
	}
}

//Automatically converts position to chunk format
func GetChunk(pos *glob.Position) *glob.MapChunk {
	chunk := glob.WorldMap[glob.Position{X: pos.X / consts.ChunkSize, Y: pos.Y / consts.ChunkSize}]
	return chunk
}

func PosToChunkPos(pos *glob.Position) glob.Position {
	return glob.Position{X: pos.X / consts.ChunkSize, Y: pos.Y / consts.ChunkSize}
}

func FloatXYToPosition(x float64, y float64) glob.Position {

	return glob.Position{X: int(x), Y: int(y)}
}

func GetNeighborObj(src *glob.MObj, pos glob.Position, dir int) *glob.MObj {

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

	//fmt.Println("Finding neighbor:", pos, dir)

	chunk := GetChunk(&pos)
	obj := GetObj(&pos, chunk)

	if obj == src {
		fmt.Println("GetNeighborObj: Self reference")
	}

	if chunk != nil && obj != nil {
		fmt.Println("Neighbor:", obj.TypeP.Name, pos)
	}
	return obj
}

func OutputMaterial(src *glob.MObj) {

	if src == nil || src.OutputObj == nil ||
		!src.Valid || !src.OutputObj.Valid {

		fmt.Println("OutputMaterial: Invalid source or output object")
		if src != nil && src.Valid {
			src.OutputObj = nil
		}
		return
	}

	dest := src.OutputObj
	if dest.InputBuffer == nil {
		dest.InputBuffer = append(dest.InputBuffer, *src.OutputBuffer)
	}
	src.OutputBuffer = nil
}

//Just move the pointer, don't copy or add.
//We only do this internally in the object so we don't do this operation twice.
//These are dedicated buffers for multithreading
func MoveMaterialOut(obj *glob.MObj) {

	if obj == nil || !obj.Valid {
		fmt.Println("MoveMaterialOut: Invalid object")
		return
	}

	if obj.OutputObj == nil {
		obj.OutputBuffer = obj.Contains
	}
	obj.Contains = nil

}

func MoveMaterialsIn(obj *glob.MObj) {
	if obj == nil || !obj.Valid {
		fmt.Println("MoveMaterialIn: Invalid object")
		return
	}

	if obj.TypeP.CapacityKG < obj.KGHeld {
		if obj.Contains == nil {
			obj.Contains = &[consts.MAT_MAX]*glob.MatData{}
		}

		for _, mats := range obj.InputBuffer {
			for mtype, mat := range mats {

				obj.Contains[mtype].Amount += mat.Amount
				obj.Contains[mtype].TypeP = mat.TypeP
				obj.Contains[mtype].Obj = mat.Obj
			}
		}

		obj.InputBuffer = nil
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
	case consts.DIR_NONE:
		return "None"
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
	case consts.DIR_NONE:
		return consts.DIR_NONE
	}

	return consts.DIR_NONE
}
