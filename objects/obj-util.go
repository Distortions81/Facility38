package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
	"fmt"
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

/* Create a multi-tile object */
func CreateObj(pos world.XY, mtype uint8, dir uint8, fast bool) *world.ObjData {

	//Make chunk if needed
	if !fast {
		ExploreMap(pos, 6, fast)
	} else {
		MakeChunk(pos)
	}
	chunk := util.GetChunk(pos)
	g := util.GetObj(pos, chunk)

	/* Obj already at this location */
	if g != nil {
		return nil
	}

	newObj := &world.ObjData{}
	newObj.Pos = pos
	newObj.Parent = chunk

	newObj.TypeP = GameObjTypes[mtype]
	newObj.Dir = dir

	if newObj.TypeP.CanContain {
		newObj.Contents = [gv.MAT_MAX]*world.MatData{}
	}

	if newObj.TypeP.MaxFuelKG > 0 {
		newObj.KGFuel = newObj.TypeP.MaxFuelKG
	}

	for p, port := range newObj.TypeP.Ports {
		newObj.Ports = append(newObj.Ports, port)
		newObj.Ports[p].Buf = &world.MatData{}
	}

	for p, port := range newObj.Ports {
		newObj.Ports[p].Dir = util.RotDir(dir, port.Dir)
	}

	initOkay := true
	/* Init obj if we have a function for it */
	if newObj.TypeP.InitObj != nil {
		if !newObj.TypeP.InitObj(newObj) {
			initOkay = false
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
	if initOkay {
		/*Spread out when tock happens */
		if newObj.TypeP.Interval > 0 {
			newObj.TickCount = uint8(rand.Intn(int(newObj.TypeP.Interval)))
		}

		/* Add sub-objects to map */
		if newObj.TypeP.Size.X > 1 ||
			newObj.TypeP.Size.Y > 1 {

			/* Check if obj fits */
			if SubObjFits(newObj.TypeP, true, pos) {

				/* If space is available, create items */
				for _, sub := range newObj.TypeP.SubObjs {
					sXY := util.AddXY(sub, pos)
					MakeChunk(sXY)
					tchunk := util.GetChunk(sXY)
					if tchunk != nil {
						tchunk.Lock.Lock()
						newB := &world.BuildingData{Obj: newObj, Pos: sXY}
						tchunk.BuildingMap[sXY] = newB
						tchunk.Lock.Unlock()
						LinkObj(newB)
					}
				}
			} else {
				return nil
			}
		} else {
			/* Add object to map */
			newObj.Parent.Lock.Lock()
			newBB := &world.BuildingData{Obj: newObj, Pos: newObj.Pos}
			chunk.BuildingMap[newObj.Pos] = newBB
			newObj.Parent.Lock.Unlock()
			LinkObj(newBB)
		}
	}

	return newObj
}

func SubObjFits(sub *world.ObjType, report bool, pos world.XY) bool {

	/* Check if object fits */
	for _, tile := range sub.SubObjs {
		subPos := util.AddXY(pos, tile)
		tchunk := util.GetChunk(subPos)
		if tchunk != nil {
			if util.GetObj(subPos, tchunk) != nil {
				if report {
					util.Chat(
						fmt.Sprintf(
							"CreateObj: (%v) Can't fit here: %v", sub.Name, util.PosToString(subPos),
						))
				}
				return false
			}
		}
	}

	return true
}

/* Quickly move material by swapping pointers */
func swapPortBuf(px, py *world.MatData) {
	*px, *py = *py, *px
}
