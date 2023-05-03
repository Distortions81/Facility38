package main

import (
	"github.com/remeh/sizedwaitgroup"
)

/* Modulo offset */
type offsetData struct {
	offset int
	ticks  []*ObjData
	tocks  []*ObjData
}

/* How often objects in this group update */
type tickInterval struct {
	interval   int
	lastOffset int
	offsets    []offsetData
}

var (
	tickIntervals []tickInterval
	wg            sizedwaitgroup.SizedWaitGroup

	activeTicks int
	activeTocks int

	tickBlocks int
	tockBlocks int
	block      [workSize]*ObjData
)

/* Init at boot */
func tickInit() {
	defer reportPanic("tickInit")
	for _, ot := range worldObjs {
		getIntervalPos(int(ot.tockInterval))
	}
	DoLog(true, "%v intervals added.", len(tickIntervals))

}

/* Return interval data, or create it if needed */
func getIntervalPos(interval int) (pos int, created bool) {
	defer reportPanic("getIntervalPos")
	foundInterval := false

	/* Eventually replace with precalc table */
	for ipos, inter := range tickIntervals {
		if inter.interval == interval {
			foundInterval = true
			return ipos, false
		}
	}
	/* Doesn't exist, create it */
	if !foundInterval {
		pos := len(tickIntervals)

		offsets := make([]offsetData, interval+1)
		for opos := range offsets {
			offsets[opos].offset = opos
		}
		tickIntervals = append(tickIntervals, tickInterval{interval: interval, offsets: offsets})
		return pos, true
	}

	DoLog(true, "Error!")
	return -1, false
}

/* Add tock event to tick interval/mod offset */
func addTock(obj *ObjData) {
	defer reportPanic("addTock")
	if obj.hasTock {
		return
	}
	i, _ := getIntervalPos(int(obj.Unique.typeP.tockInterval))

	if tickIntervals[i].lastOffset >= tickIntervals[i].interval {
		tickIntervals[i].lastOffset = 0
	}

	tickIntervals[i].offsets[tickIntervals[i].lastOffset].tocks =
		append(tickIntervals[i].offsets[tickIntervals[i].lastOffset].tocks, obj)

	tickIntervals[i].lastOffset++
	TockCount++
}

/* Remove a tock from tick interval/mod offset */
func removeTock(obj *ObjData) {
	defer reportPanic("removeTock")
	if !obj.hasTock {
		return
	}
	i, _ := getIntervalPos(int(obj.Unique.typeP.tockInterval))

	for offPos, off := range tickIntervals[i].offsets {
		/* Check if this is the correct interval */
		if uint8(off.offset) != obj.Unique.typeP.tockInterval {
			continue
		}
		/* If it is, remove object */
		for itemPos, item := range off.tocks {
			if item == obj {

				tickIntervals[i].offsets[offPos].tocks =
					append(
						tickIntervals[i].offsets[offPos].tocks[:itemPos],
						tickIntervals[i].offsets[offPos].tocks[itemPos+1:]...)

				TockCount--
				DoLog(true, "Tock Removed: %v", obj.Unique.typeP.name)
				break
			}
		}
	}
}

/* Add a tick from a tick interval/mod offset */
func addTick(obj *ObjData) {
	defer reportPanic("addTick")
	if obj.hasTick {
		return
	}

	i, _ := getIntervalPos(int(obj.Unique.typeP.tockInterval))
	if tickIntervals[i].lastOffset >= tickIntervals[i].interval {
		tickIntervals[i].lastOffset = 0
	}

	tickIntervals[i].offsets[tickIntervals[i].lastOffset].ticks =
		append(tickIntervals[i].offsets[tickIntervals[i].lastOffset].ticks, obj)

	tickIntervals[i].lastOffset++
	TickCount++
}

/* Remove a tick event from the TickInterval list */
func removeTick(obj *ObjData) {
	defer reportPanic("removeTick")

	if !obj.hasTick {
		return
	}
	/* Find our position */
	i, _ := getIntervalPos(int(obj.Unique.typeP.tockInterval))

	for offPos, off := range tickIntervals[i].offsets {
		/* Check if this is the correct interval */
		if uint8(off.offset) != obj.Unique.typeP.tockInterval {
			continue
		}
		/* If it is, remove object */
		for itemPos, item := range off.ticks {
			if item == obj {

				tickIntervals[i].offsets[offPos].ticks =
					append(
						tickIntervals[i].offsets[offPos].ticks[:itemPos],
						tickIntervals[i].offsets[offPos].ticks[itemPos+1:]...)

				TickCount--
				DoLog(true, "Tick Removed: %v", obj.Unique.typeP.name)
				break
			}
		}
	}
}

/* Single-thhread run tocks */
func newRunTocksST() {
	defer reportPanic("newRunTocksST")
	activeTocks = 0

	for _, ti := range tickIntervals {
		for _, off := range ti.offsets {
			if ti.interval == 0 || (GameTick+uint64(off.offset))%uint64(ti.interval) == 0 {
				for _, tock := range off.tocks {
					tock.Unique.typeP.updateObj(tock)
				}
				activeTocks += len(off.tocks)
			}
		}
	}

	ActiveTockCount = activeTocks
}

/* Single-thread run ticks */
func newRunTicksST() {
	defer reportPanic("newRunTicksST")
	activeTicks = 0

	for _, ti := range tickIntervals {
		for _, off := range ti.offsets {
			if ti.interval == 0 || (GameTick+uint64(off.offset))%uint64(ti.interval) == 0 {
				for _, tock := range off.ticks {
					tickObj(tock)
				}
				activeTicks += len(off.tocks)
			}
		}
	}

	ActiveTickCount = activeTicks
}

/* Threaded tock update */
func newRunTocks() {
	defer reportPanic("newRunTocks")

	numObj := 0
	activeTocks = 0
	tockBlocks = 0

	wg = sizedwaitgroup.New(NumWorkers)

	for _, ti := range tickIntervals {
		for _, off := range ti.offsets {
			if ti.interval == 0 || (GameTick+uint64(off.offset))%uint64(ti.interval) == 0 {
				for _, tock := range off.tocks {
					block[numObj] = tock
					numObj++
					if numObj == workSize {
						//Waitgroup add and done happen within here
						runTockBlock(numObj)
						activeTocks += numObj
						tockBlocks++
						numObj = 0
					}
				}
			}
		}
	}
	if numObj > 0 {
		runTockBlock(numObj)
		activeTocks += numObj
	}
	wg.Wait()
	ActiveTockCount = activeTocks
}

/* Threaded tick update */
func newRunTicks() {

	numObj := 0
	activeTicks = 0
	tickBlocks = 0
	defer reportPanic("newRunTicks")

	for _, ti := range tickIntervals {
		for _, off := range ti.offsets {
			if ti.interval == 0 || (GameTick+uint64(off.offset))%uint64(ti.interval) == 0 {
				for _, tick := range off.ticks {
					block[numObj] = tick
					numObj++
					if numObj == workSize {
						//Waitgroup add and done happen within here
						runTickBlock(numObj)
						activeTicks += numObj
						tickBlocks++
						numObj = 0
					}
				}
			}
		}
	}
	if numObj > 0 {
		runTickBlock(numObj)
		activeTicks += numObj
	}
	wg.Wait()

	ActiveTickCount = activeTicks
}

/* Run a block of ticks on a thread */
func runTickBlock(numObj int) {
	defer reportPanic("runTickBlock")

	wg.Add()
	go func(w [workSize]*ObjData, nObj int) {
		defer reportPanic("runTickBlock goroutine")
		for x := 0; x < nObj; x++ {
			tickObj(w[x])
		}
		wg.Done()
	}(block, numObj)

	block = [workSize]*ObjData{}
}

/* Run a block of tocks on a thread */
func runTockBlock(numObj int) {
	defer reportPanic("runTockBlock")
	wg.Add()
	go func(w [workSize]*ObjData, nObj int) {
		var x int
		defer reportPanic("runTockBlock goroutine")

		for x = 0; x < nObj; x++ {
			w[x].Unique.typeP.updateObj(w[x])
		}
		wg.Done()
	}(block, numObj)

	block = [workSize]*ObjData{}
}
