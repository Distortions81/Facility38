package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

var wg sizedwaitgroup.SizedWaitGroup
var GameTick uint64

/* Loops: Ticks: External, Tocks: Internal, EventQueue, ObjQueue. Locks each list one at a time. Sleeps if needed. Multi-threaded */
func ObjUpdateDaemon() {
	wg = sizedwaitgroup.New(world.NumWorkers)

	for !world.MapGenerated.Load() {
		time.Sleep(time.Millisecond * 10)
	}

	var tockState bool = true
	for {
		if tockState {
			runTocks()
			tockState = false
		} else {
			runTicks() //Move external
			GameTick++
			tockState = true
		}

		world.ObjQueueLock.Lock()
		runObjQueue() //Queue to add/remove objects
		world.ObjQueueLock.Unlock()

		world.EventQueueLock.Lock()
		RunEventQueue() //Queue to add/remove events
		world.EventQueueLock.Unlock()
	}
}

/* WASM single-thread object update */
func ObjUpdateDaemonST() {
	var start time.Time

	var tockState bool = true
	for {
		start = time.Now()

		if tockState {
			world.TockListLock.Lock()
			runTocksST() //Process objects
			world.TockListLock.Unlock()
			tockState = false
		} else {
			world.TickListLock.Lock()
			runTicksST() //Move external
			GameTick++
			world.TickListLock.Unlock()
			tockState = true
		}

		world.ObjQueueLock.Lock()
		runObjQueue() //Queue to add/remove objects
		world.ObjQueueLock.Unlock()

		world.EventQueueLock.Lock()
		RunEventQueue() //Queue to add/remove events
		world.EventQueueLock.Unlock()

		if !gv.UPSBench {
			sleepFor := time.Duration(world.ObjectUPS_ns) - time.Since(start)
			time.Sleep(sleepFor)
		}
		world.MeasuredObjectUPS_ns = int(time.Since(start).Nanoseconds())
	}
}

/* Put our OutputBuffer to another object's InputBuffer (external)*/
func tickObj(obj *world.ObjData) {

	for _, port := range obj.Outputs {

		/* If we have stuff to send */
		if port.Buf.Amount == 0 {
			continue
		}

		/* If destination is empty */
		if port.Link.Buf.Amount != 0 {
			continue
		}

		/* Swap pointers */
		*port.Link.Buf, *port.Buf = *port.Buf, *port.Link.Buf
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

/* Lock and append to TickList */
func ticklistAdd(obj *world.ObjData) {
	if !obj.HasTick {
		obj.HasTick = true
		world.TickList = append(world.TickList, world.TickEvent{Target: obj})
		world.TickCount++
	}
}

/* Lock and append to TockList */
func tockListAdd(obj *world.ObjData) {
	if !obj.HasTock {
		obj.HasTock = true

		world.TockList = append(world.TockList, world.TickEvent{Target: obj})
		world.TockCount++
	}
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
			obj.HasTick = false
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
			obj.HasTock = false
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
func RunEventQueue() {

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
			if item.Obj.TypeP.Size.X > 1 || item.Obj.TypeP.Size.Y > 1 {
				for _, sub := range item.Obj.TypeP.SubObjs {
					pos := util.AddXY(sub, item.Pos)
					removePosMap(pos)
				}
			}
			delObj(item.Obj)
		} else {
			//Add
			CreateObj(item.Pos, item.OType, item.Dir, false)
		}
	}

	world.ObjQueue = []*world.ObjectQueueData{}
}

func delObj(obj *world.ObjData) {
	UnlinkObj(obj)
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
	sChunk.PixmapDirty = true
	chunk.Lock.Unlock()

}
