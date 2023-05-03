package main

import (
	"time"
)

/* Make a test map, or skip and still start daemons */
func makeMap(gen bool) {
	defer reportPanic("makeMap")
	GameLock.Lock()
	defer GameLock.Unlock()

	nukeWorld()
	if gen {

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
		Loaded := 0

		if LoadTest {

			ty := int(xyCenter) - (rows)
			cols := 0
			for j := 0; j < rows*columns; j++ {
				cols++

				tx := int(xyCenter) - ((columns * (beltLength + hSpace)) / 3)
				placeObj(XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, objTypeBasicMiner, nil, DIR_EAST, true)
				tx++
				tx++
				Loaded++

				placeObj(XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, objTypeBasicUnloader, nil, DIR_EAST, true)
				tx++
				Loaded++

				for i := 0; i < beltLength-hSpace; i++ {
					placeObj(XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, objTypeBasicBelt, nil, DIR_EAST, true)
					tx++
					Loaded++
				}

				placeObj(XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, objTypeBasicLoader, nil, DIR_EAST, true)
				tx++
				Loaded++

				placeObj(XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, objTypeBasicBox, nil, DIR_EAST, true)
				tx++
				tx++
				Loaded++

				if cols%columns == 0 {
					ty += vSpace
					cols = 0
				}

				MapLoadPercent = (float32(Loaded) / float32(total) * 100.0)
				if Loaded%10000 == 0 {
					WASMSleep()
				}
				RunEventQueue()
			}
		}
	}

	WASMSleep()
	exploreMap(XY{X: xyCenter - (chunkSize / 2), Y: xyCenter - (chunkSize / 2)}, 16, true)

	LastSave = time.Now().UTC()

	MapLoadPercent = 100
	MapGenerated.Store(true)
}
