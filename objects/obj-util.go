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
	ExploreMap(pos, 6, false)
	MakeChunk(pos)
	chunk := util.GetChunk(pos)
	b := util.GetObj(pos, chunk)

	/* Obj already at this location */
	if b != nil {
		return nil
	}

	b = &world.BuildingData{}
	b.Obj = &world.ObjData{}
	b.Obj.Pos = pos
	b.Obj.Parent = chunk

	b.Obj.TypeP = GameObjTypes[mtype]
	b.Obj.Dir = dir

	/* Add sub-objects to map */
	if b.Obj.TypeP.Size.X > 1 ||
		b.Obj.TypeP.Size.Y > 1 {

		/* Check if obj fits */
		if SubObjFits(b.Obj.TypeP, true, pos) {

			/* If space is available, create items */
			for _, sub := range b.Obj.TypeP.SubObjs {
				sXY := util.AddXY(sub, pos)
				MakeChunk(sXY)
				tchunk := util.GetChunk(sXY)
				if tchunk != nil {
					tchunk.Lock.Lock()
					tchunk.BuildingMap[sXY] = b
					tchunk.BuildingMap[sXY].Pos = sXY
					tchunk.Lock.Unlock()
				}
			}
		} else {
			return nil
		}
	} else {
		/* Add object to map */
		b.Obj.Parent.Lock.Lock()
		b.Obj.Parent.BuildingMap[pos] = b
		b.Obj.Parent.BuildingMap[pos].Pos = pos
		b.Obj.Parent.Lock.Unlock()
	}

	if b.Obj.TypeP.CanContain {
		b.Obj.Contents = [gv.MAT_MAX]*world.MatData{}
	}

	if b.Obj.TypeP.MaxFuelKG > 0 {
		b.Obj.KGFuel = b.Obj.TypeP.MaxFuelKG
	}

	for p, port := range b.Obj.TypeP.Ports {
		b.Obj.Ports = append(b.Obj.Ports, port)
		b.Obj.Ports[p].Buf = &world.MatData{}
	}

	for p, port := range b.Obj.Ports {
		b.Obj.Ports[p].Dir = util.RotDir(dir, port.Dir)
	}

	initOkay := true
	/* Init obj if we have a function for it */
	if b.Obj.TypeP.InitObj != nil {
		if !b.Obj.TypeP.InitObj(b.Obj) {
			initOkay = false
		}
	}

	b.Obj.Parent.Lock.Lock()
	/* Add to chunk object list */
	b.Obj.Parent.ObjList =
		append(b.Obj.Parent.ObjList, b.Obj)
	b.Obj.Parent.NumObjs++

	/* Mark superchunk and visdata dirty */
	b.Obj.Parent.Parent.PixmapDirty = true
	world.VisDataDirty.Store(true)
	b.Obj.Parent.Lock.Unlock()

	/* Add to tock/tick lists */
	if initOkay {
		LinkObj(b)
		/*Spread out when tock happens */
		if b.Obj.TypeP.Interval > 0 {
			b.Obj.TickCount = uint8(rand.Intn(int(b.Obj.TypeP.Interval)))
		}
	}

	return b.Obj
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
