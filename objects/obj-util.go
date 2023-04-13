package objects

import (
	"GameTest/cwlog"
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
	"fmt"
	"math"
	"math/rand"
)

/* Delete object from ObjMap, ObjList, decerment NumObjects. Marks PixmapDirty */
func removeObj(obj *world.ObjData) {
	/* delete from map */
	obj.Parent.Lock.Lock()

	obj.Parent.NumObjs--
	delete(obj.Parent.BuildingMap, obj.Pos)

	obj.Parent.Parent.PixmapDirty = true

	obj.Parent.Lock.Unlock()
	util.ObjListDelete(obj)
}

func RotateCoord(coord world.XYs, dir uint8, size world.XYs) world.XYs {
	tempX := coord.X
	tempY := coord.Y

	if dir == 0 {
		return world.XYs{X: tempX, Y: tempY}
	} else if dir == 1 {
		return world.XYs{X: -tempY + (size.X - 1), Y: tempX}
	} else if dir == 2 {
		return world.XYs{X: -tempX + (size.X - 1), Y: -tempY + (size.Y - 1)}
	} else if dir == 3 {
		return world.XYs{X: tempY, Y: -tempX + (size.Y - 1)}
	} else {
		return world.XYs{X: 0, Y: 0}
	}
}

func RotatePosF64(coord world.XYs, dir uint8, size world.XYf64) world.XYf64 {
	tempX := float64(coord.X)
	tempY := float64(coord.Y)

	if dir == 0 {
		return world.XYf64{X: tempX, Y: tempY}
	} else if dir == 1 {
		return world.XYf64{X: -tempY + (size.Y - size.X), Y: tempX}
	} else if dir == 2 {
		return world.XYf64{X: -tempX, Y: -tempY + (size.Y - size.X)}
	} else if dir == 3 {
		return world.XYf64{X: tempY, Y: -tempX}
	} else {
		return world.XYf64{X: 0, Y: 0}
	}

}

func PrintUnit(mat *world.MatData) string {
	if mat != nil && mat.TypeP != nil {
		if world.ImperialUnits && mat.TypeP.UnitName == " kg" {
			buf := fmt.Sprintf("%0.2f lbs", mat.Amount*2.20462262185)
			return buf
		} else {
			buf := fmt.Sprintf("%0.2f%v", mat.Amount, mat.TypeP.UnitName)
			return buf
		}
	} else {
		return ""
	}
}

func PrintWeight(kg float32) string {

	if world.ImperialUnits {
		buf := fmt.Sprintf("%0.2f lbs", kg*2.20462262185)
		return buf
	} else {
		buf := fmt.Sprintf("%0.2f kg", kg)
		return buf
	}
}

func CalcVolume(mat *world.MatData) string {
	if mat != nil && mat.TypeP != nil {

		density := mat.TypeP.Density
		mass := mat.Amount
		cm3 := ((mass / density) * 1000.0)
		in3 := cm3 / 16.387064
		inSide := math.Sqrt(float64(in3))
		cmSide := math.Sqrt(float64(cm3))

		var buf string
		if world.ImperialUnits {
			buf = fmt.Sprintf("%0.2f x %0.2f in", inSide, inSide)
		} else {
			buf = fmt.Sprintf("%0.2f x %0.2f cm", cmSide, cmSide)
		}
		return buf
	} else {
		return ""
	}
}

/* Place and/or create a multi-tile object */
func PlaceObj(pos world.XY, mtype uint8, obj *world.ObjData, dir uint8, fast bool) *world.ObjData {

	/*
	 * Make chunk if needed.
	 * If not in "fast mode" then explore map area as well.
	 */
	if !fast {
		ExploreMap(pos, 6, fast)
	} else {
		MakeChunk(pos)
	}
	chunk := util.GetChunk(pos)
	g := util.GetObj(pos, chunk)

	var newObj *world.ObjData
	/* New object */
	if obj == nil {
		newObj = &world.ObjData{}
		newObj.Unique = &world.UniqueObject{TypeP: WorldObjs[mtype]}
	} else { /* Placing already existing object */
		newObj = obj
	}

	newObj.Pos = pos
	newObj.Parent = chunk
	newObj.Dir = dir

	multiTile := false
	subFits := false
	if newObj.Unique.TypeP.MultiTile {
		multiTile = true

		if SubObjFits(newObj, newObj.Unique.TypeP, true, pos) {
			subFits = true
		} else {
			return nil
		}
	} else {
		/* Obj already at this location */
		if g != nil {
			return nil
		}
	}

	initOkay := true
	if obj == nil {
		if newObj.Unique.TypeP.CanContain {
			newObj.Unique.Contents = &world.MaterialContentsType{}
			newObj.Unique.Contents.Mats = [gv.MAT_MAX]*world.MatData{}
		}

		if newObj.Unique.TypeP.MachineSettings.MaxFuelKG > 0 {
			newObj.Unique.KGFuel = newObj.Unique.TypeP.MachineSettings.MaxFuelKG
		}

		for p, port := range newObj.Unique.TypeP.Ports {
			newObj.Ports = append(newObj.Ports, port)
			newObj.Ports[p].Buf = &world.MatData{}
		}

		for p, port := range newObj.Ports {
			newObj.Ports[p].Dir = util.RotDir(dir, port.Dir)
		}

		/* Init obj if we have a function for it */
		if newObj.Unique.TypeP.InitObj != nil {
			if !newObj.Unique.TypeP.InitObj(newObj) {
				initOkay = false
			}
		}
	}

	newObj.Parent.Lock.Lock()
	/* Add to chunk object list */
	newObj.Parent.ObjList =
		append(newObj.Parent.ObjList, newObj)
	newObj.Parent.NumObjs++

	/* Mark superchunk and visdata dirty */
	newObj.Parent.Parent.PixmapDirty = true
	world.VisDataDirty.Store(true)
	newObj.Parent.Lock.Unlock()

	/* Add to tock/tick lists */

	/*Spread out when tock happens */
	if newObj.Unique.TypeP.TockInterval > 0 {
		newObj.TickCount = uint8(rand.Intn(int(newObj.Unique.TypeP.TockInterval)))
	}

	/* Place item tiles */
	if multiTile {
		if subFits {
			/* If space is available, create items */
			for _, sub := range newObj.Unique.TypeP.SubObjs {
				tile := RotateCoord(sub, dir, GetObjSize(newObj, nil))
				sXY := util.GetSubPos(pos, tile)
				MakeChunk(sXY)
				tchunk := util.GetChunk(sXY)
				if tchunk != nil {
					tchunk.Lock.Lock()
					newB := &world.BuildingData{Obj: newObj, Pos: sXY}
					util.ObjCD(newB, fmt.Sprintf("Created at: %v", util.PosToString(sXY)))
					tchunk.BuildingMap[sXY] = newB
					tchunk.Lock.Unlock()
					if initOkay {
						LinkObj(sXY, newB)
					}
				}
			}
			return newObj
		}
		return nil
	} else {
		/* Add object to map */
		newObj.Parent.Lock.Lock()
		newBB := &world.BuildingData{Obj: newObj, Pos: newObj.Pos}
		chunk.BuildingMap[newObj.Pos] = newBB
		newObj.Parent.Lock.Unlock()

		if initOkay {
			LinkObj(newObj.Pos, newBB)
		}
		return newObj
	}
}

func SubObjFits(obj *world.ObjData, TypeP *world.ObjType, report bool, pos world.XY) bool {

	size := GetObjSize(obj, TypeP)
	var dir uint8
	if obj != nil {
		dir = obj.Dir
	} else {
		dir = TypeP.Direction
	}
	/* Check if object fits */
	for _, sub := range TypeP.SubObjs {

		tile := RotateCoord(sub, dir, size)
		subPos := util.GetSubPos(pos, tile)
		tchunk := util.GetChunk(subPos)
		if tchunk != nil {
			if util.GetObj(subPos, tchunk) != nil {
				if report {
					util.Chat(
						fmt.Sprintf(
							"SubObjFits: (%v) Can't fit here: %v", TypeP.Name, util.PosToString(subPos),
						))
				}
				return false
			}
		}
	}

	return true
}

func GetObjSize(obj *world.ObjData, TypeP *world.ObjType) world.XYs {
	if obj != nil {
		if obj.Dir == 1 || obj.Dir == 3 {
			return world.XYs{X: obj.Unique.TypeP.Size.Y, Y: obj.Unique.TypeP.Size.X}
		} else {
			return obj.Unique.TypeP.Size
		}
	} else if TypeP != nil {
		if TypeP.Direction == 1 || TypeP.Direction == 3 {
			return world.XYs{X: TypeP.Size.Y, Y: TypeP.Size.X}
		} else {
			return TypeP.Size
		}
	} else {
		cwlog.DoLog(true, "GetObjSize: Obj and TypeP nil.")
		return world.XYs{X: 0, Y: 0}
	}
}

/* Quickly move material by swapping pointers */
/*
func swapPortBuf(px, py *world.MatData) {
	*px, *py = *py, *px
}
*/
