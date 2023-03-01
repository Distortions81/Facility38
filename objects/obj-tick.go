package objects

import (
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

	for _, port := range obj.Outputs {

		/* If there is a link */
		if port.Link == nil {
			continue
		}

		/* If we have stuff to send */
		if port.Buf.Amount == 0 {
			continue
		}

		/* If destination is empty */
		if port.Link.Buf.Amount != 0 {
			continue
		}

		swapPortBuf(port.Link.Buf, port.Buf)
	}
}

/* WASM single thread: Put our OutputBuffer to another object's InputBuffer (external)*/
func runTicksST() {
	if world.TickCount == 0 {
		return
	}
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
	if world.TockCount == 0 {
		return
	}

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
	if world.TockCount == 0 {
		return
	}

	for _, item := range world.TockList {
		item.Target.TypeP.UpdateObj(item.Target)
	}
}

/* Lock and append to TickList */
func ticklistAdd(obj *world.ObjData) {

	world.TickList = append(world.TickList, world.TickEvent{Target: obj})
	world.TickCount++
}

/* Lock and append to TockList */
func tockListAdd(obj *world.ObjData) {

	world.TockList = append(world.TockList, world.TickEvent{Target: obj})
	world.TockCount++
}

/* Lock and add it EventQueue */
func EventQueueAdd(obj *world.ObjData, qtype uint8, delete bool) {

	world.EventQueue = append(world.EventQueue, &world.EventQueueData{Obj: obj, QType: qtype, Delete: delete})
}

/* Lock and remove tick event */
func ticklistRemove(obj *world.ObjData) {

	for i, e := range world.TickList {
		if e.Target == obj {
			world.TickList = append(world.TickList[:i], world.TickList[i+1:]...)
			world.TickCount--
			return
		}
	}

}

/* lock and remove tock event */
func tocklistRemove(obj *world.ObjData) {
	world.TockListLock.Lock()
	defer world.TockListLock.Unlock()

	for i, e := range world.TockList {
		if e.Target == obj {
			world.TockList = append(world.TockList[:i], world.TockList[i+1:]...)
			world.TockCount--
			return
		}
	}
}

/* Add to ObjQueue (add/delete world object at end of tick) */
func ObjQueueAdd(obj *world.ObjData, otype uint8, pos world.XY, delete bool, dir uint8) {
	world.ObjQueueLock.Lock()
	world.ObjQueue = append(world.ObjQueue, &world.ObjectQueueData{Obj: obj, OType: otype, Pos: pos, Delete: delete, Dir: dir})
	world.ObjQueueLock.Unlock()
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
			if item.Obj.TypeP.SubObjs != nil {
				for _, sub := range item.Obj.TypeP.SubObjs {
					pos := util.AddXY(sub, item.Pos)
					removePosMap(pos)
				}
			}
			delObj(item.Obj)
		} else {
			//Add
			CreateObj(item.Pos, item.OType, item.Dir)
		}
	}

	world.ObjQueue = []*world.ObjectQueueData{}
}

func delObj(obj *world.ObjData) {
	UnlinkObj(obj)

	/* Remove tick and tock events */
	EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, true)
	EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, true)

	removeObj(obj)
}

/* Delete object from ObjMap, ObjList, decerment NumObjects. Marks PixmapDirty */
func removePosMap(pos world.XY) {
	/* delete from map */
	sChunk := util.GetSuperChunk(pos)
	chunk := util.GetChunk(pos)
	if chunk == nil || sChunk == nil {
		return
	}
	chunk.Lock.Lock()
	chunk.NumObjs--
	delete(chunk.BuildingMap, pos)
	chunk.Lock.Unlock()

	sChunk.PixmapDirty = true
}
