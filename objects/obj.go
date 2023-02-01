package objects

import (
	"GameTest/consts"
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/util"
	"time"

	"github.com/remeh/sizedwaitgroup"
	"github.com/sasha-s/go-deadlock"
)

var (
	gWorldTick uint64 = 0

	ListLock deadlock.Mutex
	TickList []glob.TickEvent = []glob.TickEvent{}
	TockList []glob.TickEvent = []glob.TickEvent{}

	ObjectHitlist []*glob.ObjectHitlistData
	EventHitlist  []*glob.EventHitlistData

	gTickCount    int
	gTockCount    int
	gTickWorkSize int

	TockWorkSize int
	NumWorkers   int

	wg sizedwaitgroup.SizedWaitGroup
)

func ObjUpdateDaemon() {
	var start time.Time
	wg = sizedwaitgroup.New(NumWorkers)

	for {

		if !glob.MapGenerated {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		start = time.Now()

		ListLock.Lock()
		gWorldTick++
		gTickWorkSize = gTickCount / NumWorkers
		if gTickWorkSize < 1 {
			gTickWorkSize = 1
		}
		TockWorkSize = gTockCount / NumWorkers
		if TockWorkSize < 1 {
			TockWorkSize = 1
		}

		runTocks()         //Process objects
		runTicks()         //Move external
		runEventHitlist()  //Queue to add/remove events
		runObjectHitlist() //Queue to add/remove objects
		ListLock.Unlock()

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

		if !glob.MapGenerated {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		start = time.Now()

		ListLock.Lock()
		runTocksST()       //Process objects
		runTicksST()       //Move external
		runEventHitlist()  //Queue to add/remove events
		runObjectHitlist() //Queue to add/remove objects
		ListLock.Unlock()

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
func tickObj(o *glob.WObject) {

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
	for i := 0; i < gTickCount; i++ {
		tickObj(TickList[i].Target)
	}
}

// Move materials from one object to another
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

// Process objects
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

// Process objects
func runTocksST() {
	for i := 0; i < gTockCount; i++ {
		TockList[i].Target.TypeP.UpdateObj(TockList[i].Target)
	}
}

func ticklistAdd(target *glob.WObject) {
	TickList = append(TickList, glob.TickEvent{Target: target})
	gTickCount++
	cwlog.DoLog("Added: %v to ticklist.", target.TypeP.Name)
}

func tockListAdd(target *glob.WObject) {
	TockList = append(TockList, glob.TickEvent{Target: target})
	gTockCount++
	cwlog.DoLog("Added: %v to tocklist.", target.TypeP.Name)
}

func EventHitlistAdd(obj *glob.WObject, qtype int, delete bool) {
	EventHitlist = append(EventHitlist, &glob.EventHitlistData{Obj: obj, QType: qtype, Delete: delete})
	cwlog.DoLog("Added: %v to the event type: %v hitlist. Delete: %v", obj.TypeP.Name, qtype, delete)
}

func ticklistRemove(obj *glob.WObject) {
	for i, e := range TickList {
		if e.Target == obj {
			TickList = append(TickList[:i], TickList[i+1:]...)
			cwlog.DoLog("Removed: %v from the ticklist.", obj.TypeP.Name)
			gTickCount--
			break
		}
	}
}

func tocklistRemove(obj *glob.WObject) {
	for i, e := range TockList {
		if e.Target == obj {
			TockList = append(TockList[:i], TockList[i+1:]...)
			cwlog.DoLog("Removed %v from the tocklist.", obj.TypeP.Name)
			gTockCount--
			break
		}
	}
}

func unlinkInput(obj *glob.WObject, dir int) {
	if obj.TypeP.HasMatInput > 0 {
		if obj.InputObjs[util.ReverseDirection(dir)] != nil {
			obj.OutputObj = nil
		}
	}
}

func unlinkOut(obj *glob.WObject) {
	if obj.TypeP.HasMatOutput {
		if obj.OutputObj != nil {
			/* Remove ourself from input list */
			obj.OutputObj.InputObjs[util.ReverseDirection(obj.Direction)] = nil

			/* Erase output pointer */
			obj.OutputObj = nil
		}
	}
}
func linkOut(pos glob.XY, obj *glob.WObject, dir int) {

	ppos := util.CenterXY(pos)

	/* Don't bother if we don't have outputs */
	if !obj.TypeP.HasMatOutput {
		cwlog.DoLog("(%v: %v, %v) linkOut: we do not have any outputs", obj.TypeP.Name, ppos.X, ppos.Y)
		return
	}

	/* Look for object in output direction */
	neigh, npos := util.GetNeighborObj(obj, pos, dir)

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
	chunk := util.GetChunk(npos)
	obj.OutputObj = chunk.WObject[npos]

	/* Put ourself in target's input list */
	neigh.InputObjs[util.ReverseDirection(dir)] = obj

	cwlog.DoLog("(%v: %v, %v) linkOut: Linked: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))

}

func linkIn(pos glob.XY, obj *glob.WObject, newdir int) {
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
		chunk := util.GetChunk(pos)
		if chunk == nil {
			cwlog.DoLog("(%v: %v, %v) linkIn: failed to find our chunk?", obj.TypeP.Name, ppos.X, ppos.Y)
			continue
		}
		if chunk.WObject[pos] == nil {
			cwlog.DoLog("(%v: %v, %v) linkIn: failed to find ourself?", obj.TypeP.Name, ppos.X, ppos.Y)
			continue
		}
		neigh.OutputObj = chunk.WObject[pos]

		/* Record who is on this input */
		obj.InputObjs[util.ReverseDirection(dir)] = neigh

		cwlog.DoLog("(%v: %v, %v) linkIn: linked: %v", obj.TypeP.Name, ppos.X, ppos.Y, util.DirToName(dir))
	}

}

func LinkObj(pos glob.XY, obj *glob.WObject, newdir int) {
	linkIn(pos, obj, newdir)
	linkOut(pos, obj, newdir)
}

func makeSuperChunk(pos glob.XY) {
	//Make super chunk if needed

	newPos := pos
	scpos := util.PosToSuperChunkPos(newPos)

	if glob.SuperChunkMap[scpos] == nil {
		glob.SuperChunkMap[scpos] = &glob.MapSuperChunk{}
		glob.SuperChunkMap[scpos].Chunks = make(map[glob.XY]*glob.MapChunk)
	}

}

func MakeChunk(pos glob.XY) {
	//Make chunk if needed

	newPos := pos

	makeSuperChunk(pos)

	cpos := util.PosToChunkPos(newPos)
	scpos := util.PosToSuperChunkPos(newPos)

	glob.SuperChunkMap[scpos].NumChunks++
	if glob.SuperChunkMap[scpos].Chunks[cpos] == nil {
		glob.SuperChunkMap[scpos].Chunks[cpos] = &glob.MapChunk{}
		glob.SuperChunkMap[scpos].Chunks[cpos].WObject = make(map[glob.XY]*glob.WObject)
	}
}

func ExploreMap(input int) {
	/* Explore some map */

	glob.SuperChunkMapLock.Lock()
	defer glob.SuperChunkMapLock.Unlock()

	area := input * consts.ChunkSize
	offs := int(consts.XYCenter) - (area / 2)
	for x := -area; x < area; x += consts.ChunkSize {
		for y := -area; y < area; y += consts.ChunkSize {
			pos := glob.XY{X: offs - x, Y: offs - y}

			MakeChunk(pos)
		}
	}
}

func CreateObj(pos glob.XY, mtype int, dir int) *glob.WObject {

	glob.SuperChunkMapLock.Lock()
	defer glob.SuperChunkMapLock.Unlock()

	//Make chunk if needed
	MakeChunk(pos)
	chunk := util.GetChunk(pos)

	glob.CameraDirty = true
	obj := chunk.WObject[pos]

	ppos := util.CenterXY(pos)
	if obj != nil {
		cwlog.DoLog("CreateObj: Object already exists at location: %v,%v", ppos.X, ppos.Y)
		return nil
	}

	obj = &glob.WObject{}

	obj.TypeP = GameObjTypes[mtype]
	obj.TypeI = mtype

	obj.Contents = [consts.MAT_MAX]*glob.MatData{}
	if obj.TypeP.HasMatOutput {
		obj.Direction = dir
	}

	/* Only add to list if the object calls an update function */
	if obj.TypeP.UpdateObj != nil {
		EventHitlistAdd(obj, consts.QUEUE_TYPE_TOCK, false)
	}

	if obj.TypeP.HasMatOutput {
		EventHitlistAdd(obj, consts.QUEUE_TYPE_TICK, false)
		obj.OutputBuffer = &glob.MatData{}
	}
	cpos := util.PosToChunkPos(pos)
	scpos := util.PosToSuperChunkPos(pos)
	glob.SuperChunkMap[scpos].Chunks[cpos].WObject[pos] = obj

	cwlog.DoLog("CreateObj: Make Obj %v: %v,%v", obj.TypeP.Name, ppos.X, ppos.Y)

	chunk.NumObjects++
	LinkObj(pos, obj, dir)

	return obj
}

func ObjectHitlistAdd(obj *glob.WObject, otype int, pos glob.XY, delete bool, dir int) {
	ObjectHitlist = append(ObjectHitlist, &glob.ObjectHitlistData{Obj: obj, OType: otype, Pos: pos, Delete: delete, Dir: dir})

	ppos := util.CenterXY(pos)
	cwlog.DoLog("Added: %v,%v to the object hitlist. Delete: %v", ppos.X, ppos.Y, delete)
}

func runEventHitlist() {
	for _, e := range EventHitlist {
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
	EventHitlist = []*glob.EventHitlistData{}
}

func runObjectHitlist() {

	for _, item := range ObjectHitlist {
		if item.Delete {

			/* Invalidate object, and disconnect any connections to us */
			item.Obj.Invalid = true
			for _, inputObj := range item.Obj.InputObjs {
				if inputObj != nil {
					inputObj.OutputObj = nil
				}
			}

			/* Remove tick and tock events */
			go func(obj glob.WObject, pos glob.XY) {
				ListLock.Lock()
				EventHitlistAdd(&obj, consts.QUEUE_TYPE_TICK, true)
				EventHitlistAdd(&obj, consts.QUEUE_TYPE_TOCK, true)

				removeObj(pos)
				ListLock.Unlock()
			}(*item.Obj, item.Pos)

		} else {
			//Add
			CreateObj(item.Pos, item.OType, item.Dir)
		}
	}
	ObjectHitlist = []*glob.ObjectHitlistData{}
}

func removeObj(pos glob.XY) {
	cpos := util.PosToChunkPos(pos)
	scpos := util.PosToSuperChunkPos(pos)

	/* Remove from chunk */
	glob.SuperChunkMapLock.Lock()
	glob.SuperChunkMap[scpos].Chunks[cpos].NumObjects--
	delete(glob.SuperChunkMap[scpos].Chunks[cpos].WObject, pos)
	glob.SuperChunkMapLock.Unlock()
}
