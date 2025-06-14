package main

import (
	"fmt"
	"math"

	"github.com/dustin/go-humanize"
)

/* Unlink and remove object */
func delObj(obj *ObjData) {
	defer reportPanic("delObj")

	if obj == nil {
		return
	}

	if obj.Unique.typeP.deInitObj != nil {
		obj.Unique.typeP.deInitObj(obj)
	}
	unlinkObj(obj)
	removeObj(obj)
}

/* Delete object from ObjMap, decrement Num Marks PixmapDirty */
func removePosMap(pos XY) {
	defer reportPanic("removePosMap")

	sChunk := GetSuperChunk(pos)
	chunk := GetChunk(pos)
	if chunk == nil || sChunk == nil {
		return
	}

	chunk.lock.Lock()
	chunk.numObjs--
	delete(chunk.buildingMap, pos)
	sChunk.pixmapDirty = true
	chunk.lock.Unlock()
}

/* Delete object from ObjMap, ObjList, decrement NumObj, set PixmapDirty */
func removeObj(obj *ObjData) {
	defer reportPanic("removeObj")

	/* delete from map */
	obj.chunk.lock.Lock()
	obj.chunk.numObjs--
	delete(obj.chunk.buildingMap, obj.Pos)
	obj.chunk.parent.pixmapDirty = true
	obj.chunk.lock.Unlock()

	ObjListDelete(obj)
}

/* Rotate object coords */
func rotateCoord(coord XYs, dir uint8, size XYs) XYs {
	defer reportPanic("RotateCoord")
	tempX := coord.X
	tempY := coord.Y

	if dir == 0 {
		return XYs{X: tempX, Y: tempY}
	} else if dir == 1 {
		return XYs{X: -tempY + (size.X - 1), Y: tempX}
	} else if dir == 2 {
		return XYs{X: -tempX + (size.X - 1), Y: -tempY + (size.Y - 1)}
	} else if dir == 3 {
		return XYs{X: tempY, Y: -tempX + (size.Y - 1)}
	} else {
		return XYs{X: 0, Y: 0}
	}
}

/* Rotate float64 coords */
func rotatePosF64(coord XYs, dir uint8, size XYf64) XYf64 {
	defer reportPanic("RotatePosF64")
	tempX := float64(coord.X)
	tempY := float64(coord.Y)

	if dir == 0 {
		return XYf64{X: tempX, Y: tempY}
	} else if dir == 1 {
		return XYf64{X: -tempY + (size.Y - size.X), Y: tempX}
	} else if dir == 2 {
		return XYf64{X: -tempX, Y: -tempY + (size.Y - size.X)}
	} else if dir == 3 {
		return XYf64{X: tempY, Y: -tempX}
	} else {
		return XYf64{X: 0, Y: 0}
	}

}

/* Print weight with units from material  */
func printUnit(mat *MatData) string {
	defer reportPanic("PrintUnit")
	if mat != nil && mat.typeP != nil {
		if usUnits && mat.typeP.unitName == " kg" {
			buf := fmt.Sprintf("%v lbs",
				humanize.SIWithDigits(float64(mat.Amount*2.20462262185), 2, ""))
			return buf
		} else {
			buf := fmt.Sprintf("%v%v",
				humanize.SIWithDigits(float64(mat.Amount), 2, ""), mat.typeP.unitName)
			return buf
		}
	} else {
		return ""
	}
}

/* Print weight with units from float32 */
/* This could be consolidated with PrintUnit */
func printWeight(kg float32) string {
	defer reportPanic("PrintWeight")
	if usUnits {
		buf := fmt.Sprintf("%0.2f lbs", kg*2.20462262185)
		return buf
	} else {
		buf := fmt.Sprintf("%0.2f kg", kg)
		return buf
	}
}

/* Based on object weight and density, calculate cubic volume */
func calcVolume(mat *MatData) string {
	defer reportPanic("CalcVolume")
	if mat != nil && mat.typeP != nil {

		density := mat.typeP.density
		mass := mat.Amount
		cm3 := ((mass / density) * 1000.0)
		in3 := cm3 / 16.387064
		inSide := math.Sqrt(float64(in3))
		cmSide := math.Sqrt(float64(cm3))

		var buf string
		if usUnits {
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
func placeObj(pos XY, mType uint8, obj *ObjData, dir uint8, fast bool) *ObjData {
	defer reportPanic("PlaceObj")
	/*
	 * Make chunk if needed.
	 * If not in "fast mode" then explore map area as well.
	 */
	if !fast {
		exploreMap(pos, 6, false)
	} else {
		makeChunk(pos)
	}
	chunk := GetChunk(pos)
	g := GetObj(pos, chunk)

	var newObj *ObjData
	/* New object */
	if obj == nil {
		newObj = &ObjData{}
		newObj.Unique = &UniqueObjectData{typeP: worldObjs[mType]}
	} else { /* Placing already existing object */
		newObj = obj
	}

	newObj.Pos = pos
	newObj.chunk = chunk
	newObj.Dir = dir

	multiTile := false
	subFits := false
	if newObj.Unique.typeP.multiTile {
		multiTile = true

		if subObjFits(newObj, newObj.Unique.typeP, true, pos) {
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
		if newObj.Unique.typeP.canContain {
			newObj.Unique.Contents = &materialContentsTypeData{}
			newObj.Unique.Contents.mats = [MAT_MAX]*MatData{}
		}

		if newObj.Unique.typeP.machineSettings.maxFuelKG > 0 {
			newObj.Unique.KGFuel = newObj.Unique.typeP.machineSettings.maxFuelKG
		}

		for p, port := range newObj.Unique.typeP.ports {
			newObj.Ports = append(newObj.Ports, port)
			newObj.Ports[p].Buf = &MatData{}
			newObj.Ports[p].BufNext = &MatData{}
		}

		for p, port := range newObj.Ports {
			newObj.Ports[p].Dir = RotDir(dir, port.Dir)
		}

		/* Init obj if we have a function for it */
		if newObj.Unique.typeP.initObj != nil {
			if !newObj.Unique.typeP.initObj(newObj) {
				initOkay = false
			}
		}
	}

	newObj.chunk.lock.Lock()
	/* Add to chunk object list */
	newObj.chunk.objList =
		append(newObj.chunk.objList, newObj)
	newObj.chunk.numObjs++

	/* Mark superChunk and visData dirty */
	newObj.chunk.parent.pixmapDirty = true
	visDataDirty.Store(true)
	newObj.chunk.lock.Unlock()

	/* Place item tiles */
	if multiTile {
		if subFits {
			/* If space is available, create items */
			for _, sub := range newObj.Unique.typeP.subObjs {
				tile := rotateCoord(sub, dir, getObjSize(newObj, nil))
				sXY := GetSubPos(pos, tile)
				makeChunk(sXY)
				tchunk := GetChunk(sXY)
				if tchunk != nil {
					tchunk.lock.Lock()
					newB := &buildingData{obj: newObj, pos: sXY}
					objCD(newB, fmt.Sprintf("Created at: %v", posToString(sXY)))
					tchunk.buildingMap[sXY] = newB
					tchunk.lock.Unlock()
					if initOkay {
						linkObj(sXY, newB)
					}
				}
			}
			return newObj
		}
		return nil
	} else {
		/* Add object to map */
		newObj.chunk.lock.Lock()
		newBB := &buildingData{obj: newObj, pos: newObj.Pos}
		chunk.buildingMap[newObj.Pos] = newBB
		newObj.chunk.lock.Unlock()

		if initOkay {
			linkObj(newObj.Pos, newBB)
		}
		return newObj
	}
}

/* Check if a rotated multi-tile object will fit */
func subObjFits(obj *ObjData, TypeP *objTypeData, report bool, pos XY) bool {
	defer reportPanic("SubObjFits")
	size := getObjSize(obj, TypeP)
	var dir uint8
	if obj != nil {
		dir = obj.Dir
	} else {
		dir = TypeP.direction
	}
	/* Check if object fits */
	for _, sub := range TypeP.subObjs {

		tile := rotateCoord(sub, dir, size)
		subPos := GetSubPos(pos, tile)
		tchunk := GetChunk(subPos)
		if tchunk != nil {
			if GetObj(subPos, tchunk) != nil {
				if report {
					chat(
						fmt.Sprintf(
							"SubObjFits: (%v) Can't fit here: %v", TypeP.name, posToString(subPos),
						))
				}
				return false
			}
		}
	}

	return true
}

/* Return object size */
func getObjSize(obj *ObjData, TypeP *objTypeData) XYs {
	defer reportPanic("GetObjSize")
	if obj != nil {
		if obj.Dir == 1 || obj.Dir == 3 {
			return XYs{X: obj.Unique.typeP.size.Y, Y: obj.Unique.typeP.size.X}
		} else {
			return obj.Unique.typeP.size
		}
	} else if TypeP != nil {
		if TypeP.direction == 1 || TypeP.direction == 3 {
			return XYs{X: TypeP.size.Y, Y: TypeP.size.X}
		} else {
			return TypeP.size
		}
	} else {
		doLog(true, "GetObjSize: Obj and TypeP nil.")
		return XYs{X: 0, Y: 0}
	}
}

/* Quickly move material by swapping pointers */
/*
func swapPortBuf(px, py *MatData) {
	*px, *py = *py, *px
}
*/
