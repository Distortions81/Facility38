package objects

import (
	"GameTest/consts"
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/util"
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

		if !consts.UPSBench {
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

		if !consts.UPSBench {
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
func tickObj(o *glob.ObjData) {

	if o.OutputObj != nil {
		revDir := util.ReverseDirection(o.Direction)
		if o.OutputBuffer.Amount > 0 &&
			o.OutputObj.InputBuffer[revDir] != nil &&
			o.OutputObj.InputBuffer[revDir].Amount == 0 {

			o.OutputObj.InputBuffer[revDir].Amount = o.OutputBuffer.Amount
			o.OutputObj.InputBuffer[revDir].TypeP = o.OutputBuffer.TypeP

			o.OutputBuffer.Amount = 0
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
func ticklistAdd(target *glob.ObjData) {
	TickListLock.Lock()
	defer TickListLock.Unlock()

	TickList = append(TickList, glob.TickEvent{Target: target})
	gTickCount++
	cwlog.DoLog("Added: %v to ticklist.", target.TypeP.Name)
}

/* Lock and append to TockList */
func tockListAdd(target *glob.ObjData) {
	TockListLock.Lock()
	defer TockListLock.Unlock()

	TockList = append(TockList, glob.TickEvent{Target: target})
	gTockCount++
	cwlog.DoLog("Added: %v to tocklist.", target.TypeP.Name)
}

/* Lock and add it EventQueue */
func EventQueueAdd(obj *glob.ObjData, qtype uint8, delete bool) {
	EventQueueLock.Lock()
	defer EventQueueLock.Unlock()

	EventQueue = append(EventQueue, &glob.EventQueueData{Obj: obj, QType: qtype, Delete: delete})
	cwlog.DoLog("Added: %v to the event type: %v hitlist. Delete: %v", obj.TypeP.Name, qtype, delete)
}

/* Lock and remove tick event */
func ticklistRemove(obj *glob.ObjData) {
	TickListLock.Lock()
	defer TickListLock.Unlock()

	for i, e := range TickList {
		if e.Target == obj {
			TickList = append(TickList[:i], TickList[i+1:]...)
			cwlog.DoLog("Removed: %v from the ticklist.", obj.TypeP.Name)
			gTickCount--
			break
		}
	}
}

/* lock and remove tock event */
func tocklistRemove(obj *glob.ObjData) {
	TockListLock.Lock()
	defer TockListLock.Unlock()

	for i, e := range TockList {
		if e.Target == obj {
			TockList = append(TockList[:i], TockList[i+1:]...)
			cwlog.DoLog("Removed %v from the tocklist.", obj.TypeP.Name)
			gTockCount--
			break
		}
	}
}

/* Unlink an object's (dir) input */
func unlinkInput(obj *glob.ObjData, dir uint8) {
	if obj.TypeP.HasMatInput > 0 {
		if obj.InputObjs[util.ReverseDirection(dir)] != nil {
			obj.InputObjs[util.ReverseDirection(dir)].OutputObj = nil
			obj.InputObjs[util.ReverseDirection(dir)] = nil
			obj.InputCount--
		}
	}
}

/* Unlink and object's output, also removes itself from OutputObj's inputs */
func unlinkOut(obj *glob.ObjData) {
	if obj.TypeP.HasMatOutput {
		if obj.OutputObj != nil {
			/* Remove ourself from input list */
			obj.OutputObj.InputObjs[util.ReverseDirection(obj.Direction)] = nil
			obj.OutputObj.InputCount--

			/* Erase output pointer */
			obj.OutputObj = nil
		}
	}
}

/* Link to output in (dir) */
func linkOut(pos glob.XY, obj *glob.ObjData, dir uint8) {

	ppos := util.CenterXY(pos)

	/* Don't bother if we don't have outputs */
	if !obj.TypeP.HasMatOutput {
		//cwlog.DoLog("(%v: %v, %v) linkOut: we do not have any outputs", obj.TypeP.Name, ppos.X, ppos.Y)
		return
	}

	/* Look for object in output direction */
	neigh, _ := util.GetNeighborObj(obj, pos, dir)

	/* Did we find and obj? */
	if neigh == nil {
		//cwlog.DoLog("(%v: %v, %v) linkOut: Rejected nil neighbor: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
		return
	}
	npos := util.CenterXY(neigh.Pos)

	/* Does it have inputs? */
	if neigh.TypeP.HasMatInput == 0 {
		//cwlog.DoLog("(%v: %v, %v) linkOut: Rejected: neighbor has no inputs: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
		return
	}
	/* Do they have an output? */
	if neigh.TypeP.HasMatOutput {
		/* Are we trying to connect from that direction? */
		if neigh.TypeP.Direction == util.ReverseDirection(dir) {
			cwlog.DoLog("(%v: %v, %v) linkOut: Rejected: neighbor outputs this direction: %v: %v: (%v,%v)",
				obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(util.ReverseDirection(dir)), neigh.TypeP.Name, npos.X, npos.Y)
			return
		}
	}

	/* If we have an output already, unlink it */
	if obj.OutputObj != nil {
		/* Unlink OLD output specifically */
		unlinkOut(obj.OutputObj)
		cwlog.DoLog("(%v: %v, %v) linkOut: removing our old output: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(obj.Direction))
	} else {
		obj.OutputBuffer = &glob.MatData{}
		cwlog.DoLog("(%v: %v, %v) linkOut: init our output buffer.", obj.TypeP.Name, ppos.X, ppos.Y)
	}

	/* Make sure the object has an input initialized */
	if neigh.InputBuffer[util.ReverseDirection(dir)] != nil {
		neigh.InputBuffer[util.ReverseDirection(dir)] = &glob.MatData{}
		cwlog.DoLog("(%v: %v, %v) linkOut: init neighbor input: %v: %v: (%v,%v)",
			obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir), neigh.TypeP.Name, npos.X, npos.Y)
	}

	/* Mark target as our output */
	obj.OutputObj = neigh

	/* Put ourself in target's input list */
	neigh.InputObjs[util.ReverseDirection(dir)] = obj

	cwlog.DoLog("(%v: %v, %v) linkOut: Linked: %v: %v: (%v,%v)",
		obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir), obj.OutputObj.TypeP.Name, npos.X, npos.Y)
}

/* Find and link inputs, set ourself to OutputObj of found objects */
func linkIn(pos glob.XY, obj *glob.ObjData, newdir uint8) {
	ppos := util.CenterXY(pos)

	/* Don't bother if we don't have inputs */
	if obj.TypeP.HasMatInput == 0 {
		//cwlog.DoLog("(%v: %v, %v) linkIn: we have no inputs.", obj.TypeP.Name, ppos.X, ppos.Y)
		return
	}

	var dir uint8
	for dir = consts.DIR_NORTH; dir < consts.DIR_MAX; dir++ {

		/* Don't try to connect an input the same direction as our future output */
		/* If there is an input there, remove it */
		if obj.TypeP.HasMatOutput && dir == newdir {
			unlinkInput(obj, dir)
			cwlog.DoLog("(%v: %v, %v) linkIn: unlinking input that is in direction of our new output: %v",
				obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
			continue
		}

		/* Look for neighbor object */
		neigh, _ := util.GetNeighborObj(obj, pos, dir)

		/* Did we find an object? */
		if neigh == nil {
			//cwlog.DoLog("(%v: %v, %v) linkIn: nil neighbor: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
			continue
		}
		npos := util.CenterXY(neigh.Pos)

		/* Does it have an output? */
		if !neigh.TypeP.HasMatOutput {
			//cwlog.DoLog("(%v: %v, %v) linkIn: neighbor has no outputs: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
			continue
		}

		/* Is the output unoccupied? */
		if neigh.OutputObj != nil {
			/* Is it us? */
			if neigh.OutputObj != obj {
				cwlog.DoLog("(%v: %v, %v) linkIn: neigbor output is occupied: %v: %v: (%v,%v)",
					obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir), neigh.TypeP.Name, npos.X, npos.Y)
				continue
			}
		}

		/* Is the output in our direction? */
		if neigh.Direction != util.ReverseDirection(dir) {
			cwlog.DoLog("(%v: %v, %v) linkIn: neighbor output is not in our direction: %v: %v: (%v,%v)",
				obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir), neigh.TypeP.Name, npos.X, npos.Y)
			continue
		}

		/* Unlink old input from this direction if it exists */
		unlinkInput(obj, dir)

		/* Make sure they have an output initalized */
		if neigh.OutputBuffer == nil {
			neigh.OutputBuffer = &glob.MatData{}
			cwlog.DoLog("(%v: %v, %v) linkIn: initializing neighbor output: %v: %v: (%v,%v)",
				obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir), neigh.TypeP.Name, npos.X, npos.Y)
		}

		/* Make sure we have a input initalized */
		if obj.InputBuffer[dir] == nil {
			obj.InputBuffer[dir] = &glob.MatData{}
			cwlog.DoLog("(%v: %v, %v) linkIn: initializing our input : %v",
				obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
		}

		/* Set ourself as their output */
		linkOut(neigh.Pos, neigh, neigh.Direction)

		/* Record who is on this input */
		obj.InputObjs[util.ReverseDirection(dir)] = neigh
		obj.InputCount++

		cwlog.DoLog("(%v: %v, %v) linkIn: linked: %v: %v: (%v,%v)",
			obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir), neigh.TypeP.Name, npos.X, npos.Y)
	}

}

/* Link inputs and outputs, with output direction (newdir) */
func LinkObj(pos glob.XY, obj *glob.ObjData, newdir uint8) {
	linkIn(pos, obj, newdir)
	linkOut(pos, obj, newdir)
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

	area := input * consts.ChunkSize
	offs := int(consts.XYCenter) - (area / 2)
	for x := -area; x < area; x += consts.ChunkSize {
		for y := -area; y < area; y += consts.ChunkSize {
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
		cwlog.DoLog("CreateObj: Object already exists at location: %v,%v", ppos.X, ppos.Y)
		return nil
	}

	glob.VisDataDirty.Store(true)

	obj = &glob.ObjData{}

	obj.Pos = pos
	obj.Parent = chunk

	obj.TypeP = GameObjTypes[mtype]

	obj.Contents = [consts.MAT_MAX]*glob.MatData{}
	if obj.TypeP.HasMatOutput {
		obj.Direction = dir
	}

	/* Only add to list if the object calls an update function */
	if obj.TypeP.UpdateObj != nil {
		EventQueueAdd(obj, consts.QUEUE_TYPE_TOCK, false)
	}

	if obj.TypeP.HasMatOutput {
		EventQueueAdd(obj, consts.QUEUE_TYPE_TICK, false)
		obj.OutputBuffer = &glob.MatData{}
	}

	obj.Parent.Lock.Lock()
	obj.Parent.ObjMap[pos] = obj
	obj.Parent.ObjList =
		append(obj.Parent.ObjList, obj)
	obj.Parent.Parent.PixmapDirty = true
	obj.Parent.NumObjects++
	obj.Parent.Lock.Unlock()

	cwlog.DoLog("CreateObj: Make Obj %v: %v,%v", obj.TypeP.Name, ppos.X, ppos.Y)

	LinkObj(pos, obj, dir)

	return obj
}

/* Add to ObjQueue (add/delete world object at end of tick) */
func ObjQueueAdd(obj *glob.ObjData, otype uint8, pos glob.XY, delete bool, dir uint8) {
	ObjQueueLock.Lock()
	ObjQueue = append(ObjQueue, &glob.ObjectQueuetData{Obj: obj, OType: otype, Pos: pos, Delete: delete, Dir: dir})
	ObjQueueLock.Unlock()

	ppos := util.CenterXY(pos)
	cwlog.DoLog("Added: %v,%v to the object hitlist. Delete: %v", ppos.X, ppos.Y, delete)
}

/* Add/remove tick/tock events from the lists */
func runEventQueue() {

	for _, e := range EventQueue {
		if e.Delete {
			switch e.QType {
			case consts.QUEUE_TYPE_TICK:
				ticklistRemove(e.Obj)
			case consts.QUEUE_TYPE_TOCK:
				tocklistRemove(e.Obj)
			}
		} else {
			switch e.QType {
			case consts.QUEUE_TYPE_TICK:
				ticklistAdd(e.Obj)
			case consts.QUEUE_TYPE_TOCK:
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

			/* Invalidate object, and disconnect any connections to us */
			for _, inputObj := range item.Obj.InputObjs {
				if inputObj != nil {
					inputObj.OutputObj = nil
				}
			}

			/* Remove tick and tock events */
			EventQueueAdd(item.Obj, consts.QUEUE_TYPE_TICK, true)
			EventQueueAdd(item.Obj, consts.QUEUE_TYPE_TOCK, true)

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
	/* delete from map */
	obj.Parent.Lock.Lock()
	defer obj.Parent.Lock.Unlock()

	obj.Parent.NumObjects--
	delete(obj.Parent.ObjMap, obj.Pos)
	util.ObjListDelete(obj)

	obj.Parent.Parent.PixmapDirty = true
}
