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

	/* Move this to delete event for object */
	if obj.TypeP.TypeI == gv.ObjTypeBasicMiner {
		obj.Parent.Parent.ResouceDirty = true
	}

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
	obj := util.GetObj(pos, chunk)

	if obj != nil {
		return nil
	}

	world.VisDataDirty.Store(true)

	obj = &world.ObjData{}

	obj.Pos = pos
	obj.Parent = chunk

	obj.TypeP = GameObjTypes[mtype]

	obj.Parent.Lock.Lock()

	obj.Parent.BuildingMap[pos] = &world.BuildingData{}
	obj.Parent.BuildingMap[pos].Obj = obj

	/*Multi-tile object*/

	obj.Parent.ObjList =
		append(obj.Parent.ObjList, obj)
	obj.Parent.Parent.PixmapDirty = true
	obj.Parent.NumObjs++
	obj.Parent.Lock.Unlock()

	for p, port := range obj.TypeP.Ports {
		if obj.Ports[p] == nil {
			obj.Ports[p] = &world.ObjPortData{}
		}
		obj.Ports[p].PortDir = port
	}

	obj.Dir = dir

	for x := 0; x < int(dir); x++ {
		util.RotatePortsCW(obj)
	}

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
