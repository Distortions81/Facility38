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
		for i := 0; total < NumTestObjects; i++ {
			if i%2 == 0 {
				rows++
			} else {
				columns++
			}

			total = (rows * columns) * (bLen + 4)
		}
		Loaded := 0

		if LoadTest {

			ty := int(XYCenter) - (rows)
			cols := 0
			for j := 0; j < rows*columns; j++ {
				cols++

				tx := int(XYCenter) - ((columns * (beltLength + hSpace)) / 3)
				placeObj(XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, ObjTypeBasicMiner, nil, DIR_EAST, true)
				tx++
				tx++
				Loaded++

				placeObj(XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, ObjTypeBasicUnloader, nil, DIR_EAST, true)
				tx++
				Loaded++

				for i := 0; i < beltLength-hSpace; i++ {
					placeObj(XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, ObjTypeBasicBelt, nil, DIR_EAST, true)
					tx++
					Loaded++
				}

				placeObj(XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, ObjTypeBasicLoader, nil, DIR_EAST, true)
				tx++
				Loaded++

				placeObj(XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, ObjTypeBasicBox, nil, DIR_EAST, true)
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
	exploreMap(XY{X: XYCenter - (ChunkSize / 2), Y: XYCenter - (ChunkSize / 2)}, 16, true)

	LastSave = time.Now().UTC()

	MapLoadPercent = 100
	MapGenerated.Store(true)
}
