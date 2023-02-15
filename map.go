package main

import (
	"GameTest/gv"
	"GameTest/objects"
	"GameTest/util"
	"GameTest/world"
	"time"
)

/* Make a test map, or skip and still start daemons */
func makeTestMap(skip bool) {

	objects.PerlinNoiseInit()

	if !skip {
		//start := time.Now()

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
				objects.CreateObj(world.XY{X: tx + (cols * beltLength), Y: ty}, gv.ObjTypeBasicMiner, gv.DIR_EAST)
				Loaded++

				for i := 0; i < beltLength-hSpace; i++ {
					tx++
					objects.CreateObj(world.XY{X: tx + (cols * beltLength), Y: ty}, gv.ObjTypeBasicBelt, gv.DIR_EAST)
					Loaded++
				}
				tx++
				objects.CreateObj(world.XY{X: tx + (cols * beltLength), Y: ty}, gv.ObjTypeBasicBox, gv.DIR_EAST)
				Loaded++

				if cols%columns == 0 {
					ty += vSpace
					cols = 0
				}

				world.MapLoadPercent = (float32(Loaded) / float32(total) * 100.0)
			}
		} else {
			/* Default map generator */
			tx := int(gv.XYCenter - 5)
			ty := int(gv.XYCenter)
			total = 16

			objects.CreateObj(world.XY{X: tx, Y: ty}, gv.ObjTypeBasicMiner, gv.DIR_EAST)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				tx++
				objects.CreateObj(world.XY{X: tx, Y: ty}, gv.ObjTypeBasicBelt, gv.DIR_EAST)
				Loaded++
			}
			tx++
			objects.CreateObj(world.XY{X: tx, Y: ty}, gv.ObjTypeBasicBox, gv.DIR_EAST)
			Loaded++

			tx = int(gv.XYCenter - 5)
			ty = int(gv.XYCenter - 2)
			objects.CreateObj(world.XY{X: tx, Y: ty}, gv.ObjTypeBasicMiner, gv.DIR_WEST)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				tx--
				objects.CreateObj(world.XY{X: tx, Y: ty}, gv.ObjTypeBasicBelt, gv.DIR_WEST)
				Loaded++
			}
			tx--
			objects.CreateObj(world.XY{X: tx, Y: ty}, gv.ObjTypeBasicBox, gv.DIR_WEST)
			Loaded++

			tx = int(gv.XYCenter - 5)
			ty = int(gv.XYCenter + 2)
			objects.CreateObj(world.XY{X: tx, Y: ty}, gv.ObjTypeBasicMiner, gv.DIR_SOUTH)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				ty++
				objects.CreateObj(world.XY{X: tx, Y: ty}, gv.ObjTypeBasicBelt, gv.DIR_SOUTH)
				Loaded++
			}
			ty++
			objects.CreateObj(world.XY{X: tx, Y: ty}, gv.ObjTypeBasicBox, gv.DIR_SOUTH)
			Loaded++

			tx = int(gv.XYCenter - 5)
			ty = int(gv.XYCenter - 4)
			objects.CreateObj(world.XY{X: tx, Y: ty}, gv.ObjTypeBasicMiner, gv.DIR_NORTH)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				ty--
				objects.CreateObj(world.XY{X: tx, Y: ty}, gv.ObjTypeBasicBelt, gv.DIR_NORTH)
				Loaded++
			}
			ty--
			objects.CreateObj(world.XY{X: tx, Y: ty}, gv.ObjTypeBasicBox, gv.DIR_NORTH)
			Loaded++

			world.MapLoadPercent = (float32(Loaded) / float32(total) * 100.0)
		}
	}

	objects.ExploreMap(world.XY{X: gv.XYCenter - (gv.ChunkSize / 2), Y: gv.XYCenter - (gv.ChunkSize / 2)}, 16)

	world.MapGenerated.Store(true)
	util.Chat("Map loaded, click or press any key to continue.")

	for !world.SpritesLoaded.Load() ||
		!world.PlayerReady.Load() {
		time.Sleep(time.Millisecond * 10)
	}
	util.Chat("Welcome! Click an item in the toolbar to select it, click ground to build.")

	if !gv.WASMMode {
		go objects.RenderTerrainDaemon()
		go objects.PixmapRenderDaemon()
		go objects.ObjUpdateDaemon()
	} else {
		go objects.ObjUpdateDaemonST()
	}
}
