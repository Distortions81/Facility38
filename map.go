package main

import (
	"GameTest/gv"
	"GameTest/objects"
	"GameTest/util"
	"GameTest/world"
)

/* Make a test map, or skip and still start daemons */
func makeMap(gen bool) {

	if gen {
		objects.NukeWorld()

		/* Test load map generator parameters */
		total := 0
		rows := 0
		columns := 0
		hSpace := 4
		vSpace := 2
		bLen := 2
		beltLength := hSpace + bLen
		for i := 0; total < gv.NumTestObjects; i++ {
			if i%2 == 0 {
				rows++
			} else {
				columns++
			}

			total = (rows * columns) * (bLen + 2)
		}
		Loaded := 0

		if gv.LoadTest {

			ty := int(gv.XYCenter) - (rows)
			cols := 0
			for j := 0; j < rows*columns; j++ {
				cols++

				tx := int(gv.XYCenter) - (columns*(beltLength+hSpace))/2
				objects.CreateObj(world.XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, gv.ObjTypeBasicMiner, gv.DIR_EAST, true)
				tx++
				Loaded++

				for i := 0; i < beltLength-hSpace; i++ {
					tx++
					objects.CreateObj(world.XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, gv.ObjTypeBasicBelt, gv.DIR_EAST, true)
					Loaded++
				}
				tx++
				objects.CreateObj(world.XY{X: uint16(tx + (cols * beltLength)), Y: uint16(ty)}, gv.ObjTypeBasicBox, gv.DIR_EAST, true)
				Loaded++

				if cols%columns == 0 {
					ty += vSpace
					cols = 0
				}

				world.MapLoadPercent = (float32(Loaded) / float32(total) * 100.0)
				if Loaded%10000 == 0 {
					util.WASMSleep()
				}
				objects.RunEventQueue()
			}
		} else {
			/* Default map generator */
			tx := int(gv.XYCenter - 5)
			ty := int(gv.XYCenter)
			total = 16

			objects.CreateObj(world.XY{X: uint16(tx), Y: uint16(ty)}, gv.ObjTypeBasicMiner, gv.DIR_EAST, true)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				tx++
				objects.CreateObj(world.XY{X: uint16(tx), Y: uint16(ty)}, gv.ObjTypeBasicBelt, gv.DIR_EAST, true)
				Loaded++
			}
			tx++
			objects.CreateObj(world.XY{X: uint16(tx), Y: uint16(ty)}, gv.ObjTypeBasicBox, gv.DIR_EAST, true)
			Loaded++

			tx = int(gv.XYCenter - 5)
			ty = int(gv.XYCenter - 2)
			objects.CreateObj(world.XY{X: uint16(tx), Y: uint16(ty)}, gv.ObjTypeBasicMiner, gv.DIR_WEST, true)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				tx--
				objects.CreateObj(world.XY{X: uint16(tx), Y: uint16(ty)}, gv.ObjTypeBasicBelt, gv.DIR_WEST, true)
				Loaded++
			}
			tx--
			objects.CreateObj(world.XY{X: uint16(tx), Y: uint16(ty)}, gv.ObjTypeBasicBox, gv.DIR_WEST, true)
			Loaded++

			tx = int(gv.XYCenter - 5)
			ty = int(gv.XYCenter + 2)
			objects.CreateObj(world.XY{X: uint16(tx), Y: uint16(ty)}, gv.ObjTypeBasicMiner, gv.DIR_SOUTH, true)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				ty++
				objects.CreateObj(world.XY{X: uint16(tx), Y: uint16(ty)}, gv.ObjTypeBasicBelt, gv.DIR_SOUTH, true)
				Loaded++
			}
			ty++
			objects.CreateObj(world.XY{X: uint16(tx), Y: uint16(ty)}, gv.ObjTypeBasicBox, gv.DIR_SOUTH, true)
			Loaded++

			tx = int(gv.XYCenter - 5)
			ty = int(gv.XYCenter - 4)
			objects.CreateObj(world.XY{X: uint16(tx), Y: uint16(ty)}, gv.ObjTypeBasicMiner, gv.DIR_NORTH, true)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				ty--
				objects.CreateObj(world.XY{X: uint16(tx), Y: uint16(ty)}, gv.ObjTypeBasicBelt, gv.DIR_NORTH, true)
				Loaded++
			}
			ty--
			objects.CreateObj(world.XY{X: uint16(tx), Y: uint16(ty)}, gv.ObjTypeBasicBox, gv.DIR_NORTH, true)
			Loaded++

			world.MapLoadPercent = (float32(Loaded) / float32(total) * 100.0)
			if Loaded%10000 == 0 {
				util.WASMSleep()
			}
		}
	}

	util.WASMSleep()
	objects.ExploreMap(world.XY{X: gv.XYCenter - (gv.ChunkSize / 2), Y: gv.XYCenter - (gv.ChunkSize / 2)}, 16, true)

	world.MapLoadPercent = 100
	world.MapGenerated.Store(true)
}
