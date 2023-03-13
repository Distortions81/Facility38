package objects

import (
	"GameTest/world"
	"fmt"
	"math"
	"time"
)

const (
	STATE_TOCK  = 0
	STATE_TICK  = 1
	STATE_QUEUE = 2

	minWorkSize = 1000
	margin      = 1.8
	minSleep    = 200 //Sleeping for less than this does not appear effective.

	largeThreshold  = minWorkSize
	blocksPerWorker = 10
)

/* Process internally in an object, multi-threaded*/
func runTicks() {
	if world.TickCount == 0 {
		time.Sleep(time.Millisecond)
		return
	}
	var lastTick int
	var sleepFor time.Duration
	var maxBlocks = world.NumWorkers * blocksPerWorker
	tickStart := time.Now()
	wSize := minWorkSize

	if world.TickCount > largeThreshold {
		wSize = int(math.Ceil(float64(world.TickCount)/float64(maxBlocks))) + minWorkSize
	}

	world.TickListLock.Lock()
	for {

		startTime := time.Now()

		/* If worksize is larger than remaining work, adjust worksize */
		if wSize > world.TickCount-lastTick {
			wSize = world.TickCount - lastTick
		}

		wg.Add()
		go func(wSize, lastTick int) {
			for i := lastTick; i < lastTick+wSize; i++ {
				world.TickList[i].Target.TypeP.UpdateObj(world.TickList[i].Target)
			}
			wg.Done()
		}(wSize, lastTick)

		lastTick = lastTick + wSize
		if lastTick >= world.TickCount {
			break
		}
		if world.TockCount-lastTick <= 0 {
			break
		}

		sleepFor = time.Duration(world.ObjectUPS_ns/int(float64(world.TickCount)/(float64(wSize)/margin))) - time.Since(startTime)
		if sleepFor > minSleep*time.Microsecond {
			time.Sleep(sleepFor)
		}
	}
	wg.Wait()
	world.TickListLock.Unlock()

	timeLeft := time.Duration(world.ObjectUPS_ns) - time.Since(tickStart)
	if timeLeft > minSleep {
		time.Sleep(timeLeft)
	}

	world.MeasuredObjectUPS_ns = int(time.Since(tickStart).Nanoseconds())
	fmt.Printf("TICK: sleep-per: %v, endSleep: %v, workSize: %v\n", sleepFor.String(), timeLeft.String(), wSize)

}

func runTocks() {
	if world.TockCount == 0 {
		time.Sleep(time.Millisecond)
		return
	}
	var lastTock int = 0
	var sleepFor time.Duration
	var maxBlocks = world.NumWorkers * blocksPerWorker
	tockStart := time.Now()
	wSize := minWorkSize

	if world.TockCount > largeThreshold {
		wSize = int(math.Ceil(float64(world.TockCount)/float64(maxBlocks))) + minWorkSize
	}

	world.TockListLock.Lock()
	for {

		startTime := time.Now()

		/* If worksize is larger than remaining work, adjust worksize */
		if wSize > world.TockCount-lastTock {
			wSize = world.TockCount - lastTock
		}

		wg.Add()
		go func(wSize, lastTock int) {
			for i := lastTock; i < lastTock+wSize; i++ {
				world.TockList[i].Target.TypeP.UpdateObj(world.TockList[i].Target)
			}
			wg.Done()
		}(wSize, lastTock)

		lastTock = lastTock + wSize
		if lastTock >= world.TockCount {
			break
		}
		if world.TockCount-lastTock <= 0 {
			break
		}

		sleepFor = time.Duration(world.ObjectUPS_ns/int(float64(world.TockCount)/(float64(wSize)/margin))) - time.Since(startTime)
		if sleepFor > minSleep*time.Microsecond {
			time.Sleep(sleepFor)
		}
	}
	wg.Wait()
	world.TockListLock.Unlock()

	timeLeft := time.Duration(world.ObjectUPS_ns) - time.Since(tockStart)
	if timeLeft > minSleep {
		time.Sleep(timeLeft)
	}

	world.MeasuredObjectUPS_ns = int(time.Since(tockStart).Nanoseconds())
	fmt.Printf("TOCK: sleep-per: %v, endSleep: %v, workSize: %v\n", sleepFor.String(), timeLeft.String(), wSize)

}
