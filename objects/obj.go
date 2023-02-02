package objects

import (
	"GameTest/consts"
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/util"
	"sync/atomic"
	"time"

	"github.com/remeh/sizedwaitgroup"
	"github.com/sasha-s/go-deadlock"
)

var (
	gWorldTick uint64 = 0

	TickList     []glob.TickEvent = []glob.TickEvent{}
	TickListLock deadlock.RWMutex

	TockList     []glob.TickEvent = []glob.TickEvent{}
	TockListLock deadlock.RWMutex

	ObjQueue     []*glob.ObjectHitlistData
	ObjQueueLock deadlock.RWMutex

	EventQueue     []*glob.EventHitlistData
	EventQueueLock deadlock.RWMutex

	gTickCount    int
	gTockCount    int
	gTickWorkSize int

	TockWorkSize atomic.Uint32
	NumWorkers   int

	wg sizedwaitgroup.SizedWaitGroup
)

func ObjUpdateDaemon() {
	var start time.Time
	wg = sizedwaitgroup.New(NumWorkers)

	for {

		gen := glob.MapGenerated.Load()

		if !gen {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		start = time.Now()

		gWorldTick++
		gTickWorkSize = gTickCount / NumWorkers
		if gTickWorkSize < 1 {
			gTickWorkSize = 1
		}
		TockWorkSize.Store(uint32(gTockCount / NumWorkers))
		if TockWorkSize.Load() < 1 {
			TockWorkSize.Store(1)
		}

		runTocks()         //Process objects
		runTicks()         //Move external
		runEventHitlist()  //Queue to add/remove events
		runObjectHitlist() //Queue to add/remove objects

		if !consts.UPSBench {
			sleepFor := glob.ObjectUPS_ns - time.Since(start)
			time.Sleep(sleepFor)
		} else {
			if glob.FixWASM {
				time.Sleep(time.Millisecond)
			}
		}
		glob.MeasuredObjectUPS_ns = time.Since(start)
	}

}

func ObjUpdateDaemonST() {
	var start time.Time

	for {

		if !glob.MapGenerated.Load() {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		start = time.Now()

		runTocksST()       //Process objects
		runTicksST()       //Move external
		runEventHitlist()  //Queue to add/remove events
		runObjectHitlist() //Queue to add/remove objects

		if !consts.UPSBench {
			sleepFor := glob.ObjectUPS_ns - time.Since(start)
			time.Sleep(sleepFor)
		} else {
			if glob.FixWASM {
				time.Sleep(time.Millisecond)
			}
		}
		glob.MeasuredObjectUPS_ns = time.Since(start)
	}
}

/* Output to another object */
func tickObj(o *glob.ObjData) {

	if o.OutputObj != nil {
		revDir := util.ReverseDirection(o.Direction)
		if o.OutputBuffer.Amount > 0 &&
			o.OutputObj.InputBuffer[revDir] != nil &&
			o.OutputObj.InputBuffer[revDir].Amount == 0 {

			o.OutputObj.InputBuffer[revDir].Amount = o.OutputBuffer.Amount
			o.OutputObj.InputBuffer[revDir].TypeP = o.OutputBuffer.TypeP

			o.OutputBuffer.Amount = 0

			o.BlinkGreen = 2
		}
	} else {
		o.BlinkRed = 4
	}
}

// Move materials from one object to another
func runTicksST() {

	TickListLock.RLock()
	TickListTemp := TickList
	TickListLock.RUnlock()

	for _, item := range TickListTemp {
		tickObj(item.Target)
	}
}

// Move materials from one object to another
func runTicks() {

	TickListLock.RLock()
	TickListTmp := TickList
	TickListLock.RUnlock()

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
				tickObj(TickListTmp[i].Target)
			}
			wg.Done()
		}(p, p+each)
		p += each

	}
	wg.Wait()
}

// Process objects
func runTocks() {

	TockListLock.RLock()
	TockListTmp := TockList
	TockListLock.RUnlock()

	l := gTockCount - 1
	if l < 1 {
		return
	} else if TockWorkSize.Load() == 0 {
		return
	}

	numWorkers := l / int(TockWorkSize.Load())
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
				TockListTmp[i].Target.TypeP.UpdateObj(TockList[i].Target)
			}
			wg.Done()
		}(p, p+each, tickNow)
		p += each

	}
	wg.Wait()
}

// Process objects
func runTocksST() {
	TockListLock.RLock()
	TockListTmp := TockList
	TockListLock.RUnlock()

	for _, item := range TockListTmp {
		item.Target.TypeP.UpdateObj(item.Target)
	}
}

func ticklistAdd(target *glob.ObjData) {
	TickListLock.Lock()
	defer TickListLock.Unlock()

	TickList = append(TickList, glob.TickEvent{Target: target})
	gTickCount++
	cwlog.DoLog("Added: %v to ticklist.", target.TypeP.Name)
}

func tockListAdd(target *glob.ObjData) {
	TockListLock.Lock()
	defer TockListLock.Unlock()

	TockList = append(TockList, glob.TickEvent{Target: target})
	gTockCount++
	cwlog.DoLog("Added: %v to tocklist.", target.TypeP.Name)
}

func EventQueueAdd(obj *glob.ObjData, qtype int, delete bool) {
	EventQueueLock.Lock()
	defer EventQueueLock.Unlock()

	EventQueue = append(EventQueue, &glob.EventHitlistData{Obj: obj, QType: qtype, Delete: delete})
	cwlog.DoLog("Added: %v to the event type: %v hitlist. Delete: %v", obj.TypeP.Name, qtype, delete)
}

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

func unlinkInput(obj *glob.ObjData, dir int) {
	if obj.TypeP.HasMatInput > 0 {
		if obj.InputObjs[util.ReverseDirection(dir)] != nil {
			obj.OutputObj = nil
		}
	}
}

func unlinkOut(obj *glob.ObjData) {
	if obj.TypeP.HasMatOutput {
		if obj.OutputObj != nil {
			/* Remove ourself from input list */
			obj.OutputObj.InputObjs[util.ReverseDirection(obj.Direction)] = nil

			/* Erase output pointer */
			obj.OutputObj = nil
		}
	}
}
func linkOut(pos glob.XY, obj *glob.ObjData, dir int) {

	ppos := util.CenterXY(pos)

	/* Don't bother if we don't have outputs */
	if !obj.TypeP.HasMatOutput {
		cwlog.DoLog("(%v: %v, %v) linkOut: we do not have any outputs", obj.TypeP.Name, ppos.X, ppos.Y)
		return
	}

	/* Look for object in output direction */
	neigh, _ := util.GetNeighborObj(obj, pos, dir)

	/* Did we find and obj? */
	if neigh == nil {
		cwlog.DoLog("(%v: %v, %v) linkOut: Rejected nil neighbor: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
		return
	}
	/* Does it have inputs? */
	if neigh.TypeP.HasMatInput == 0 {
		cwlog.DoLog("(%v: %v, %v) linkOut: Rejected: neighbor has no inputs: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
		return
	}
	/* Do they have an output? */
	if neigh.TypeP.HasMatOutput {
		/* Are we trying to connect from that direction? */
		if neigh.TypeP.Direction == util.ReverseDirection(dir) {
			cwlog.DoLog("(%v: %v, %v) linkOut: Rejected: neighbor outputs this direction: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(util.ReverseDirection(dir)))
			return
		}
	}

	/* If we have an output already, unlink it */
	if obj.OutputObj != nil {
		/* Unlink OLD output specifically */
		unlinkOut(obj.OutputObj)
		cwlog.DoLog("(%v: %v, %v) linkOut: removing our output: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(obj.Direction))
	}

	/* Make sure the object has an input initialized */
	if neigh.InputBuffer[util.ReverseDirection(dir)] != nil {
		neigh.InputBuffer[util.ReverseDirection(dir)] = &glob.MatData{}
		cwlog.DoLog("(%v: %v, %v) linkOut: init neighbor input: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
	}

	/* Make sure our output is initalized */
	if obj.OutputBuffer == nil {
		obj.OutputBuffer = &glob.MatData{}
		cwlog.DoLog("(%v: %v, %v) linkOut: init our output.", obj.TypeP.Name, ppos.X, ppos.Y)
	}

	/* Mark target as our output */
	obj.OutputObj = neigh

	/* Put ourself in target's input list */
	neigh.InputObjs[util.ReverseDirection(dir)] = obj

	cwlog.DoLog("(%v: %v, %v) linkOut: Linked: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))

}

func linkIn(pos glob.XY, obj *glob.ObjData, newdir int) {
	ppos := util.CenterXY(pos)

	/* Don't bother if we don't have inputs */
	if obj.TypeP.HasMatInput == 0 {
		cwlog.DoLog("(%v: %v, %v) linkIn: we have no inputs.", obj.TypeP.Name, ppos.X, ppos.Y)
		return
	}

	for dir := consts.DIR_NORTH; dir < consts.DIR_MAX; dir++ {

		/* Don't try to connect an input the same direction as our future output */
		/* If there is an input there, remove it */
		if obj.TypeP.HasMatOutput && dir == newdir {
			unlinkInput(obj, dir)
			cwlog.DoLog("(%v: %v, %v) linkIn: unlinking input that is in direction of our new output: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
			continue
		}

		/* Look for neighbor object */
		neigh, _ := util.GetNeighborObj(obj, pos, dir)

		/* Did we find an object? */
		if neigh == nil {
			cwlog.DoLog("(%v: %v, %v) linkIn: nil neighbor: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
			continue
		}

		/* Does it have an output? */
		if !neigh.TypeP.HasMatOutput {
			cwlog.DoLog("(%v: %v, %v) linkIn: neighbor has no outputs: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
			continue
		}

		/* Is the output unoccupied? */
		if neigh.OutputObj != nil {
			/* Is it us? */
			if neigh.OutputObj != obj {
				cwlog.DoLog("(%v: %v, %v) linkIn: neigbor output is occupied: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
				continue
			}
		}

		/* Is the output in our direction? */
		if neigh.Direction != util.ReverseDirection(dir) {
			cwlog.DoLog("(%v: %v, %v) linkIn: neighbor output is not in our direction: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
			continue
		}

		/* Unlink old input from this direction if it exists */
		unlinkInput(obj, dir)

		/* Make sure they have an output initalized */
		if neigh.OutputBuffer == nil {
			neigh.OutputBuffer = &glob.MatData{}
			cwlog.DoLog("(%v: %v, %v) linkIn: initializing neighbor output: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
		}

		/* Make sure we have a input initalized */
		if obj.InputBuffer[dir] == nil {
			obj.InputBuffer[dir] = &glob.MatData{}
			cwlog.DoLog("(%v: %v, %v) linkIn: initializing our input : %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
		}

		/* Set ourself as their output */
		neigh.OutputObj = neigh

		/* Record who is on this input */
		obj.InputObjs[util.ReverseDirection(dir)] = neigh

		cwlog.DoLog("(%v: %v, %v) linkIn: linked: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
	}

}

func LinkObj(pos glob.XY, obj *glob.ObjData, newdir int) {
	linkIn(pos, obj, newdir)
	linkOut(pos, obj, newdir)
}

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

		/* Save position */
		glob.SuperChunkMap[scpos].ChunkMap[cpos].Pos = cpos

		/* Save parent */
		glob.SuperChunkMap[scpos].ChunkMap[cpos].Parent = glob.SuperChunkMap[scpos]

		glob.SuperChunkMap[scpos].Lock.Unlock()
	}
}

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

func CreateObj(pos glob.XY, mtype int, dir int) *glob.ObjData {

	//Make chunk if needed
	MakeChunk(pos)
	chunk := util.GetChunk(pos)

	glob.CameraDirty.Store(true)

	obj := util.GetObj(pos, chunk)

	ppos := util.CenterXY(pos)
	if obj != nil {
		cwlog.DoLog("CreateObj: Object already exists at location: %v,%v", ppos.X, ppos.Y)
		return nil
	}

	obj = &glob.ObjData{}

	obj.Pos = pos
	obj.Parent = chunk

	obj.TypeP = GameObjTypes[mtype]
	obj.TypeI = mtype

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
	cpos := util.PosToChunkPos(pos)
	scpos := util.PosToSuperChunkPos(pos)

	glob.SuperChunkMapLock.Lock()
	chunk.Lock.Lock()

	glob.SuperChunkMap[scpos].ChunkMap[cpos].ObjMap[pos] = obj

	glob.SuperChunkMap[scpos].ChunkMap[cpos].ObjList =
		append(glob.SuperChunkMap[scpos].ChunkMap[cpos].ObjList, obj)

	glob.SuperChunkMap[scpos].PixmapDirty = true

	chunk.NumObjects++

	chunk.Lock.Unlock()
	glob.SuperChunkMapLock.Unlock()

	cwlog.DoLog("CreateObj: Make Obj %v: %v,%v", obj.TypeP.Name, ppos.X, ppos.Y)

	LinkObj(pos, obj, dir)

	return obj
}

func ObjectHitlistAdd(obj *glob.ObjData, otype int, pos glob.XY, delete bool, dir int) {
	ObjQueueLock.Lock()
	ObjQueue = append(ObjQueue, &glob.ObjectHitlistData{Obj: obj, OType: otype, Pos: pos, Delete: delete, Dir: dir})
	ObjQueueLock.Unlock()

	ppos := util.CenterXY(pos)
	cwlog.DoLog("Added: %v,%v to the object hitlist. Delete: %v", ppos.X, ppos.Y, delete)
}

func runEventHitlist() {

	EventQueueLock.RLock()
	EventQueueTmp := EventQueue
	EventQueueLock.RUnlock()

	for _, e := range EventQueueTmp {
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

	EventQueueLock.Lock()
	EventQueue = []*glob.EventHitlistData{}
	EventQueueLock.Unlock()
}

func runObjectHitlist() {

	ObjQueueLock.RLock()
	ObjQueueTmp := ObjQueue
	ObjQueueLock.RUnlock()

	for _, item := range ObjQueueTmp {
		if item.Delete {

			/* Invalidate object, and disconnect any connections to us */
			item.Obj.Invalid = true
			for _, inputObj := range item.Obj.InputObjs {
				if inputObj != nil {
					inputObj.OutputObj = nil
				}
			}

			/* Remove tick and tock events */
			EventQueueAdd(item.Obj, consts.QUEUE_TYPE_TICK, true)
			EventQueueAdd(item.Obj, consts.QUEUE_TYPE_TOCK, true)

			removeObj(item.Pos)

		} else {
			//Add
			CreateObj(item.Pos, item.OType, item.Dir)
		}
	}

	ObjQueueLock.Lock()
	ObjQueue = []*glob.ObjectHitlistData{}
	ObjQueueLock.Unlock()
}

func removeObj(pos glob.XY) {
	cpos := util.PosToChunkPos(pos)
	scpos := util.PosToSuperChunkPos(pos)

	/* delete from map */
	glob.SuperChunkMapLock.Lock()
	glob.SuperChunkMap[scpos].Lock.Lock()

	glob.SuperChunkMap[scpos].ChunkMap[cpos].NumObjects--
	delete(glob.SuperChunkMap[scpos].ChunkMap[cpos].ObjMap, pos)

	/* Remove from obj list */
	for index, obj := range glob.SuperChunkMap[scpos].ChunkMap[cpos].ObjList {
		if obj == glob.SuperChunkMap[scpos].ChunkMap[cpos].ObjMap[pos] {
			glob.SuperChunkMap[scpos].ChunkMap[cpos].ObjList =
				append(glob.SuperChunkMap[scpos].ChunkMap[cpos].ObjList[:index],
					glob.SuperChunkMap[scpos].ChunkMap[cpos].ObjList[index+1])
		}
	}

	glob.SuperChunkMap[scpos].PixmapDirty = true

	glob.SuperChunkMap[scpos].Lock.Unlock()
	glob.SuperChunkMapLock.Unlock()
}
