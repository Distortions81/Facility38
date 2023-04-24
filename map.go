package main

import (
	"Facility38/gv"
	"Facility38/util"
	"Facility38/world"
)

/* Make a test map, or skip and still start daemons */
func MakeMap(gen bool) {
	GameLock.Lock()
	defer GameLock.Unlock()

	NukeWorld()
	if gen {

		/* Test load map generator parameters */
		total := 0
		rows := 0
		columns := 0
		hSpace := 10
		vSpace := 4
		bLen := 3
		beltLength := hSpace + bLen
		for i := 0; total < gv.NumTestObjects; i++ {
			if i%2 == 0 {
				rows++
			} else {
				columns++
			}

			total = (rows * columns) * (bLen + 4)
		}
		Loaded := 0

		if gv.LoadTest {

			ty := int(gv.XYCenter) - (rows)
			cols := 0
			for j := 0; j < rows*columns; j++ {
				cols++

				tx := int(gv.XYCenter) - ((columns * (beltLength + hSpace)) / 3)
				PlaceObj(world.XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, gv.ObjTypeBasicMiner, nil, gv.DIR_EAST, true)
				tx++
				tx++
				Loaded++

				PlaceObj(world.XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, gv.ObjTypeBasicUnloader, nil, gv.DIR_EAST, true)
				tx++
				Loaded++

				for i := 0; i < beltLength-hSpace; i++ {
					PlaceObj(world.XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, gv.ObjTypeBasicBelt, nil, gv.DIR_EAST, true)
					tx++
					Loaded++
				}

				PlaceObj(world.XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, gv.ObjTypeBasicLoader, nil, gv.DIR_EAST, true)
				tx++
				Loaded++

				PlaceObj(world.XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, gv.ObjTypeBasicBox, nil, gv.DIR_EAST, true)
				tx++
				tx++
				Loaded++

				if cols%columns == 0 {
					ty += vSpace
					cols = 0
				}

				world.MapLoadPercent = (float32(Loaded) / float32(total) * 100.0)
				if Loaded%10000 == 0 {
					util.WASMSleep()
				}
				RunEventQueue()
			}
		}
	}

	util.WASMSleep()
	ExploreMap(world.XY{X: gv.XYCenter - (gv.ChunkSize / 2), Y: gv.XYCenter - (gv.ChunkSize / 2)}, 16, gv.WASMMode)

	world.MapLoadPercent = 100
	world.MapGenerated.Store(true)
}
