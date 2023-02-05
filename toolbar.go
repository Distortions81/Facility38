package main

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/objects"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	toolbarCache     *ebiten.Image
	ToolbarMax       int
	SelectedItemType int = 0
	ToolbarItems         = []glob.ToolbarItem{}
)

/* Make default toolbar list */
func init() {

	ToolbarMax = 0
	for spos, stype := range objects.SubTypes {
		if spos == consts.ObjSubUI || spos == consts.ObjSubGame {
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
		toolbarCache = ebiten.NewImage(consts.ToolBarScale*ToolbarMax, consts.ToolBarScale)
	}

	for pos := 0; pos < ToolbarMax; pos++ {
		item := ToolbarItems[pos]

		x := float64(consts.ToolBarScale * int(pos))

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
				op.GeoM.Rotate(cNinetyDeg * float64(item.OType.Direction))
				op.GeoM.Translate(x, y)
			}

			if item.OType.Image.Bounds().Max.X != consts.ToolBarScale {
				op.GeoM.Scale(1.0/(float64(iSize.Max.X)/consts.ToolBarScale), 1.0/(float64(iSize.Max.Y)/consts.ToolBarScale))
			}
			op.GeoM.Translate(x, 0)

			toolbarCache.DrawImage(item.OType.Image, op)
		}

		if item.SType == consts.ObjSubGame {
			if item.OType.TypeI == SelectedItemType {
				ebitenutil.DrawRect(toolbarCache,
					consts.ToolBarOffsetX+float64(pos)*consts.ToolBarScale,
					consts.ToolBarOffsetY, consts.TBThick, consts.ToolBarScale, glob.ColorTBSelected)
				ebitenutil.DrawRect(toolbarCache,
					consts.ToolBarOffsetX+float64(pos)*consts.ToolBarScale,
					consts.ToolBarOffsetY, consts.ToolBarScale, consts.TBThick, glob.ColorTBSelected)

				ebitenutil.DrawRect(toolbarCache,
					consts.ToolBarOffsetX+float64(pos)*consts.ToolBarScale,
					consts.ToolBarOffsetY+consts.ToolBarScale-consts.TBThick, consts.ToolBarScale, consts.TBThick, glob.ColorTBSelected)
				ebitenutil.DrawRect(toolbarCache,
					consts.ToolBarOffsetX+(float64(pos)*consts.ToolBarScale)+consts.ToolBarScale-consts.TBThick,
					consts.ToolBarOffsetY, consts.TBThick, consts.ToolBarScale, glob.ColorTBSelected)
			}
		}
	}
}
