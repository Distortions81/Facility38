package objects

import (
	"GameTest/gv"
	"GameTest/world"
	"math"
	"runtime"
	"strings"
	"time"
)

var (
	minWorkSize = 1000
	margin      = 1.8

	largeThreshold  = minWorkSize
	minSleep        = 200 * time.Microsecond //Sleeping for less than this does not appear effective.
	BlocksPerWorker = 10
)

func init() {
	if strings.EqualFold(runtime.GOOS, "windows") || gv.WASMMode {
		minSleep = time.Millisecond //Windows time resolution sucks
	}
}

/* Process internally in an object, multi-threaded*/
func runTicks() {
	if world.TickCount == 0 {
		time.Sleep(time.Millisecond)
		return
	}
	var lastTick int = 0
	var sleepFor time.Duration
	var maxBlocks = world.NumWorkers * BlocksPerWorker
	wSize := minWorkSize

	if world.TickCount > largeThreshold {
		wSize = int(math.Ceil(float64(world.TickCount)/float64(maxBlocks))) + minWorkSize
	}

	world.TickListLock.Lock()
	for {

		startTime := time.Now()

		/* If worksize is larger than remaining work, adjust worksize */
		if lastTick+wSize > world.TickCount {
			wSize = world.TickCount - lastTick
		}

		wg.Add()
		go func(wSize, lastTick int) {
			for i := lastTick; i < lastTick+wSize; i++ {
				tickObj(world.TickList[i].Target)
			}
			wg.Done()
		}(wSize, lastTick)

		lastTick = lastTick + wSize + 1

		if lastTick >= world.TickCount {
			break
		}

		if !gv.UPSBench {
			sleepFor = time.Duration(world.ObjectUPS_ns/int(float64(world.TickCount)/(float64(wSize)/margin))) - time.Since(startTime)
			if sleepFor > minSleep {
				time.Sleep(sleepFor)
			}
		}
	}
	wg.Wait()
	world.TickListLock.Unlock()

	//fmt.Printf("TICK: sleep-per: %v, workSize: %v\n", sleepFor.String(), wSize)

}

func runTocks() {
	if world.TockCount == 0 {
		time.Sleep(time.Millisecond)
		return
	}
	var lastTock int = 0
	var sleepFor time.Duration
	var maxBlocks = world.NumWorkers * BlocksPerWorker
	wSize := minWorkSize

	if world.TockCount > largeThreshold {
		wSize = int(math.Ceil(float64(world.TockCount)/float64(maxBlocks))) + minWorkSize
	}

	world.TockListLock.Lock()
	for {

		startTime := time.Now()

		/* If worksize is larger than remaining work, adjust worksize */
		if lastTock+wSize > world.TockCount {
			wSize = world.TockCount - lastTock
		}
		wg.Add()
		go func(wSize, lastTock int) {
			for i := lastTock; i < lastTock+wSize; i++ {
				/* Don't tock if blocked */
				if !world.TockList[i].Target.Blocked {
					world.TockList[i].Target.Unique.TypeP.UpdateObj(world.TockList[i].Target)
				}
			}
			wg.Done()
		}(wSize, lastTock)

		lastTock = lastTock + wSize + 1
		if lastTock >= world.TockCount {
			break
		}

		if !gv.UPSBench {
			sleepFor = time.Duration(world.ObjectUPS_ns/int(float64(world.TockCount)/(float64(wSize)/margin))) - time.Since(startTime)
			if sleepFor > minSleep {
				time.Sleep(sleepFor)
			}
		}
	}
	wg.Wait()
	world.TockListLock.Unlock()

	//fmt.Printf("TOCK: sleep-per: %v, workSize: %v\n", sleepFor.String(), wSize)
}

/* WASM single-thread: Run all object tocks (interal) */
func runTocksST() {
	for i := range world.TockList {
		/* Don't tock if blocked */
		if !world.TockList[i].Target.Blocked {
			world.TockList[i].Target.Unique.TypeP.UpdateObj(world.TockList[i].Target)
		}
	}
}

/* WASM single thread: Put our OutputBuffer to another object's InputBuffer (external)*/
func runTicksST() {
	for i := range world.TickList {
		tickObj(world.TickList[i].Target)
	}
}
