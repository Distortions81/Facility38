package main

import (
	"Facility38/def"
	"Facility38/util"
	"Facility38/world"
	"fmt"
	"image/color"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	GameTick    uint64
	GameRunning bool
	GameLock    sync.Mutex
	minSleep    = 400 * time.Microsecond //Sleeping for less than this does not appear effective.
)

func init() {
	defer util.ReportPanic("obj-tick init")
	if strings.EqualFold(runtime.GOOS, "windows") || world.WASMMode {
		minSleep = (time.Millisecond * 2) //Windows and WASM time resolution sucks
	}
}

/* Loops: Ticks: External, Tocks: Internal, EventQueue, ObjQueue. Locks each list one at a time. Sleeps if needed. Multi-threaded */
func ObjUpdateDaemon() {
	defer util.ReportPanic("ObjUpdateDaemon")

	for !world.MapGenerated.Load() {
		time.Sleep(time.Millisecond * 100)
	}

	var tockState bool = true

	for GameRunning {
		GameLock.Lock()
		start := time.Now()

		if tockState {
			NewRunTocks()
			tockState = false
		} else {
			NewRunTicks() //Move external
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

		if !world.UPSBench {
			sleepFor := time.Duration(world.ObjectUPS_ns) - time.Since(start)
			if sleepFor > minSleep {
				time.Sleep(sleepFor - time.Microsecond)
			}
		}
		world.MeasuredObjectUPS_ns = int(time.Since(start).Nanoseconds())
		world.ActualUPS = (1000000000.0 / float32(world.MeasuredObjectUPS_ns))

		handleAutosave()
	}
}

/* WASM single-thread object update */
func ObjUpdateDaemonST() {
	defer util.ReportPanic("ObjUpdateDaemonST")
	var start time.Time

	for !world.MapGenerated.Load() {
		time.Sleep(time.Millisecond * 100)
	}

	var tockState bool = true

	for GameRunning {
		GameLock.Lock()

		start = time.Now()

		if tockState {
			NewRunTocksST() //Process objects
			tockState = false
		} else {
			NewRunTicksST() //Move external
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

		if !world.UPSBench {
			sleepFor := time.Duration(world.ObjectUPS_ns) - time.Since(start)
			if sleepFor > minSleep {
				time.Sleep(sleepFor - time.Microsecond)
			}
		}

		world.MeasuredObjectUPS_ns = int(time.Since(start).Nanoseconds())
		world.ActualUPS = (1000000000.0 / float32(world.MeasuredObjectUPS_ns))

		handleAutosave()
	}
}

func handleAutosave() {
	if !world.WASMMode && world.Autosave && time.Since(world.LastSave) > time.Minute*5 {
		world.LastSave = time.Now().UTC()
		SaveGame()
	}
}

/* Put our OutputBuffer to another object's InputBuffer (external)*/
func tickObj(obj *world.ObjData) {
	defer util.ReportPanic("tickObj")

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

		/* If destination is empty */
		if port.Link.Buf.Amount != 0 {
			blockedOut++
			continue
		}

		/* Swap pointers */
		*port.Link.Buf, *port.Buf = *port.Buf, *port.Link.Buf
	}

	if obj.Unique.TypeP.Category == def.ObjCatBelt {
		if obj.NumOut+obj.NumFOut == blockedOut {
			if !obj.Blocked {
				obj.Blocked = true

			}
		} else {
			if obj.Blocked {
				obj.Blocked = false
			}
		}
	}
}

func RotateListAdd(b *world.BuildingData, cw bool, pos world.XY) {
	defer util.ReportPanic("RotateListAdd")
	world.RotateListLock.Lock()

	world.RotateList = append(world.RotateList, world.RotateEvent{Build: b, Clockwise: cw})
	world.RotateCount++

	world.RotateListLock.Unlock()
}

/* Lock and add it EventQueue */
func EventQueueAdd(obj *world.ObjData, qtype uint8, delete bool) {
	defer util.ReportPanic("EventQueueAdd")
	world.EventQueueLock.Lock()
	world.EventQueue = append(world.EventQueue, &world.EventQueueData{Obj: obj, QType: qtype, Delete: delete})
	world.EventQueueLock.Unlock()
}

/* Add to ObjQueue (add/delete world object at end of tick) */
func ObjQueueAdd(obj *world.ObjData, otype uint8, pos world.XY, delete bool, dir uint8) {
	defer util.ReportPanic("ObjQueueAdd")
	world.ObjQueueLock.Lock()
	world.ObjQueue = append(world.ObjQueue, &world.ObjectQueueData{Obj: obj, OType: otype, Pos: pos, Delete: delete, Dir: dir})
	world.ObjQueueLock.Unlock()
}

func runRotates() {
	defer util.ReportPanic("RunRotates")
	world.RotateListLock.Lock()
	defer world.RotateListLock.Unlock()

	for _, rot := range world.RotateList {
		var objSave world.ObjData
		b := rot.Build

		if b != nil {
			obj := b.Obj
			CleanPorts(obj)

			if obj.Unique.TypeP.NonSquare {
				var newdir uint8
				var olddir uint8 = obj.Dir

				/* Save a copy of the object */
				objSave = *obj

				if !rot.Clockwise {
					newdir = util.RotCCW(objSave.Dir)
				} else {
					newdir = util.RotCW(objSave.Dir)
				}

				/* Remove object from the world */
				UnlinkObj(obj)
				removeObj(obj)
				for _, sub := range objSave.Unique.TypeP.SubObjs {
					tile := RotateCoord(sub, objSave.Dir, GetObjSize(&objSave, nil))
					pos := util.GetSubPos(objSave.Pos, tile)
					removePosMap(pos)
				}
				found := PlaceObj(objSave.Pos, 0, &objSave, newdir, false)
				if found == nil {
					/* Unable to rotate, undo */
					util.ChatDetailed(fmt.Sprintf("Unable to rotate: %v at %v", obj.Unique.TypeP.Name, util.PosToString(obj.Pos)), world.ColorRed, time.Second*15)
					found = PlaceObj(objSave.Pos, 0, &objSave, olddir, false)
					if found == nil {
						util.ChatDetailed(fmt.Sprintf("Unable to place item back: %v at %v", obj.Unique.TypeP.Name, util.PosToString(obj.Pos)), world.ColorRed, time.Second*15)
					}
				}
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

/* TO DO: process if possible, move, or spill */
func CleanPorts(obj *world.ObjData) {
	defer util.ReportPanic("CleanPorts")

	for p, port := range obj.Ports {
		if port.Buf != nil && port.Buf.Amount > 0 {
			obj.Ports[p].Buf.Amount = 0
		}
	}
}

/* Add/remove tick/tock events from the lists
 */
func RunEventQueue() {
	defer util.ReportPanic("RunEventQueue")

	for _, e := range world.EventQueue {
		if e.Delete {
			switch e.QType {
			case def.QUEUE_TYPE_TICK:
				RemoveTick(e.Obj)
			case def.QUEUE_TYPE_TOCK:
				RemoveTock(e.Obj)
			}
		} else {
			switch e.QType {
			case def.QUEUE_TYPE_TICK:
				AddTick(e.Obj)
			case def.QUEUE_TYPE_TOCK:
				AddTock(e.Obj)
			}
		}

	}

	world.EventQueue = []*world.EventQueueData{}
}

/* Add/remove objects from game world at end of tick/tock cycle */
func runObjQueue() {
	defer util.ReportPanic("runObjQueue")

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
	defer util.ReportPanic("delObj")
	UnlinkObj(obj)
	removeObj(obj)
}

/* Delete object from ObjMap, decerment Num Marks PixmapDirty */
func removePosMap(pos world.XY) {
	defer util.ReportPanic("removePosMap")

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
