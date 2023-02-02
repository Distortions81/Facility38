package main

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/objects"
	"GameTest/terrain"
)

func makeTestMap(skip bool) {

	if !skip {
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

		if consts.LoadTest {

			ty := int(consts.XYCenter) - (rows)
			cols := 0
			for j := 0; j < rows*columns; j++ {
				cols++

				tx := int(consts.XYCenter) - (columns*(beltLength+hSpace))/2
				objects.FastCreateObj(glob.XY{X: tx + (cols * beltLength), Y: ty}, consts.ObjTypeBasicMiner, consts.DIR_EAST)

				for i := 0; i < beltLength-hSpace; i++ {
					tx++
					objects.FastCreateObj(glob.XY{X: tx + (cols * beltLength), Y: ty}, consts.ObjTypeBasicBelt, consts.DIR_EAST)

				}
				tx++
				objects.FastCreateObj(glob.XY{X: tx + (cols * beltLength), Y: ty}, consts.ObjTypeBasicBox, consts.DIR_EAST)

				if cols%columns == 0 {
					ty += vSpace
					cols = 0
				}
			}
		} else {
			/* Default map generator */
			tx := int(consts.XYCenter - 5)
			ty := int(consts.XYCenter)
			objects.FastCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicMiner, consts.DIR_EAST)
			for i := 0; i < beltLength-hSpace; i++ {
				tx++
				objects.FastCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBelt, consts.DIR_EAST)
			}
			tx++
			objects.FastCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBox, consts.DIR_EAST)

			tx = int(consts.XYCenter - 5)
			ty = int(consts.XYCenter - 2)
			objects.FastCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicMiner, consts.DIR_WEST)
			for i := 0; i < beltLength-hSpace; i++ {
				tx--
				objects.FastCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBelt, consts.DIR_WEST)
			}
			tx--
			objects.FastCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBox, consts.DIR_WEST)

			tx = int(consts.XYCenter - 5)
			ty = int(consts.XYCenter + 2)
			objects.FastCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicMiner, consts.DIR_SOUTH)
			for i := 0; i < beltLength-hSpace; i++ {
				ty++
				objects.FastCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBelt, consts.DIR_SOUTH)
			}
			ty++
			objects.FastCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBox, consts.DIR_SOUTH)

			tx = int(consts.XYCenter - 5)
			ty = int(consts.XYCenter - 4)
			objects.FastCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicMiner, consts.DIR_NORTH)
			for i := 0; i < beltLength-hSpace; i++ {
				ty--
				objects.FastCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBelt, consts.DIR_NORTH)
			}
			ty--
			objects.FastCreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBox, consts.DIR_NORTH)

		}
	}
	if !glob.WASMMode {
		go terrain.RenderTerrainDaemon()
		go terrain.PixmapRenderDaemon()
	}
	glob.MapGenerated.Store(true)
}
