package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
	"fmt"
	"image/color"
	"sync"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

var wg sizedwaitgroup.SizedWaitGroup
var GameTick uint64
var GameRunning bool
var GameLock sync.Mutex

/* Loops: Ticks: External, Tocks: Internal, EventQueue, ObjQueue. Locks each list one at a time. Sleeps if needed. Multi-threaded */
func ObjUpdateDaemon() {
	wg = sizedwaitgroup.New(world.NumWorkers)

	for !world.MapGenerated.Load() {
		time.Sleep(time.Millisecond * 100)
	}

	var tockState bool = true

	for GameRunning {
		GameLock.Lock()
		start := time.Now()

		if tockState {
			runTocks()
			tockState = false
		} else {
			runTicks() //Move external
			GameTick++
			tockState = true
		}

		runRotates()

		world.ObjQueueLock.Lock()
		runObjQueue() //Queue to add/remove objects
		world.ObjQueueLock.Unlock()

		world.EventQueueLock.Lock()
		RunEventQueue() //Queue to add/remove events
		world.EventQueueLock.Unlock()

		GameLock.Unlock()

		if !gv.UPSBench {
			sleepFor := time.Duration(world.ObjectUPS_ns) - time.Since(start)
			if sleepFor > minSleep {
				time.Sleep(sleepFor - time.Microsecond)
			}
		}
		world.MeasuredObjectUPS_ns = int(time.Since(start).Nanoseconds())
		world.UPSAvr.Add(1000000000.0 / float64(world.MeasuredObjectUPS_ns) / 2)
	}
}

/* WASM single-thread object update */
func ObjUpdateDaemonST() {
	var start time.Time

	for !world.MapGenerated.Load() {
		time.Sleep(time.Millisecond * 100)
	}

	var tockState bool = true

	for GameRunning {
		GameLock.Lock()

		start = time.Now()

		if tockState {
			runTocksST() //Process objects
			tockState = false
		} else {
			runTicksST() //Move external
			GameTick++
			tockState = true
		}

		runRotates()

		world.ObjQueueLock.Lock()
		runObjQueue() //Queue to add/remove objects
		world.ObjQueueLock.Unlock()

		world.EventQueueLock.Lock()
		RunEventQueue() //Queue to add/remove events
		world.EventQueueLock.Unlock()

		GameLock.Unlock()

		if !gv.UPSBench {
			sleepFor := time.Duration(world.ObjectUPS_ns) - time.Since(start)
			if sleepFor > minSleep {
				time.Sleep(sleepFor - time.Microsecond)
			}
		}
		time.Sleep(time.Millisecond)

		world.MeasuredObjectUPS_ns = int(time.Since(start).Nanoseconds())
		world.UPSAvr.Add(1000000000.0 / float64(world.MeasuredObjectUPS_ns) / 2)
	}
}

/* Put our OutputBuffer to another object's InputBuffer (external)*/
func tickObj(obj *world.ObjData) {

	var blockedOut uint8 = 0
	for _, port := range obj.Outputs {

		/* If we have stuff to send */
		if port.Buf.Amount == 0 {
			continue
		}

		/* If destination is empty */
		if port.Link.Buf.Amount != 0 {
			blockedOut++
			continue
		}

		/* Swap pointers */
		*port.Link.Buf, *port.Buf = *port.Buf, *port.Link.Buf
	}
	for _, port := range obj.FuelOut {

		/* If we have stuff to send */
		if port.Buf.Amount == 0 {
			continue
		}

		//* If destination is empty */
		if port.Link.Buf.Amount != 0 {
			blockedOut++
			continue
		}

		/* Swap pointers */
		*port.Link.Buf, *port.Buf = *port.Buf, *port.Link.Buf
	}

	if obj.NumOut+obj.NumFOut == blockedOut {
		if !obj.Blocked {
			obj.Blocked = true

		}
		if obj.Active {
			obj.Active = false
		}
	} else {
		if obj.Blocked {
			obj.Blocked = false
		}
	}
}

func RotateListAdd(b *world.BuildingData, cw bool, pos world.XY) {
	world.RotateListLock.Lock()

	world.RotateList = append(world.RotateList, world.RotateEvent{Build: b, Clockwise: cw})
	world.RotateCount++

	world.RotateListLock.Unlock()
}

/* Lock and append to TickList */
func ticklistAdd(obj *world.ObjData) {
	if obj.HasTick {
		return
	}
	obj.HasTick = true
	world.TickList = append(world.TickList, world.TickEvent{Target: obj})
	world.TickCount++
}

/* Lock and append to TockList */
func tockListAdd(obj *world.ObjData) {
	if obj.HasTock {
		return
	}
	obj.HasTock = true
	world.TockList = append(world.TockList, world.TickEvent{Target: obj})
	world.TockCount++
}

/* Lock and add it EventQueue */
func EventQueueAdd(obj *world.ObjData, qtype uint8, delete bool) {
	world.EventQueueLock.Lock()
	world.EventQueue = append(world.EventQueue, &world.EventQueueData{Obj: obj, QType: qtype, Delete: delete})
	world.EventQueueLock.Unlock()
}

/* Lock and remove tick event */
func ticklistRemove(obj *world.ObjData) {

	if !obj.HasTick {
		return
	}
	for i, e := range world.TickList {
		if e.Target == obj {
			world.TickList = append(world.TickList[:i], world.TickList[i+1:]...)
			obj.HasTick = false
			obj.Active = false
			world.TickCount--
			return
		}
	}
}

/* lock and remove tock event */
func tocklistRemove(obj *world.ObjData) {

	if !obj.HasTock {
		return
	}
	for i, e := range world.TockList {
		if e.Target == obj {
			world.TockList = append(world.TockList[:i], world.TockList[i+1:]...)
			obj.HasTock = false
			obj.Active = false
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

func runRotates() {
	world.RotateListLock.Lock()
	defer world.RotateListLock.Unlock()

	for _, rot := range world.RotateList {
		var objSave world.ObjData
		b := rot.Build

		if b != nil {

			obj := b.Obj
			if obj.Unique.TypeP.NonSquare {
				var newdir uint8

				/* Save a copy of the object */
				objSave = *obj

				if !rot.Clockwise {
					newdir = util.RotCCW(objSave.Dir)
				} else {
					newdir = util.RotCW(objSave.Dir)
				}
				objSave.Dir = newdir

				/* Remove object from the world */
				removeObj(obj)
				PlaceObj(objSave.Pos, 0, &objSave, newdir, false)
				continue
			}

			var newdir uint8

			UnlinkObj(obj)
			if !rot.Clockwise {
				newdir = util.RotCCW(obj.Dir)
				for p, port := range obj.Ports {
					obj.Ports[p].Dir = util.RotCCW(port.Dir)
				}

				util.ChatDetailed(fmt.Sprintf("Rotated %v counter-clockwise at %v", obj.Unique.TypeP.Name, util.PosToString(obj.Pos)), color.White, time.Second*5)
			} else {
				newdir = util.RotCW(obj.Dir)
				for p, port := range obj.Ports {
					obj.Ports[p].Dir = util.RotCW(port.Dir)
				}

				util.ChatDetailed(fmt.Sprintf("Rotated %v clockwise at %v", obj.Unique.TypeP.Name, util.PosToString(obj.Pos)), color.White, time.Second*5)
			}
			obj.Dir = newdir

			if obj.Unique.TypeP.MultiTile {
				for _, subObj := range obj.Unique.TypeP.SubObjs {
					subPos := util.GetSubPos(obj.Pos, subObj)
					LinkObj(subPos, b)
				}
			} else {
				LinkObj(obj.Pos, b)
			}
		}
	}

	//Done, erase list.
	world.RotateList = []world.RotateEvent{}
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
			if item.Obj.Unique.TypeP.MultiTile {
				for _, sub := range item.Obj.Unique.TypeP.SubObjs {
					tile := RotateCoord(sub, item.Obj.Dir, GetObjSize(item.Obj, nil))
					pos := util.GetSubPos(item.Pos, tile)
					removePosMap(pos)
				}
			}
			delObj(item.Obj)
			world.VisDataDirty.Store(true)
		} else {
			//Add
			PlaceObj(item.Pos, item.OType, nil, item.Dir, false)
			world.VisDataDirty.Store(true)
		}
	}

	world.ObjQueue = []*world.ObjectQueueData{}
}

func delObj(obj *world.ObjData) {
	UnlinkObj(obj)
	removeObj(obj)
}

/* Delete object from ObjMap, decerment NumObjects. Marks PixmapDirty */
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
