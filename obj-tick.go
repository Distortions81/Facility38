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
	gameRunning bool
	gameLock    sync.Mutex
	minSleep    = 400 * time.Microsecond //Sleeping for less than this does not appear effective.
)

func init() {
	defer reportPanic("obj-tick init")
	if strings.EqualFold(runtime.GOOS, "windows") || wasmMode {
		minSleep = (time.Millisecond * 2) //Windows and WASM time resolution sucks
	}
}

/* Loops: Ticks: External, Tocks: Internal, EventQueue, ObjQueue. Locks each list one at a time. Sleeps if needed. Multi-threaded */
func objUpdateDaemon() {
	defer reportPanic("objUpdateDaemon")

	for !mapGenerated.Load() || !authorized.Load() {
		time.Sleep(time.Millisecond * 100)
	}

	var tockState bool = true

	for gameRunning {
		gameLock.Lock()
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

		objQueueLock.Lock()
		runObjQueue() //Queue to add/remove objects
		objQueueLock.Unlock()

		eventQueueLock.Lock()
		runEventQueue() //Queue to add/remove events
		eventQueueLock.Unlock()

		swapAllPortBuffers()

		gameLock.Unlock()

		if !upsBench {
			sleepFor := time.Duration(objectUPS_ns) - time.Since(start)
			if sleepFor > minSleep {
				time.Sleep(sleepFor - time.Microsecond)
			}
		}
		measuredObjectUPS_ns = int(time.Since(start).Nanoseconds())
		actualUPS = (1000000000.0 / float32(measuredObjectUPS_ns))

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

	for !mapGenerated.Load() {
		time.Sleep(time.Millisecond * 100)
	}

	var tockState bool = true

	for gameRunning {
		gameLock.Lock()

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

		objQueueLock.Lock()
		runObjQueue() //Queue to add/remove objects
		objQueueLock.Unlock()

		eventQueueLock.Lock()
		runEventQueue() //Queue to add/remove events
		eventQueueLock.Unlock()

		swapAllPortBuffers()

		gameLock.Unlock()

		if !upsBench {
			sleepFor := time.Duration(objectUPS_ns) - time.Since(start)
			if sleepFor > minSleep {
				time.Sleep(sleepFor - time.Microsecond)
			}
		}

		measuredObjectUPS_ns = int(time.Since(start).Nanoseconds())
		actualUPS = (1000000000.0 / float32(measuredObjectUPS_ns))

		handleAutosave()
	}
}

/* Autosave */
func handleAutosave() {
	if autoSave && !wasmMode && time.Now().UTC().Sub(lastSave) > time.Minute*5 {
		lastSave = time.Now().UTC()
		saveGame()
	}
}

/* Swap active and next buffers for all ports */
func swapAllPortBuffers() {
	superChunkListLock.RLock()
	for _, sChunk := range superChunkList {
		for _, chunk := range sChunk.chunkList {
			chunk.lock.Lock()
			for _, obj := range chunk.objList {
				for p := range obj.Ports {
					obj.Ports[p].Buf, obj.Ports[p].BufNext = obj.Ports[p].BufNext, obj.Ports[p].Buf
					obj.Ports[p].BufNext.Amount = 0
					obj.Ports[p].BufNext.typeP = nil
					obj.Ports[p].BufNext.Rot = 0
				}
			}
			chunk.lock.Unlock()
		}
	}
	superChunkListLock.RUnlock()
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
		if port.link.BufNext.Amount != 0 {
			blockedOut++
			continue
		}

		*port.link.BufNext = *port.Buf
		port.Buf.Amount = 0
		port.Buf.typeP = nil
	}
	for _, port := range obj.fuelOut {

		/* If we have stuff to send */
		if port.Buf.Amount == 0 {
			continue
		}

		/* If destination is empty */
		if port.link.BufNext.Amount != 0 {
			blockedOut++
			continue
		}

		*port.link.BufNext = *port.Buf
		port.Buf.Amount = 0
		port.Buf.typeP = nil
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
	rotateListLock.Lock()

	rotateList = append(rotateList, rotateEventData{build: b, clockwise: cw})

	rotateListLock.Unlock()
}

/* Add to event queue (list of tock and tick events) */
func eventQueueAdd(obj *ObjData, qtype uint8, delete bool) {
	defer reportPanic("EventQueueAdd")
	eventQueueLock.Lock()
	eventQueue = append(eventQueue, &eventQueueData{obj: obj, qType: qtype, delete: delete})
	eventQueueLock.Unlock()
}

/* Add to ObjQueue (add/delete world object at end of tick) */
func objQueueAdd(obj *ObjData, otype uint8, pos XY, delete bool, dir uint8) {
	defer reportPanic("ObjQueueAdd")
	objQueueLock.Lock()
	objQueue = append(objQueue, &objectQueueData{obj: obj, oType: otype, pos: pos, delete: delete, dir: dir})
	objQueueLock.Unlock()
}

/* Perform object rotations between ticks */
func runRotates() {
	defer reportPanic("RunRotates")
	rotateListLock.Lock()
	defer rotateListLock.Unlock()

	for _, rot := range rotateList {
		var objSave ObjData
		b := rot.build

		/* Valid building */
		if b != nil {
			obj := b.obj

			/* Non-square multi-tile objects */
			if obj.Unique.typeP.nonSquare {
				var newDir uint8
				var oldDir uint8 = obj.Dir

				/* Save a copy of the object */
				objSave = *obj

				if !rot.clockwise {
					newDir = RotCCW(objSave.Dir)
				} else {
					newDir = RotCW(objSave.Dir)
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
				found := placeObj(objSave.Pos, 0, &objSave, newDir, false)

				/* Problem found, wont fit, undo! */
				if found == nil {
					/* Unable to rotate, undo */
					chatDetailed(fmt.Sprintf("Unable to rotate: %v at %v", obj.Unique.typeP.name, posToString(obj.Pos)), ColorRed, time.Second*15)
					found = placeObj(objSave.Pos, 0, &objSave, oldDir, false)
					if found == nil {
						chatDetailed(fmt.Sprintf("Unable to place item back: %v at %v", obj.Unique.typeP.name, posToString(obj.Pos)), ColorRed, time.Second*15)
					}
				}
				continue
			}

			var newDir uint8

			/* Unlink */
			unlinkObj(obj)

			/* Rotate ports */
			if !rot.clockwise {
				newDir = RotCCW(obj.Dir)
				for p, port := range obj.Ports {
					obj.Ports[p].Dir = RotCCW(port.Dir)
				}

				chatDetailed(fmt.Sprintf("Rotated %v counter-clockwise at %v", obj.Unique.typeP.name, posToString(obj.Pos)), color.White, time.Second*5)
			} else {
				newDir = RotCW(obj.Dir)
				for p, port := range obj.Ports {
					obj.Ports[p].Dir = RotCW(port.Dir)
				}

				chatDetailed(fmt.Sprintf("Rotated %v clockwise at %v", obj.Unique.typeP.name, posToString(obj.Pos)), color.White, time.Second*5)
			}
			obj.Dir = newDir

			linkObj(obj.Pos, b)
		}
	}

	//Done, reset list.
	rotateList = []rotateEventData{}
}

/* Add/remove tick/tock events from the lists */
func runEventQueue() {
	defer reportPanic("RunEventQueue")

	for _, e := range eventQueue {
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
	eventQueue = []*eventQueueData{}
}

/* Add/remove objects from game world at end of tick/tock cycle */
func runObjQueue() {
	defer reportPanic("runObjQueue")

	for _, item := range objQueue {
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
			visDataDirty.Store(true)
		} else {

			/* Place object in world */
			//Add
			placeObj(item.pos, item.oType, nil, item.dir, false)
			visDataDirty.Store(true)
		}
	}

	/* Done, reset list */
	objQueue = []*objectQueueData{}
}
