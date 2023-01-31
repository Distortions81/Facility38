package objects

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/util"
	"sync"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

var (
	gWorldTick uint64 = 0

	ListLock sync.Mutex
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

	/*var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	} */

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

	//pprof.StopCPUProfile()

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
		if o.OutputObj.InputBuffer[o.Direction] != nil &&
			o.OutputObj.InputBuffer[o.Direction].Amount == 0 &&
			o.OutputBuffer.Amount > 0 {

			o.OutputObj.InputBuffer[o.Direction].Amount = o.OutputBuffer.Amount
			o.OutputObj.InputBuffer[o.Direction].TypeP = o.OutputBuffer.TypeP

			o.OutputBuffer.Amount = 0
		}
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
}

func tockListAdd(target *glob.WObject) {
	TockList = append(TockList, glob.TickEvent{Target: target})
	gTockCount++
}

func EventHitlistAdd(obj *glob.WObject, qtype int, delete bool) {
	EventHitlist = append(EventHitlist, &glob.EventHitlistData{Obj: obj, QType: qtype, Delete: delete})
}

func ticklistRemove(obj *glob.WObject) {
	for i, e := range TickList {
		if e.Target == obj {

			if len(TickList) > 1 {
				TickList = append(TickList[:i], TickList[i+1:]...)
			} else {
				TickList = []glob.TickEvent{}
			}
			break
		}
	}
	gTickCount--
}

func tocklistRemove(obj *glob.WObject) {
	for i, e := range TockList {
		if e.Target == obj {

			if len(TockList) > 1 {
				TockList = append(TockList[:i], TockList[i+1:]...)
			} else {
				TockList = []glob.TickEvent{}
			}
			break
		}
	}
	gTockCount--
}

func unlinkInput(obj *glob.WObject, dir int) {
	if obj.TypeP.HasMatInput > 0 {
		if obj.InputObjs[dir] != nil {
			obj.OutputObj = nil
		}
	}
}

func unlinkOut(obj *glob.WObject) {
	if obj.TypeP.HasMatOutput {
		if obj.OutputObj != nil {
			/* Remove ourself from input list */
			obj.OutputObj.InputObjs[obj.Direction] = nil

			/* Erase output pointer */
			obj.OutputObj = nil
		}
	}
}
func linkOut(pos glob.XY, obj *glob.WObject, dir int) {

	/* Don't bother if we don't have outputs */
	if !obj.TypeP.HasMatOutput {
		return
	}

	destObj := util.GetNeighborObj(obj, pos, dir)

	/* Did we find and obj? */
	if destObj == nil {
		return
	}
	/* Does it have inputs? */
	if destObj.TypeP.HasMatInput > 0 {
		return
	}
	/* If we have an output already, unlink it */
	if obj.OutputObj != nil {
		unlinkOut(obj)
	}

	/* Mark target as our output */
	obj.OutputObj = destObj

	/* Make sure the object has an input initialized */
	if destObj.InputBuffer[dir] != nil {
		destObj.InputBuffer[dir] = &glob.MatData{}
	}

	/* Make sure our output is initalized */
	if obj.OutputBuffer == nil {
		obj.OutputBuffer = &glob.MatData{}
	}

	/* Put ourself in target's input list */
	destObj.InputObjs[dir] = obj

}

func linkIn(pos glob.XY, obj *glob.WObject) {

	/* Don't bother if we don't have inputs */
	numInputs := obj.TypeP.HasMatInput
	if numInputs <= 0 {
		return
	}
	for dir := consts.DIR_NORTH; dir < consts.DIR_MAX && numInputs > 0; dir++ {

		/* Don't try to connect an input the same direction as our output */
		if obj.TypeP.HasMatOutput && dir == obj.Direction {
			continue
		}
		neigh := util.GetNeighborObj(obj, pos, dir)

		/* Did we find an object? */
		if neigh != nil {
			/* Does it have an output? */
			if neigh.TypeP.HasMatOutput {
				/* Is the output unoccupied? */
				if neigh.OutputObj == nil {
					/* Don't leave other obj's outputs dangling */
					unlinkInput(obj, dir)

					/* Set ourself as the output */
					neigh.OutputObj = obj

					/* Make sure we have a input */
					if obj.InputBuffer[dir] == nil {
						obj.InputBuffer[dir] = &glob.MatData{}
					}

					/* Record who is on this input */
					obj.InputObjs[dir] = neigh
				}
			}
		}

	}

}

func LinkObj(pos glob.XY, obj *glob.WObject, newdir int) {
	linkIn(pos, obj)
	linkOut(pos, obj, newdir)
}

func makeSuperChunk(pos glob.XY) {
	//Make super chunk if needed

	newPos := pos
	scpos := util.PosToSuperChunkPos(&newPos)

	glob.SuperChunkMapLock.Lock()
	if glob.SuperChunkMap[scpos] == nil {
		glob.SuperChunkMap[scpos] = &glob.MapSuperChunk{}
		glob.SuperChunkMap[scpos].Chunks = make(map[glob.XY]*glob.MapChunk)
	}
	glob.SuperChunkMapLock.Unlock()

}

func MakeChunk(pos glob.XY) {
	//Make chunk if needed

	newPos := pos

	makeSuperChunk(pos)

	cpos := util.PosToChunkPos(&newPos)
	scpos := util.PosToSuperChunkPos(&newPos)

	glob.SuperChunkMap[scpos].NumChunks++

	glob.SuperChunkMapLock.Lock()
	if glob.SuperChunkMap[scpos].Chunks[cpos] == nil {
		glob.SuperChunkMap[scpos].Chunks[cpos] = &glob.MapChunk{}
		glob.SuperChunkMap[scpos].Chunks[cpos].WObject = make(map[glob.XY]*glob.WObject)
	}
	glob.SuperChunkMapLock.Unlock()
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

func CreateObj(pos glob.XY, mtype int, dir int) *glob.WObject {

	//Make chunk if needed
	MakeChunk(pos)
	chunk := util.GetChunk(&pos)
	glob.CameraDirty = true
	obj := chunk.WObject[pos]

	if obj != nil {
		//fmt.Println("Object already exists at:", pos)
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
	}

	cpos := util.PosToChunkPos(&pos)
	scpos := util.PosToSuperChunkPos(&pos)

	glob.SuperChunkMap[scpos].Chunks[cpos].WObject[pos] = obj
	//fmt.Println("Made obj:", pos, obj.TypeP.Name)

	chunk.NumObjects++
	LinkObj(pos, obj, dir)

	return obj
}

func ObjectHitlistAdd(obj *glob.WObject, otype int, pos *glob.XY, delete bool, dir int) {
	ObjectHitlist = append(ObjectHitlist, &glob.ObjectHitlistData{Obj: obj, OType: otype, Pos: pos, Delete: delete, Dir: dir})
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
			EventHitlistAdd(item.Obj, consts.QUEUE_TYPE_TICK, true)
			EventHitlistAdd(item.Obj, consts.QUEUE_TYPE_TOCK, true)

			cpos := util.PosToChunkPos(item.Pos)
			scpos := util.PosToSuperChunkPos(item.Pos)

			/* Remove from chunk */
			glob.SuperChunkMapLock.Lock()
			glob.SuperChunkMap[scpos].Chunks[cpos].NumObjects--
			delete(glob.SuperChunkMap[scpos].Chunks[cpos].WObject, *item.Pos)
			glob.SuperChunkMapLock.Unlock()

		} else {
			//Add
			CreateObj(*item.Pos, item.OType, item.Dir)
		}
	}
	ObjectHitlist = []*glob.ObjectHitlistData{}
}
