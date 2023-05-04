package main

import (
	"time"
)

/* Make a test map, or skip and still start daemons */
func makeMap() {
	defer reportPanic("makeMap")
	gameLock.Lock()
	defer gameLock.Unlock()

	nukeWorld()
	if loadTest {

		/* Test load map generator parameters */
		total := 0
		rows := 0
		columns := 0
		hSpace := 10
		vSpace := 4
		bLen := 3
		beltLength := hSpace + bLen
		for i := 0; total < numTestObjects; i++ {
			if i%2 == 0 {
				rows++
			} else {
				columns++
			}

			total = (rows * columns) * (bLen + 4)
		}
		loaded := 0

		ty := int(xyCenter) - (rows)
		cols := 0
		for j := 0; j < rows*columns; j++ {
			cols++

			tx := int(xyCenter) - ((columns * (beltLength + hSpace)) / 3)
			placeObj(XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, objTypeBasicMiner, nil, DIR_EAST, true)
			tx++
			tx++
			loaded++

			placeObj(XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, objTypeBasicUnloader, nil, DIR_EAST, true)
			tx++
			loaded++

			for i := 0; i < beltLength-hSpace; i++ {
				placeObj(XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, objTypeBasicBelt, nil, DIR_EAST, true)
				tx++
				loaded++
			}

			placeObj(XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, objTypeBasicLoader, nil, DIR_EAST, true)
			tx++
			loaded++

			placeObj(XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, objTypeBasicBox, nil, DIR_EAST, true)
			tx++
			tx++
			loaded++

			if cols%columns == 0 {
				ty += vSpace
				cols = 0
			}

			mapLoadPercent = (float32(loaded) / float32(total) * 100.0)
			if loaded%10000 == 0 {
				wasmSleep()
			}
			runEventQueue()
		}

	}

	wasmSleep()
	exploreMap(XY{X: xyCenter - (chunkSize / 2), Y: xyCenter - (chunkSize / 2)}, 16, true)

	lastSave = time.Now().UTC()

	mapLoadPercent = 100
	mapGenerated.Store(true)
}
