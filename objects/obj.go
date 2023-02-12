package objects

import (
	"GameTest/cwlog"
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

var wg sizedwaitgroup.SizedWaitGroup

/* Loops: Ticks: External, Tocks: Internal, EventQueue, ObjQueue. Locks each list one at a time. Sleeps if needed. Multi-threaded */
func ObjUpdateDaemon() {
	wg = sizedwaitgroup.New(world.NumWorkers)
	var start time.Time

	for !world.MapGenerated.Load() {
		time.Sleep(time.Millisecond * 10)
	}

	for {
		start = time.Now()

		world.TickWorkSize = world.TickCount / (world.NumWorkers * world.WorkChunks)
		if world.TickWorkSize < 1 {
			world.TickWorkSize = 1
		}
		world.TockWorkSize = world.TockCount / (world.NumWorkers * world.WorkChunks)
		if world.TockWorkSize < 1 {
			world.TockWorkSize = 1
		}

		world.TockListLock.Lock()
		runTocks() //Process objects
		world.TockListLock.Unlock()

		world.TickListLock.Lock()
		runTicks() //Move external
		world.TickListLock.Unlock()

		world.EventQueueLock.Lock()
		runEventQueue() //Queue to add/remove events
		world.EventQueueLock.Unlock()

		world.ObjQueueLock.Lock()
		runObjQueue() //Queue to add/remove objects
		world.ObjQueueLock.Unlock()

		if !gv.UPSBench {
			sleepFor := world.ObjectUPS_ns - time.Since(start)
			time.Sleep(sleepFor)
		}

		world.MeasuredObjectUPS_ns = time.Since(start)
	}
}

/* WASM single-thread object update */
func ObjUpdateDaemonST() {
	var start time.Time

	for {
		start = time.Now()

		world.TockListLock.Lock()
		runTocksST() //Process objects
		world.TockListLock.Unlock()

		world.TickListLock.Lock()
		runTicksST() //Move external
		world.TickListLock.Unlock()

		world.EventQueueLock.Lock()
		runEventQueue() //Queue to add/remove events
		world.EventQueueLock.Unlock()

		world.ObjQueueLock.Lock()
		runObjQueue() //Queue to add/remove objects
		world.ObjQueueLock.Unlock()

		if !gv.UPSBench {
			sleepFor := world.ObjectUPS_ns - time.Since(start)
			time.Sleep(sleepFor)
		}
		world.MeasuredObjectUPS_ns = time.Since(start)
	}
}

/* Put our OutputBuffer to another object's InputBuffer (external)*/
func tickObj(obj *world.ObjData) {

	cwlog.DoLog("tick %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
	for p, port := range obj.Ports {

		/* Valid object */
		if port.Obj == nil {
			continue
		}

		/* Only process our outputs */
		if port.PortDir != gv.PORT_OUTPUT {
			//cwlog.DoLog("tickObj: Our port not an output. %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
			continue
		}

		/* If we have stuff to send */
		if port.Buf.Amount == 0 {
			continue
		}

		/* That go to inputs */
		if port.Obj.Ports[util.ReverseDirection(uint8(p))].PortDir != gv.PORT_INPUT {
			//cwlog.DoLog("tickObj: Their port is not an input %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
			continue
		}

		/* And there is somewhere empty to send it */
		if port.Obj.Ports[util.ReverseDirection(uint8(p))].Buf.Amount != 0 {
			//cwlog.DoLog("tickObj: Their input isn't empty %v %v", obj.TypeP.Name, util.CenterXY(obj.Pos))
			continue
		}

		/* Don't send if the object is blocked */
		if port.Obj.Blocked {
			continue
		}

		/*cwlog.DoLog("TICK: %v: %v: %v\n",
		port.Obj.TypeP.Name,
		port.Obj.Ports[p].Buf.TypeP.Name,
		port.Obj.Ports[p].Buf.Amount) */

		port.Obj.Ports[util.ReverseDirection(uint8(p))].Buf.Amount = port.Buf.Amount
		port.Obj.Ports[util.ReverseDirection(uint8(p))].Buf.TypeP = port.Buf.TypeP
		port.Obj.Ports[util.ReverseDirection(uint8(p))].Buf.Rot = port.Buf.Rot
		obj.Ports[p].Buf.Amount = 0
	}
}

/* WASM single thread: Put our OutputBuffer to another object's InputBuffer (external)*/
func runTicksST() {
	for _, item := range world.TickList {
		tickObj(item.Target)
	}
}

/* Process internally in an object, multi-threaded*/
func runTicks() {

	l := world.TickCount - 1
	if l < 1 {
		return
	} else if world.TickWorkSize == 0 {
		return
	}

	numWorkers := l / world.TickWorkSize
	if numWorkers < 1 {
		numWorkers = 1
	}
	each := (l / numWorkers)
	p := 0

	if each < 1 {
		each = l + 1
		numWorkers = 1
	}

	for n := 0; n < numWorkers; n++ {
		//Handle remainder on last worker
		if n == numWorkers-1 {
			each = l + 1 - p
		}

		wg.Add()
		go func(start int, end int) {
			for i := start; i < end; i++ {
				tickObj(world.TickList[i].Target)
			}
			wg.Done()
		}(p, p+each)
		p += each

	}
	wg.Wait()
}

/* Run all object tocks (interal) multi-threaded */
func runTocks() {

	l := world.TockCount - 1
	if l < 1 {
		return
	} else if world.TockWorkSize == 0 {
		return
	}

	numWorkers := l / world.TockWorkSize
	if numWorkers < 1 {
		numWorkers = 1
	}
	each := (l / numWorkers)
	p := 0

	if each < 1 {
		each = l + 1
		numWorkers = 1
	}

	tickNow := time.Now()
	for n := 0; n < numWorkers; n++ {
		//Handle remainder on last worker
		if n == numWorkers-1 {
			each = l + 1 - p
		}

		wg.Add()
		go func(start int, end int, tickNow time.Time) {
			for i := start; i < end; i++ {
				world.TockList[i].Target.TypeP.UpdateObj(world.TockList[i].Target)
			}
			wg.Done()
		}(p, p+each, tickNow)
		p += each

	}
	wg.Wait()
}

/* WASM single-thread: Run all object tocks (interal) */
func runTocksST() {
	for _, item := range world.TockList {
		item.Target.TypeP.UpdateObj(item.Target)
	}
}

/* Lock and append to TickList */
func ticklistAdd(obj *world.ObjData) {

	oPos := util.CenterXY(obj.Pos)
	cwlog.DoLog("tockListAdd: Added %v to TICK list (%v,%v)", obj.TypeP.Name, oPos.X, oPos.Y)

	world.TickList = append(world.TickList, world.TickEvent{Target: obj})
	world.TickCount++
}

/* Lock and append to TockList */
func tockListAdd(obj *world.ObjData) {

	oPos := util.CenterXY(obj.Pos)
	cwlog.DoLog("tockListAdd: Added %v to TOCK list (%v,%v)", obj.TypeP.Name, oPos.X, oPos.Y)

	world.TockList = append(world.TockList, world.TickEvent{Target: obj})
	world.TockCount++
}

/* Lock and add it EventQueue */
func EventQueueAdd(obj *world.ObjData, qtype uint8, delete bool) {

	prefixStr := "Add"
	if delete {
		prefixStr = "Delete"
	}

	qtypeStr := "NONE"
	if qtype == 1 {
		qtypeStr = "TOCK"
	} else if qtype == 2 {
		qtypeStr = "TICK"
	}

	world.EventQueue = append(world.EventQueue, &world.EventQueueData{Obj: obj, QType: qtype, Delete: delete})

	oPos := util.CenterXY(obj.Pos)
	cwlog.DoLog("EventQueue: %v %v for %v (%v,%v)", prefixStr, qtypeStr, obj.TypeP.Name, oPos.X, oPos.Y)
}

/* Lock and remove tick event */
func ticklistRemove(obj *world.ObjData) {

	oPos := util.CenterXY(obj.Pos)
	for i, e := range world.TickList {
		if e.Target == obj {
			cwlog.DoLog("ticklistRemove: Removed %v from the TICK list. (%v,%v)", obj.TypeP.Name, oPos.X, oPos.Y)
			world.TickList = append(world.TickList[:i], world.TickList[i+1:]...)
			world.TickCount--
			return
		}
	}
	cwlog.DoLog("ticklistRemove:Not found in TICK list: %v (%v,%v)", obj.TypeP.Name, oPos.X, oPos.Y)

}

/* lock and remove tock event */
func tocklistRemove(obj *world.ObjData) {
	world.TockListLock.Lock()
	defer world.TockListLock.Unlock()

	oPos := util.CenterXY(obj.Pos)
	for i, e := range world.TockList {
		if e.Target == obj {
			cwlog.DoLog("tocklistRemove: Removed %v from the TOCK list. (%v,%v)", obj.TypeP.Name, oPos.X, oPos.Y)
			world.TockList = append(world.TockList[:i], world.TockList[i+1:]...)
			world.TockCount--
			return
		}
	}
	cwlog.DoLog("tocklistRemove: Not found in TOCK list: %v (%v,%v)", obj.TypeP.Name, oPos.X, oPos.Y)
}

/* UnlinkObj an object's (dir) input */
func UnlinkObj(obj *world.ObjData) {
	oPos := util.CenterXY(obj.Pos)

	for dir, port := range obj.Ports {

		/* Change object port accounting */
		if port.PortDir == gv.PORT_INPUT {
			obj.NumInputs--
			if port.Obj != nil {
				cwlog.DoLog("Unlink: %v (%v,%v): INPUT: %v", obj.TypeP.Name, oPos.X, oPos.Y, util.DirToName(uint8(dir)))
				port.Obj.NumOutputs--

				rObj := port.Obj
				rObj.Ports[util.ReverseDirection(uint8(dir))].Obj = nil

				obj.Ports[dir].Obj = nil
			}
		} else if port.PortDir == gv.PORT_OUTPUT {
			obj.NumOutputs++
			if port.Obj != nil {
				cwlog.DoLog("Unlink: %v (%v,%v): OUTPUT: %v", obj.TypeP.Name, oPos.X, oPos.Y, util.DirToName(uint8(dir)))
				port.Obj.NumInputs--

				rObj := port.Obj
				rObj.Ports[util.ReverseDirection(uint8(dir))].Obj = nil

				obj.Ports[dir].Obj = nil
			}
		}
	}
}

/* Make a superchunk */
func makeSuperChunk(pos world.XY) {
	//Make super chunk if needed

	newPos := pos
	scpos := util.PosToSuperChunkPos(newPos)

	world.SuperChunkMapLock.Lock()  //Lock Superclunk map
	world.SuperChunkListLock.Lock() //Lock Superchunk list

	if world.SuperChunkMap[scpos] == nil {

		/* Make new superchunk in map at pos */
		newSuperChunk := &world.MapSuperChunk{}

		world.SuperChunkMap[scpos] = newSuperChunk
		world.SuperChunkMap[scpos].Lock.Lock() //Lock chunk

		world.SuperChunkList =
			append(world.SuperChunkList, world.SuperChunkMap[scpos])
		world.SuperChunkMap[scpos].ChunkMap = make(map[world.XY]*world.MapChunk)

		/* Save position */
		world.SuperChunkMap[scpos].Pos = scpos

		drawMineral(newSuperChunk)

		world.SuperChunkMap[scpos].Lock.Unlock()
	}
	world.SuperChunkListLock.Unlock()
	world.SuperChunkMapLock.Unlock()
}

/* Make a chunk, insert into superchunk */
func MakeChunk(pos world.XY) bool {
	//Make chunk if needed

	newPos := pos

	makeSuperChunk(pos)

	cpos := util.PosToChunkPos(newPos)
	scpos := util.PosToSuperChunkPos(newPos)

	world.SuperChunkMapLock.Lock()  //Lock Superclunk map
	world.SuperChunkListLock.Lock() //Lock Superchunk list

	if world.SuperChunkMap[scpos].ChunkMap[cpos] == nil {
		/* Increase chunk count */
		world.SuperChunkMap[scpos].NumChunks++

		/* Make a new empty chunk in the map at pos */
		world.SuperChunkMap[scpos].ChunkMap[cpos] = &world.MapChunk{}
		world.SuperChunkMap[scpos].Lock.Lock() //Lock chunk

		/* Append to chunk list */
		world.SuperChunkMap[scpos].ChunkList =
			append(world.SuperChunkMap[scpos].ChunkList, world.SuperChunkMap[scpos].ChunkMap[cpos])

		world.SuperChunkMap[scpos].ChunkMap[cpos].ObjMap = make(map[world.XY]*world.ObjData)

		/* Terrain img */
		world.SuperChunkMap[scpos].ChunkMap[cpos].TerrainImg = world.TempChunkImage
		world.SuperChunkMap[scpos].ChunkMap[cpos].UsingTemporary = true

		/* Save position */
		world.SuperChunkMap[scpos].ChunkMap[cpos].Pos = cpos

		/* Save parent */
		world.SuperChunkMap[scpos].ChunkMap[cpos].Parent = world.SuperChunkMap[scpos]

		world.SuperChunkMap[scpos].Lock.Unlock()

		world.SuperChunkListLock.Unlock()
		world.SuperChunkMapLock.Unlock()
		return true
	}

	world.SuperChunkListLock.Unlock()
	world.SuperChunkMapLock.Unlock()
	return false
}

/* Explore (input) chunks + and - */
func ExploreMap(pos world.XY, input int) {
	/* Explore some map */

	area := input * gv.ChunkSize
	offx := int(pos.X) - (area / 2)
	offy := int(pos.Y) - (area / 2)
	for x := -area; x < area; x += gv.ChunkSize {
		for y := -area; y < area; y += gv.ChunkSize {
			pos := world.XY{X: offx - x, Y: offy - y}
			MakeChunk(pos)
		}
	}
}

/* Create an object, place self in superchunk, chunk and ObjMap, ObjList, add tick/tock events, link inputs/outputs */
func CreateObj(pos world.XY, mtype uint8, dir uint8) *world.ObjData {

	//Make chunk if needed
	if MakeChunk(pos) {
		ExploreMap(pos, 4)
	}
	chunk := util.GetChunk(pos)
	obj := util.GetObj(pos, chunk)

	ppos := util.CenterXY(pos)
	if obj != nil {
		cwlog.DoLog("CreateObj: Object already exists at location: (%v,%v)", ppos.X, ppos.Y)
		return nil
	}

	world.VisDataDirty.Store(true)

	obj = &world.ObjData{}

	obj.Pos = pos
	obj.Parent = chunk

	obj.TypeP = GameObjTypes[mtype]

	cwlog.DoLog("CreateObj: Make %v: (%v,%v)", obj.TypeP.Name, ppos.X, ppos.Y)

	obj.Parent.Lock.Lock()
	obj.Parent.ObjMap[pos] = obj
	obj.Parent.ObjList =
		append(obj.Parent.ObjList, obj)
	obj.Parent.Parent.PixmapDirty = true
	obj.Parent.NumObjects++
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

/* Add to ObjQueue (add/delete world object at end of tick) */
func ObjQueueAdd(obj *world.ObjData, otype uint8, pos world.XY, delete bool, dir uint8) {
	world.ObjQueueLock.Lock()
	world.ObjQueue = append(world.ObjQueue, &world.ObjectQueuetData{Obj: obj, OType: otype, Pos: pos, Delete: delete, Dir: dir})
	world.ObjQueueLock.Unlock()

	prefixStr := "Add"
	if delete {
		prefixStr = "Delete"
	}

	oPos := util.CenterXY(pos)
	cwlog.DoLog("ObjQueue: %v %v (%v,%v)", prefixStr, GameObjTypes[otype].Name, oPos.X, oPos.Y)
}

/* Add/remove tick/tock events from the lists */
func runEventQueue() {

	for _, e := range world.EventQueue {
		if e.Delete {
			switch e.QType {
			case gv.QUEUE_TYPE_TICK:
				ticklistRemove(e.Obj)
			case gv.QUEUE_TYPE_TOCK:
				tocklistRemove(e.Obj)
			}
		} else {
			switch e.QType {
			case gv.QUEUE_TYPE_TICK:
				ticklistAdd(e.Obj)
			case gv.QUEUE_TYPE_TOCK:
				tockListAdd(e.Obj)
			}
		}
	}

	world.EventQueue = []*world.EventQueueData{}
}

/* Add/remove objects from game world at end of tick/tock cycle */
func runObjQueue() {

	for _, item := range world.ObjQueue {
		if item.Delete {

			UnlinkObj(item.Obj)

			/* Remove tick and tock events */
			EventQueueAdd(item.Obj, gv.QUEUE_TYPE_TICK, true)
			EventQueueAdd(item.Obj, gv.QUEUE_TYPE_TOCK, true)

			removeObj(item.Obj)

		} else {
			//Add
			CreateObj(item.Pos, item.OType, item.Dir)
		}
	}

	world.ObjQueue = []*world.ObjectQueuetData{}
}

/* Delete object from ObjMap, ObjList, decerment NumObjects. Marks PixmapDirty */
func removeObj(obj *world.ObjData) {

	oPos := util.CenterXY(obj.Pos)
	cwlog.DoLog("removeObj: Deleted %v from chunk ObjMap at (%v,%v)", obj.TypeP.Name, oPos.X, oPos.Y)

	/* delete from map */
	obj.Parent.Lock.Lock()
	defer obj.Parent.Lock.Unlock()

	obj.Parent.NumObjects--
	delete(obj.Parent.ObjMap, obj.Pos)
	util.ObjListDelete(obj)

	obj.Parent.Parent.PixmapDirty = true
}
