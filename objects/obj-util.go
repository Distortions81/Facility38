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

	/* Obj already at his location */
	if b != nil {
		return nil
	}

	world.VisDataDirty.Store(true)

	b = &world.BuildingData{}
	b.Obj = &world.ObjData{}
	b.Obj.Pos = pos
	b.Obj.Parent = chunk

	b.Obj.TypeP = GameObjTypes[mtype]

	b.Obj.Parent.Lock.Lock()

	b.Obj.Parent.BuildingMap[pos] = &world.BuildingData{}
	b.Obj.Parent.BuildingMap[pos].Obj = b.Obj

	/*Multi-tile object*/

	b.Obj.Parent.ObjList =
		append(b.Obj.Parent.ObjList, b.Obj)
	b.Obj.Parent.Parent.PixmapDirty = true
	b.Obj.Parent.NumObjs++
	b.Obj.Parent.Lock.Unlock()

	for p, port := range b.Obj.TypeP.Ports {
		b.Obj.Ports[p].Dir = port.Dir
		b.Obj.Ports[p].Type = port.Type
	}

	b.Obj.Dir = dir

	if obj.TypeP.CanContain {
		obj.Contents = [gv.MAT_MAX]*world.MatData{}
	}

	if obj.TypeP.MaxFuelKG > 0 {
		obj.KGFuel = obj.TypeP.MaxFuelKG
	}

	LinkObj(obj)

	/* Only add to list if the object calls an update function */
	if obj.TypeP.UpdateObj != nil {
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, false)
	}

	if util.ObjHasPort(obj, gv.PORT_OUTPUT) {
		EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, false)
	}

	/* Init obj if we have a function for it */
	if obj.TypeP.InitObj != nil {
		obj.TypeP.InitObj(obj)
	}

	return obj
}
