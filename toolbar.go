package main

import (
	"GameTest/glob"
	"GameTest/gv"
	"GameTest/objects"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	toolbarCache     *ebiten.Image
	ToolbarMax       int
	SelectedItemType uint8 = 0
	ToolbarItems           = []glob.ToolbarItem{}
)

/* Make default toolbar list */
func init() {

	ToolbarMax = 0
	for spos, stype := range objects.SubTypes {
		if spos == gv.ObjSubUI || spos == gv.ObjSubGame {
			for _, otype := range stype {
				ToolbarMax++
				ToolbarItems = append(ToolbarItems, glob.ToolbarItem{SType: spos, OType: otype})

			}
		}
	}
}

/* Draw toolbar to an image */
func DrawToolbar() {
	if toolbarCache == nil {
		toolbarCache = ebiten.NewImage(gv.ToolBarScale*ToolbarMax, gv.ToolBarScale)
	}
	toolbarCache.Clear()

	for pos := 0; pos < ToolbarMax; pos++ {
		item := ToolbarItems[pos]

		x := float64(gv.ToolBarScale * int(pos))

		if item.OType.Image == nil {
			return
		} else {
			var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

			op.GeoM.Reset()
			op.GeoM.Translate(x, 0)
			toolbarCache.DrawImage(glob.ToolBG, op)

			op.GeoM.Reset()
			iSize := item.OType.Image.Bounds()

			if item.OType.Rotatable && item.OType.Direction > 0 {
				x := float64(iSize.Size().X / 2)
				y := float64(iSize.Size().Y / 2)
				op.GeoM.Translate(-x, -y)
				op.GeoM.Rotate(gv.CNinetyDeg * float64(item.OType.Direction))
				op.GeoM.Translate(x, y)
			}

			if item.OType.Image.Bounds().Max.X != gv.ToolBarScale {
				op.GeoM.Scale(1.0/(float64(iSize.Max.X)/gv.ToolBarScale), 1.0/(float64(iSize.Max.Y)/gv.ToolBarScale))
			}
			op.GeoM.Translate(x, 0)

			toolbarCache.DrawImage(item.OType.Image, op)
		}

		if item.SType == gv.ObjSubGame {
			if item.OType.TypeI == SelectedItemType {
				ebitenutil.DrawRect(toolbarCache,
					gv.ToolBarOffsetX+float64(pos)*gv.ToolBarScale,
					gv.ToolBarOffsetY, gv.TBThick, gv.ToolBarScale, glob.ColorTBSelected)
				ebitenutil.DrawRect(toolbarCache,
					gv.ToolBarOffsetX+float64(pos)*gv.ToolBarScale,
					gv.ToolBarOffsetY, gv.ToolBarScale, gv.TBThick, glob.ColorTBSelected)

				ebitenutil.DrawRect(toolbarCache,
					gv.ToolBarOffsetX+float64(pos)*gv.ToolBarScale,
					gv.ToolBarOffsetY+gv.ToolBarScale-gv.TBThick, gv.ToolBarScale, gv.TBThick, glob.ColorTBSelected)
				ebitenutil.DrawRect(toolbarCache,
					gv.ToolBarOffsetX+(float64(pos)*gv.ToolBarScale)+gv.ToolBarScale-gv.TBThick,
					gv.ToolBarOffsetY, gv.TBThick, gv.ToolBarScale, glob.ColorTBSelected)
			}
		}
	}
}
