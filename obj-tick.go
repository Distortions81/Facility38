package main

import (
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
	defer reportPanic("obj-tick init")
	if strings.EqualFold(runtime.GOOS, "windows") || WASMMode {
		minSleep = (time.Millisecond * 2) //Windows and WASM time resolution sucks
	}
}

/* Loops: Ticks: External, Tocks: Internal, EventQueue, ObjQueue. Locks each list one at a time. Sleeps if needed. Multi-threaded */
func objUpdateDaemon() {
	defer reportPanic("objUpdateDaemon")

	for !MapGenerated.Load() {
		time.Sleep(time.Millisecond * 100)
	}

	var tockState bool = true

	for GameRunning {
		GameLock.Lock()
		start := time.Now()

		if tockState {
			newRunTocks()
			tockState = false
		} else {
			newRunTicks() //Move external
			GameTick++
			tockState = true
		}

		runRotates()

		ObjQueueLock.Lock()
		runObjQueue() //Queue to add/remove objects
		ObjQueueLock.Unlock()

		EventQueueLock.Lock()
		RunEventQueue() //Queue to add/remove events
		EventQueueLock.Unlock()

		GameLock.Unlock()

		if !UPSBench {
			sleepFor := time.Duration(ObjectUPS_ns) - time.Since(start)
			if sleepFor > minSleep {
				time.Sleep(sleepFor - time.Microsecond)
			}
		}
		MeasuredObjectUPS_ns = int(time.Since(start).Nanoseconds())
		ActualUPS = (1000000000.0 / float32(MeasuredObjectUPS_ns))

		handleAutosave()
	}
}

/*
 * WASM single-thread object update
 * Simple, but does not run well with large worlds
 * This really needs to process a bit at a time at the end of each frame instead
 * This works, but gets very 'jerky' with large maps due to uneven process load
 * Leading to wildly varying frame times
 * WASM just ins't a high priority at the moment
 */
func ObjUpdateDaemonST() {
	defer reportPanic("ObjUpdateDaemonST")
	var start time.Time

	for !MapGenerated.Load() {
		time.Sleep(time.Millisecond * 100)
	}

	var tockState bool = true

	for GameRunning {
		GameLock.Lock()

		start = time.Now()

		if tockState {
			newRunTocksST() //Process objects
			tockState = false
		} else {
			newRunTicksST() //Move external
			GameTick++
			tockState = true
		}

		runRotates()

		ObjQueueLock.Lock()
		runObjQueue() //Queue to add/remove objects
		ObjQueueLock.Unlock()

		EventQueueLock.Lock()
		RunEventQueue() //Queue to add/remove events
		EventQueueLock.Unlock()

		GameLock.Unlock()

		if !UPSBench {
			sleepFor := time.Duration(ObjectUPS_ns) - time.Since(start)
			if sleepFor > minSleep {
				time.Sleep(sleepFor - time.Microsecond)
			}
		}

		MeasuredObjectUPS_ns = int(time.Since(start).Nanoseconds())
		ActualUPS = (1000000000.0 / float32(MeasuredObjectUPS_ns))

		handleAutosave()
	}
}

/* Autosave */
func handleAutosave() {
	if Autosave && !WASMMode && time.Since(LastSave) > time.Minute*5 {
		LastSave = time.Now().UTC()
		saveGame()
	}
}

/* Put our OutputBuffer to another object's InputBuffer (external)*/
func tickObj(obj *ObjData) {
	defer reportPanic("tickObj")

	var blockedOut uint8 = 0
	for _, port := range obj.outputs {

		/* If we have stuff to send */
		if port.Buf.Amount == 0 {
			continue
		}

		/* If destination is empty */
		if port.link.Buf.Amount != 0 {
			blockedOut++
			continue
		}

		/* Swap pointers */
		*port.link.Buf, *port.Buf = *port.Buf, *port.link.Buf
	}
	for _, port := range obj.fuelOut {

		/* If we have stuff to send */
		if port.Buf.Amount == 0 {
			continue
		}

		/* If destination is empty */
		if port.link.Buf.Amount != 0 {
			blockedOut++
			continue
		}

		/* Swap pointers */
		*port.link.Buf, *port.Buf = *port.Buf, *port.link.Buf
	}

	/* Don't bother with blocking except on belts */
	if obj.Unique.typeP.category == objCatBelt {
		if obj.numOut+obj.numFOut == blockedOut {
			if !obj.blocked {
				obj.blocked = true

			}
		} else {
			if obj.blocked {
				obj.blocked = false
			}
		}
	}
}

/* A queue of object rotations to perform between ticks */
func rotateListAdd(b *buildingData, cw bool, pos XY) {
	defer reportPanic("RotateListAdd")
	RotateListLock.Lock()

	RotateList = append(RotateList, rotateEvent{build: b, clockwise: cw})
	RotateCount++

	RotateListLock.Unlock()
}

/* Add to event queue (list of tock and tick events) */
func EventQueueAdd(obj *ObjData, qtype uint8, delete bool) {
	defer reportPanic("EventQueueAdd")
	EventQueueLock.Lock()
	EventQueue = append(EventQueue, &eventQueueData{obj: obj, qType: qtype, delete: delete})
	EventQueueLock.Unlock()
}

/* Add to ObjQueue (add/delete world object at end of tick) */
func objQueueAdd(obj *ObjData, otype uint8, pos XY, delete bool, dir uint8) {
	defer reportPanic("ObjQueueAdd")
	ObjQueueLock.Lock()
	ObjQueue = append(ObjQueue, &objectQueueData{obj: obj, oType: otype, pos: pos, delete: delete, dir: dir})
	ObjQueueLock.Unlock()
}

/* Perform object rotations between ticks */
func runRotates() {
	defer reportPanic("RunRotates")
	RotateListLock.Lock()
	defer RotateListLock.Unlock()

	for _, rot := range RotateList {
		var objSave ObjData
		b := rot.build

		/* Valid building */
		if b != nil {
			obj := b.obj

			/* Non-square multi-tile objects */
			if obj.Unique.typeP.nonSquare {
				var newdir uint8
				var olddir uint8 = obj.Dir

				/* Save a copy of the object */
				objSave = *obj

				if !rot.clockwise {
					newdir = RotCCW(objSave.Dir)
				} else {
					newdir = RotCW(objSave.Dir)
				}

				/* Remove object from the world */
				unlinkObj(obj)
				removeObj(obj)

				/* Rotate sub-object map */
				for _, sub := range objSave.Unique.typeP.subObjs {
					tile := rotateCoord(sub, objSave.Dir, getObjSize(&objSave, nil))
					pos := GetSubPos(objSave.Pos, tile)
					removePosMap(pos)
				}

				/* Place back into world */
				found := placeObj(objSave.Pos, 0, &objSave, newdir, false)

				/* Problem found, wont fit, undo! */
				if found == nil {
					/* Unable to rotate, undo */
					ChatDetailed(fmt.Sprintf("Unable to rotate: %v at %v", obj.Unique.typeP.name, PosToString(obj.Pos)), ColorRed, time.Second*15)
					found = placeObj(objSave.Pos, 0, &objSave, olddir, false)
					if found == nil {
						ChatDetailed(fmt.Sprintf("Unable to place item back: %v at %v", obj.Unique.typeP.name, PosToString(obj.Pos)), ColorRed, time.Second*15)
					}
				}
				continue
			}

			var newdir uint8

			/* Unlink */
			unlinkObj(obj)

			/* Rotate ports */
			if !rot.clockwise {
				newdir = RotCCW(obj.Dir)
				for p, port := range obj.Ports {
					obj.Ports[p].Dir = RotCCW(port.Dir)
				}

				ChatDetailed(fmt.Sprintf("Rotated %v counter-clockwise at %v", obj.Unique.typeP.name, PosToString(obj.Pos)), color.White, time.Second*5)
			} else {
				newdir = RotCW(obj.Dir)
				for p, port := range obj.Ports {
					obj.Ports[p].Dir = RotCW(port.Dir)
				}

				ChatDetailed(fmt.Sprintf("Rotated %v clockwise at %v", obj.Unique.typeP.name, PosToString(obj.Pos)), color.White, time.Second*5)
			}
			obj.Dir = newdir

			/* TODO: move this code to LinkObj */
			/* multi-tile object relink */
			if obj.Unique.typeP.multiTile {
				for _, subObj := range obj.Unique.typeP.subObjs {
					subPos := GetSubPos(obj.Pos, subObj)
					linkObj(subPos, b)
				}
			} else {
				/* Standard relink */
				linkObj(obj.Pos, b)
			}
		}
	}

	//Done, reset list.
	RotateList = []rotateEvent{}
}

/* Add/remove tick/tock events from the lists */
func RunEventQueue() {
	defer reportPanic("RunEventQueue")

	for _, e := range EventQueue {
		if e.delete {
			switch e.qType {
			case QUEUE_TYPE_TICK:
				removeTick(e.obj)
			case QUEUE_TYPE_TOCK:
				removeTock(e.obj)
			}
		} else {
			switch e.qType {
			case QUEUE_TYPE_TICK:
				addTick(e.obj)
			case QUEUE_TYPE_TOCK:
				addTock(e.obj)
			}
		}

	}

	/* Done, reset list */
	EventQueue = []*eventQueueData{}
}

/* Add/remove objects from game world at end of tick/tock cycle */
func runObjQueue() {
	defer reportPanic("runObjQueue")

	for _, item := range ObjQueue {
		if item.delete {

			/* Handle multi-tile item delete */
			if item.obj.Unique.typeP.multiTile {
				for _, sub := range item.obj.Unique.typeP.subObjs {
					tile := rotateCoord(sub, item.obj.Dir, getObjSize(item.obj, nil))
					pos := GetSubPos(item.pos, tile)
					removePosMap(pos)
				}
			}
			delObj(item.obj)
			VisDataDirty.Store(true)
		} else {

			/* Place object in world */
			//Add
			placeObj(item.pos, item.oType, nil, item.dir, false)
			VisDataDirty.Store(true)
		}
	}

	/* Done, reset list */
	ObjQueue = []*objectQueueData{}
}

/* Unlink and remove object */
func delObj(obj *ObjData) {
	defer reportPanic("delObj")
	unlinkObj(obj)
	removeObj(obj)
}

/* TODO this and obj-go are duplicates and need to be consolidated */
/* Delete object from ObjMap, decerment Num Marks PixmapDirty */
func removePosMap(pos XY) {
	defer reportPanic("removePosMap")

	sChunk := GetSuperChunk(pos)
	chunk := GetChunk(pos)
	if chunk == nil || sChunk == nil {
		return
	}

	chunk.lock.Lock()
	chunk.numObjs--
	delete(chunk.buildingMap, pos)
	sChunk.pixmapDirty = true
	chunk.lock.Unlock()
}
