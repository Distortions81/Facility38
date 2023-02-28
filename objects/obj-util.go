package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
)

/* Delete object from ObjMap, ObjList, decerment NumObjects. Marks PixmapDirty */
func removeObj(obj *world.ObjData) {
	/* delete from map */
	obj.Parent.Lock.Lock()
	defer obj.Parent.Lock.Unlock()

	obj.Parent.NumObjs--
	delete(obj.Parent.BuildingMap, obj.Pos)
	util.ObjListDelete(obj)

	obj.Parent.Parent.PixmapDirty = true
}

/* Create a multi-tile object */
func CreateObj(pos world.XY, mtype uint8, dir uint8) *world.ObjData {

	//Make chunk if needed
	if MakeChunk(pos) {
		ExploreMap(pos, 4)
	}
	chunk := util.GetChunk(pos)
	b := util.GetObj(pos, chunk)

	/* Obj already at this location */
	if b != nil {
		return nil
	}

	world.VisDataDirty.Store(true)

	b = &world.BuildingData{}
	b.Obj = &world.ObjData{}
	b.Obj.Pos = pos
	b.Obj.Parent = chunk

	b.Obj.TypeP = GameObjTypes[mtype]
	b.Obj.Dir = dir

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
		b.Obj.Ports[p].Dir = util.RotDir(b.Obj.TypeP.Direction, port.Dir)
	}

	LinkObj(b)

	/* We should add/remove this based on object links */
	/* Only add to list if the object calls an update function */
	if b.Obj.TypeP.UpdateObj != nil {
		EventQueueAdd(b.Obj, gv.QUEUE_TYPE_TOCK, false)
	}

	EventQueueAdd(b.Obj, gv.QUEUE_TYPE_TICK, false)

	/* Init obj if we have a function for it */
	if b.Obj.TypeP.InitObj != nil {
		b.Obj.TypeP.InitObj(b.Obj)
	}

	b.Obj.Parent.Lock.Lock()

	/* Add sub-objects to map */
	if b.Obj.TypeP.Size.X > 1 ||
		b.Obj.TypeP.Size.Y > 1 {
		for _, sub := range b.Obj.TypeP.SubObjs {
			sXY := util.AddXY(sub, b.Obj.Pos)
			b.Obj.Parent.BuildingMap[sXY] = b
		}
	} else {
		/* Add object to map */
		b.Obj.Parent.BuildingMap[pos] = b
	}

	b.Obj.Parent.ObjList =
		append(b.Obj.Parent.ObjList, b.Obj)
	b.Obj.Parent.Parent.PixmapDirty = true
	b.Obj.Parent.NumObjs++
	b.Obj.Parent.Lock.Unlock()

	return b.Obj
}

func swapPortBuf(px, py *world.MatData) {
	*px, *py = *py, *px
}
