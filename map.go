package main

import (
	"GameTest/consts"
	"GameTest/glob"
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
		for i := 0; total < consts.TestObjects; i++ {
			if i%2 == 0 {
				rows++
			} else {
				columns++
			}

			total = (rows * columns) * (bLen + 2)
		}
		Loaded := 0

		if consts.LoadTest {

			ty := int(consts.XYCenter) - (rows)
			cols := 0
			for j := 0; j < rows*columns; j++ {
				cols++

				tx := int(consts.XYCenter) - (columns*(beltLength+hSpace))/2
				objects.UnsafeCreateObj(glob.XY{X: tx + (cols * beltLength), Y: ty}, consts.ObjTypeBasicMiner, consts.DIR_EAST)
				Loaded++

				for i := 0; i < beltLength-hSpace; i++ {
					tx++
					objects.UnsafeCreateObj(glob.XY{X: tx + (cols * beltLength), Y: ty}, consts.ObjTypeBasicBelt, consts.DIR_EAST)
					Loaded++
				}
				tx++
				objects.UnsafeCreateObj(glob.XY{X: tx + (cols * beltLength), Y: ty}, consts.ObjTypeBasicBox, consts.DIR_EAST)
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
			tx := int(consts.XYCenter - 5)
			ty := int(consts.XYCenter)
			objects.UnsafeCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicMiner, consts.DIR_EAST)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				tx++
				objects.UnsafeCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBelt, consts.DIR_EAST)
				Loaded++
			}
			tx++
			objects.UnsafeCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBox, consts.DIR_EAST)
			Loaded++

			tx = int(consts.XYCenter - 5)
			ty = int(consts.XYCenter - 2)
			objects.UnsafeCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicMiner, consts.DIR_WEST)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				tx--
				objects.UnsafeCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBelt, consts.DIR_WEST)
				Loaded++
			}
			tx--
			objects.UnsafeCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBox, consts.DIR_WEST)
			Loaded++

			tx = int(consts.XYCenter - 5)
			ty = int(consts.XYCenter + 2)
			objects.UnsafeCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicMiner, consts.DIR_SOUTH)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				ty++
				objects.UnsafeCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBelt, consts.DIR_SOUTH)
				Loaded++
			}
			ty++
			objects.UnsafeCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBox, consts.DIR_SOUTH)
			Loaded++

			tx = int(consts.XYCenter - 5)
			ty = int(consts.XYCenter - 4)
			objects.UnsafeCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicMiner, consts.DIR_NORTH)
			Loaded++
			for i := 0; i < beltLength-hSpace; i++ {
				ty--
				objects.UnsafeCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBelt, consts.DIR_NORTH)
				Loaded++
			}
			ty--
			objects.UnsafeCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBox, consts.DIR_NORTH)
			Loaded++

			glob.MapLoadPercent = (float64(Loaded) / float64(total) * 100.0)
			if glob.WASMMode {
				time.Sleep(time.Nanosecond)
			}
		}
		if glob.WASMMode {
			time.Sleep(time.Millisecond * 100)
		}
		objects.UnsafeMakeObjLists()
		objects.UnsafeMakeEventLists()
	}
	if !glob.WASMMode {
		go terrain.RenderTerrainDaemon()
		go terrain.PixmapRenderDaemon()
		go objects.ObjUpdateDaemon()
	} else {
		go objects.ObjUpdateDaemonST()
	}
	glob.MapGenerated.Store(true)
	if glob.WASMMode {
		time.Sleep(time.Millisecond * 100)
	}
}
