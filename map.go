package main

import (
	"GameTest/glob"
	"GameTest/gv"
	"GameTest/objects"
	"GameTest/terrain"
	"time"
)

/* Make a test map, or skip and still start daemons */
func makeTestMap(skip bool) {
	time.Sleep(time.Second)

	if !skip {
		//start := time.Now()

		/* Test load map generator parameters */
		total := 0
		rows := 0
		columns := 0
		hSpace := 4
		vSpace := 4
		bLen := 2
		beltLength := hSpace + bLen
		for i := 0; total < gv.TestObjects; i++ {
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
				objects.CreateObj(glob.XY{X: tx + (cols * beltLength), Y: ty}, gv.ObjTypeBasicMiner, gv.DIR_EAST)
				Loaded++

				for i := 0; i < beltLength-hSpace; i++ {
					tx++
					objects.CreateObj(glob.XY{X: tx + (cols * beltLength), Y: ty}, gv.ObjTypeBasicBelt, gv.DIR_EAST)
					Loaded++
				}
				tx++
				objects.CreateObj(glob.XY{X: tx + (cols * beltLength), Y: ty}, gv.ObjTypeBasicBox, gv.DIR_EAST)
				Loaded++

				if cols%columns == 0 {
					ty += vSpace
					cols = 0
				}

				glob.MapLoadPercent = (float64(Loaded) / float64(total) * 100.0)
				if glob.WASMMode {
					time.Sleep(time.Nanosecond)
				}
			}
		} else {
			/* Default map generator */
			tx := int(gv.XYCenter - 5)
			ty := int(gv.XYCenter)
			total = 16

			objects.CreateObj(glob.XY{X: tx, Y: ty}, gv.ObjTypeBasicMiner, gv.DIR_EAST)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				tx++
				objects.CreateObj(glob.XY{X: tx, Y: ty}, gv.ObjTypeBasicBelt, gv.DIR_EAST)
				Loaded++
			}
			tx++
			objects.CreateObj(glob.XY{X: tx, Y: ty}, gv.ObjTypeBasicBox, gv.DIR_EAST)
			Loaded++

			tx = int(gv.XYCenter - 5)
			ty = int(gv.XYCenter - 2)
			objects.CreateObj(glob.XY{X: tx, Y: ty}, gv.ObjTypeBasicMiner, gv.DIR_WEST)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				tx--
				objects.CreateObj(glob.XY{X: tx, Y: ty}, gv.ObjTypeBasicBelt, gv.DIR_WEST)
				Loaded++
			}
			tx--
			objects.CreateObj(glob.XY{X: tx, Y: ty}, gv.ObjTypeBasicBox, gv.DIR_WEST)
			Loaded++

			tx = int(gv.XYCenter - 5)
			ty = int(gv.XYCenter + 2)
			objects.CreateObj(glob.XY{X: tx, Y: ty}, gv.ObjTypeBasicMiner, gv.DIR_SOUTH)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				ty++
				objects.CreateObj(glob.XY{X: tx, Y: ty}, gv.ObjTypeBasicBelt, gv.DIR_SOUTH)
				Loaded++
			}
			ty++
			objects.CreateObj(glob.XY{X: tx, Y: ty}, gv.ObjTypeBasicBox, gv.DIR_SOUTH)
			Loaded++

			tx = int(gv.XYCenter - 5)
			ty = int(gv.XYCenter - 4)
			objects.CreateObj(glob.XY{X: tx, Y: ty}, gv.ObjTypeBasicMiner, gv.DIR_NORTH)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				ty--
				objects.CreateObj(glob.XY{X: tx, Y: ty}, gv.ObjTypeBasicBelt, gv.DIR_NORTH)
				Loaded++
			}
			ty--
			objects.CreateObj(glob.XY{X: tx, Y: ty}, gv.ObjTypeBasicBox, gv.DIR_NORTH)
			Loaded++

			glob.MapLoadPercent = (float64(Loaded) / float64(total) * 100.0)
			if glob.WASMMode {
				time.Sleep(time.Nanosecond)
			}
		}
		if glob.WASMMode {
			time.Sleep(time.Nanosecond)
		}
		//objects.UnsafeMakeObjLists()
	}

	glob.MapGenerated.Store(true)

	for !glob.SpritesLoaded.Load() ||
		!glob.PlayerReady.Load() {
		time.Sleep(time.Millisecond * 100)
	}

	if !glob.WASMMode {
		go terrain.RenderTerrainDaemon()
		go terrain.PixmapRenderDaemon()
		go objects.ObjUpdateDaemon()
	} else {
		go objects.ObjUpdateDaemonST()
	}
}
