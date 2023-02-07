package objects

import (
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/gv"
	"GameTest/util"
	"fmt"
	"sync"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

var (
	TickList     []glob.TickEvent = []glob.TickEvent{}
	TickListLock sync.Mutex

	TockList     []glob.TickEvent = []glob.TickEvent{}
	TockListLock sync.Mutex

	ObjQueue     []*glob.ObjectQueuetData
	ObjQueueLock sync.Mutex

	EventQueue     []*glob.EventQueueData
	EventQueueLock sync.Mutex

	gTickCount    int
	gTockCount    int
	gTickWorkSize int

	TockWorkSize int
	NumWorkers   int

	wg sizedwaitgroup.SizedWaitGroup
)

/* Loops: Ticks: External, Tocks: Internal, EventQueue, ObjQueue. Locks each list one at a time. Sleeps if needed. Multi-threaded */
func ObjUpdateDaemon() {
	wg = sizedwaitgroup.New(NumWorkers)
	var start time.Time

	for !glob.MapGenerated.Load() {
		time.Sleep(time.Millisecond * 100)
	}

	for {
		start = time.Now()

		gTickWorkSize = gTickCount / NumWorkers
		if gTickWorkSize < 1 {
			gTickWorkSize = 1
		}
		TockWorkSize = (gTockCount / NumWorkers)
		if TockWorkSize < 1 {
			TockWorkSize = 1
		}

		runTocks() //Process objects
		runTicks() //Move external
		EventQueueLock.Lock()
		runEventQueue() //Queue to add/remove events
		EventQueueLock.Unlock()
		ObjQueueLock.Lock()
		runObjQueue() //Queue to add/remove objects
		ObjQueueLock.Unlock()

		if !gv.UPSBench {
			sleepFor := glob.ObjectUPS_ns - time.Since(start)
			time.Sleep(sleepFor)
		} else {
			if glob.WASMMode {
				time.Sleep(time.Nanosecond)
			}
		}

		glob.MeasuredObjectUPS_ns = time.Since(start)
	}
}

/* WASM single-thread object update */
func ObjUpdateDaemonST() {
	var start time.Time

	time.Sleep(time.Second)
	for !glob.MapGenerated.Load() {
		time.Sleep(time.Millisecond * 100)
	}

	for {
		start = time.Now()

		runTocksST()    //Process objects
		runTicksST()    //Move external
		runEventQueue() //Queue to add/remove events
		runObjQueue()   //Queue to add/remove objects

		if !gv.UPSBench {
			sleepFor := glob.ObjectUPS_ns - time.Since(start)
			time.Sleep(sleepFor)
		} else {
			if glob.WASMMode {
				time.Sleep(time.Nanosecond)
			}
		}
		glob.MeasuredObjectUPS_ns = time.Since(start)
	}
}

/* Put our OutputBuffer to another object's InputBuffer (external)*/
func tickObj(obj *glob.ObjData) {
	if obj.NumOutputs > 0 {
		cwlog.DoLog("meep %v %v", obj.TypeP.Name, obj.Pos.X)
		for p, port := range obj.Ports {

			/* Valid object */
			if port.Obj == nil {
				continue
			}

			/* Only process our outputs */
			if port.PortDir != gv.PORT_OUTPUT {
				continue
			}

			/* If we have stuff to send */
			if port.Buf.Amount == 0 {
				continue
			}

			/* That go to inputs */
			if port.Obj.Ports[util.ReverseDirection(uint8(p))].PortDir != gv.PORT_INPUT {
				continue
			}

			/* And there is somewhere empty to send it */
			if port.Obj.Ports[util.ReverseDirection(uint8(p))].Buf.Amount == 0 {
				continue
			}

			fmt.Printf("TICK: %v: %v: %v\n",
				port.Obj.TypeP.Name,
				port.Obj.Ports[p].Buf.TypeP.Name,
				port.Obj.Ports[p].Buf.Amount)

			port.Obj.Ports[p].Buf.Amount = port.Buf.Amount
			port.Obj.Ports[p].Buf.TypeP = port.Buf.TypeP
			port.Buf.Amount = 0
			obj.Blocked = false
		}

	}

}

/* WASM single thread: Put our OutputBuffer to another object's InputBuffer (external)*/
func runTicksST() {
	for _, item := range TickList {
		tickObj(item.Target)
	}
}

/* Process internally in an object, multi-threaded*/
func runTicks() {

	l := gTickCount - 1
	if l < 1 {
		return
	} else if gTickWorkSize == 0 {
		return
	}

	numWorkers := l / gTickWorkSize
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
				tickObj(TickList[i].Target)
			}
			wg.Done()
		}(p, p+each)
		p += each

	}
	wg.Wait()
}

/* Run all object tocks (interal) multi-threaded */
func runTocks() {

	l := gTockCount - 1
	if l < 1 {
		return
	} else if TockWorkSize == 0 {
		return
	}

	numWorkers := l / TockWorkSize
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
				TockList[i].Target.TypeP.UpdateObj(TockList[i].Target)
			}
			wg.Done()
		}(p, p+each, tickNow)
		p += each

	}
	wg.Wait()
}

/* WASM single-thread: Run all object tocks (interal) */
func runTocksST() {
	for _, item := range TockList {
		item.Target.TypeP.UpdateObj(item.Target)
	}
}

/* Lock and append to TickList */
func ticklistAdd(obj *glob.ObjData) {
	TickListLock.Lock()
	defer TickListLock.Unlock()

	oPos := util.CenterXY(obj.Pos)
	cwlog.DoLog("tockListAdd: Added %v to TICK list (%v,%v)", obj.TypeP.Name, oPos.X, oPos.Y)

	TickList = append(TickList, glob.TickEvent{Target: obj})
	gTickCount++
}

/* Lock and append to TockList */
func tockListAdd(obj *glob.ObjData) {
	TockListLock.Lock()
	defer TockListLock.Unlock()

	oPos := util.CenterXY(obj.Pos)
	cwlog.DoLog("tockListAdd: Added %v to TOCK list (%v,%v)", obj.TypeP.Name, oPos.X, oPos.Y)

	TockList = append(TockList, glob.TickEvent{Target: obj})
	gTockCount++
}

/* Lock and add it EventQueue */
func EventQueueAdd(obj *glob.ObjData, qtype uint8, delete bool) {
	EventQueueLock.Lock()
	defer EventQueueLock.Unlock()

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

	EventQueue = append(EventQueue, &glob.EventQueueData{Obj: obj, QType: qtype, Delete: delete})

	oPos := util.CenterXY(obj.Pos)
	cwlog.DoLog("EventQueue: %v %v for %v (%v,%v)", prefixStr, qtypeStr, obj.TypeP.Name, oPos.X, oPos.Y)
}

/* Lock and remove tick event */
func ticklistRemove(obj *glob.ObjData) {
	TickListLock.Lock()
	defer TickListLock.Unlock()

	oPos := util.CenterXY(obj.Pos)
	for i, e := range TickList {
		if e.Target == obj {
			cwlog.DoLog("ticklistRemove: Removed %v from the TICK list. (%v,%v)", obj.TypeP.Name, oPos.X, oPos.Y)
			TickList = append(TickList[:i], TickList[i+1:]...)
			gTickCount--
			return
		}
	}
	cwlog.DoLog("ticklistRemove:Not found in TICK list: %v (%v,%v)", obj.TypeP.Name, oPos.X, oPos.Y)

}

/* lock and remove tock event */
func tocklistRemove(obj *glob.ObjData) {
	TockListLock.Lock()
	defer TockListLock.Unlock()

	oPos := util.CenterXY(obj.Pos)
	for i, e := range TockList {
		if e.Target == obj {
			cwlog.DoLog("tocklistRemove: Removed %v from the TOCK list. (%v,%v)", obj.TypeP.Name, oPos.X, oPos.Y)
			TockList = append(TockList[:i], TockList[i+1:]...)
			gTockCount--
			return
		}
	}
	cwlog.DoLog("tocklistRemove: Not found in TOCK list: %v (%v,%v)", obj.TypeP.Name, oPos.X, oPos.Y)
}

/* UnlinkObj an object's (dir) input */
func UnlinkObj(obj *glob.ObjData) {
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
func makeSuperChunk(pos glob.XY) {
	//Make super chunk if needed

	newPos := pos
	scpos := util.PosToSuperChunkPos(newPos)

	glob.SuperChunkMapLock.Lock()
	defer glob.SuperChunkMapLock.Unlock()

	if glob.SuperChunkMap[scpos] == nil {
		/* Make new superchunk in map at pos */
		glob.SuperChunkMap[scpos] = &glob.MapSuperChunk{}
		glob.SuperChunkMap[scpos].Lock.Lock()

		/* Append to superchunk list */
		glob.SuperChunkListLock.Lock()
		glob.SuperChunkList =
			append(glob.SuperChunkList, glob.SuperChunkMap[scpos])
		glob.SuperChunkListLock.Unlock()

		glob.SuperChunkMap[scpos].ChunkMap = make(map[glob.XY]*glob.MapChunk)

		/* Save position */
		glob.SuperChunkMap[scpos].Pos = scpos

		glob.SuperChunkMap[scpos].Lock.Unlock()
	}

}

/* Make a chunk, insert into superchunk */
func MakeChunk(pos glob.XY) {
	//Make chunk if needed

	newPos := pos

	makeSuperChunk(pos)

	cpos := util.PosToChunkPos(newPos)
	scpos := util.PosToSuperChunkPos(newPos)

	if glob.SuperChunkMap[scpos].ChunkMap[cpos] == nil {

		/* Increase chunk count */
		glob.SuperChunkMap[scpos].NumChunks++

		/* Make a new empty chunk in the map at pos */
		glob.SuperChunkMap[scpos].ChunkMap[cpos] = &glob.MapChunk{}
		glob.SuperChunkMap[scpos].Lock.Lock()

		/* Append to chunk list */
		glob.SuperChunkMap[scpos].ChunkList =
			append(glob.SuperChunkMap[scpos].ChunkList, glob.SuperChunkMap[scpos].ChunkMap[cpos])

		glob.SuperChunkMap[scpos].ChunkMap[cpos].ObjMap = make(map[glob.XY]*glob.ObjData)

		/* Terrain img */
		glob.SuperChunkMap[scpos].ChunkMap[cpos].TerrainImg = glob.TempChunkImage
		glob.SuperChunkMap[scpos].ChunkMap[cpos].UsingTemporary = true

		/* Save position */
		glob.SuperChunkMap[scpos].ChunkMap[cpos].Pos = cpos

		/* Save parent */
		glob.SuperChunkMap[scpos].ChunkMap[cpos].Parent = glob.SuperChunkMap[scpos]

		glob.SuperChunkMap[scpos].Lock.Unlock()
	}
}

/* Explore (input) chunks + and - */
func ExploreMap(input int) {
	/* Explore some map */

	area := input * gv.ChunkSize
	offs := int(gv.XYCenter) - (area / 2)
	for x := -area; x < area; x += gv.ChunkSize {
		for y := -area; y < area; y += gv.ChunkSize {
			pos := glob.XY{X: offs - x, Y: offs - y}

			MakeChunk(pos)
		}
	}
}

/* Create an object, place self in superchunk, chunk and ObjMap, ObjList, add tick/tock events, link inputs/outputs */
func CreateObj(pos glob.XY, mtype uint8, dir uint8) *glob.ObjData {

	//Make chunk if needed
	MakeChunk(pos)
	chunk := util.GetChunk(pos)
	obj := util.GetObj(pos, chunk)

	ppos := util.CenterXY(pos)
	if obj != nil {
		cwlog.DoLog("CreateObj: Object already exists at location: (%v,%v)", ppos.X, ppos.Y)
		return nil
	}

	glob.VisDataDirty.Store(true)

	obj = &glob.ObjData{}

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
		obj.Ports[p].PortDir = port
	}

	if obj.TypeP.Rotatable {
		obj.Dir = dir
	}

	for x := 0; x < int(dir); x++ {
		util.RotatePortsCW(obj)
	}

	if obj.TypeP.CanContain {
		obj.Contents = [gv.MAT_MAX]*glob.MatData{}
	}

	LinkObj(obj)

	/* Only add to list if the object calls an update function */
	if obj.TypeP.UpdateObj != nil {
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, false)
	}

	for _, port := range obj.TypeP.Ports {
		if port == gv.PORT_OUTPUT {
			EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, false)
			break
		}
	}

	return obj
}

/* Add to ObjQueue (add/delete world object at end of tick) */
func ObjQueueAdd(obj *glob.ObjData, otype uint8, pos glob.XY, delete bool, dir uint8) {
	ObjQueueLock.Lock()
	ObjQueue = append(ObjQueue, &glob.ObjectQueuetData{Obj: obj, OType: otype, Pos: pos, Delete: delete, Dir: dir})
	ObjQueueLock.Unlock()

	prefixStr := "Add"
	if delete {
		prefixStr = "Delete"
	}

	oPos := util.CenterXY(pos)
	cwlog.DoLog("ObjQueue: %v %v (%v,%v)", prefixStr, GameObjTypes[otype].Name, oPos.X, oPos.Y)
}

/* Add/remove tick/tock events from the lists */
func runEventQueue() {

	for _, e := range EventQueue {
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

	EventQueue = []*glob.EventQueueData{}
}

/* Add/remove objects from game world at end of tick/tock cycle */
func runObjQueue() {

	for _, item := range ObjQueue {
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

	ObjQueue = []*glob.ObjectQueuetData{}
}

/* Delete object from ObjMap, ObjList, decerment NumObjects. Marks PixmapDirty */
func removeObj(obj *glob.ObjData) {

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
