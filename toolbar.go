package main

import (
	"GameTest/gv"
	"GameTest/objects"
	"GameTest/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	toolbarCache     *ebiten.Image
	ToolbarMax       int
	SelectedItemType uint8 = 255
	ToolbarItems           = []world.ToolbarItem{}
)

/* Make default toolbar list */
func InitToolbar() {

	ToolbarMax = 0
	for spos, stype := range objects.SubTypes {
		if spos == gv.ObjSubUI || spos == gv.ObjSubGame {
			for _, otype := range stype {
				/* Skips some items for WASM */
				if gv.WASMMode && otype.ExcludeWASM {
					continue
				}
				ToolbarMax++
				ToolbarItems = append(ToolbarItems, world.ToolbarItem{SType: spos, OType: otype})

			}
		}
	}
}

/* Draw toolbar to an image */
func DrawToolbar() {
	if toolbarCache == nil {

		toolbarCache = ebiten.NewImage((gv.ToolBarScale+gv.ToolBarSpacing)*ToolbarMax, gv.ToolBarScale+gv.ToolBarSpacing)
	}
	toolbarCache.Clear()

	for pos := 0; pos < ToolbarMax; pos++ {
		item := ToolbarItems[pos]
		x := float64((gv.ToolBarScale + gv.ToolBarSpacing) * int(pos))

		img := item.OType.Image
		if item.OType.TBarImage != nil {
			img = item.OType.TBarImage
		}
		if img == nil {
			return
		}

		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

		op.GeoM.Reset()
		iSize := img.Bounds()

		if img.Bounds().Max.X != gv.ToolBarScale {
			op.GeoM.Scale(1.0/(float64(iSize.Max.X)/gv.ToolBarIcons), 1.0/(float64(iSize.Max.Y)/gv.ToolBarIcons))
		}

		if item.OType.Rotatable && item.OType.Direction > 0 {
			x := float64(gv.ToolBarScale / 2)
			y := float64(gv.ToolBarScale / 2)
			op.GeoM.Translate(-x, -y)
			op.GeoM.Rotate(gv.NinetyDeg * float64(item.OType.Direction))
			op.GeoM.Translate(x, y)
		}

		vector.DrawFilledRect(toolbarCache, gv.ToolBarSpacing+float32(pos)*(gv.ToolBarScale+gv.ToolBarSpacing), gv.TbOffY, gv.ToolBarScale, gv.ToolBarScale, world.ColorToolTipBG)
		op.GeoM.Translate(x+(gv.ToolBarScale-gv.ToolBarIcons)-1, (gv.ToolBarSpacing*2)+1)

		if item.SType == gv.ObjSubGame {

			if item.OType.TypeI == SelectedItemType+1 {
				vector.DrawFilledRect(toolbarCache, gv.ToolBarSpacing+float32(pos)*(gv.ToolBarScale+gv.ToolBarSpacing), gv.TbOffY, gv.ToolBarScale, gv.ToolBarScale, world.ColorLightAqua)
			}
		}
		toolbarCache.DrawImage(img, op)

		if item.SType == gv.ObjSubGame {

			if item.OType.TypeI == SelectedItemType+1 {
				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(gv.ToolBarScale+gv.ToolBarSpacing)+1,
					gv.TbOffY,

					(gv.TBSelThick),
					gv.ToolBarScale,
					world.ColorTBSelected)

				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(gv.ToolBarScale+gv.ToolBarSpacing)+1,
					gv.TbOffY,

					gv.ToolBarScale,
					(gv.TBSelThick),
					world.ColorTBSelected)

				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(gv.ToolBarScale+gv.ToolBarSpacing)+1,
					gv.TbOffY+gv.ToolBarScale-(gv.TBSelThick),

					gv.ToolBarScale,
					(gv.TBSelThick),
					world.ColorTBSelected)

				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(gv.ToolBarScale+gv.ToolBarSpacing)+gv.TbOffY+gv.ToolBarScale-(gv.TBSelThick)+1,
					2,

					(gv.TBSelThick),
					gv.ToolBarScale,
					world.ColorTBSelected)

			}
		}

		if item.OType.ToolBarArrow {
			var aop *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

			arrow := objects.ObjOverlayTypes[item.OType.Direction].Image
			if arrow != nil {
				if arrow.Bounds().Max.X != gv.ToolBarScale {
					aop.GeoM.Scale(1.0/(float64(arrow.Bounds().Max.X)/gv.ToolBarScale), 1.0/(float64(arrow.Bounds().Max.Y)/gv.ToolBarScale))
				}
				aop.GeoM.Translate(x, 0)
				toolbarCache.DrawImage(arrow, aop)
			}
		}
	}
}
