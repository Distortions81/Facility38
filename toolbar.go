package main

import (
	"GameTest/gv"
	"GameTest/objects"
	"GameTest/world"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	toolbarCache     *ebiten.Image
	toolbarCacheLock sync.RWMutex
	ToolbarMax       int
	SelectedItemType uint8 = gv.MaxItemType
	ToolbarItems           = []world.ToolbarItem{}

	lastClick    time.Time
	ToolbarHover bool
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
func DrawToolbar(click, hover bool, index int) {
	if time.Since(lastClick) < time.Millisecond*150 {
		return
	}
	toolbarCacheLock.Lock()
	defer toolbarCacheLock.Unlock()

	if toolbarCache == nil {

		toolbarCache = ebiten.NewImage((gv.ToolBarScale+gv.ToolBarSpacing)*ToolbarMax, gv.ToolBarScale+gv.ToolBarSpacing)
	}
	toolbarCache.Clear()

	for pos := 0; pos < ToolbarMax; pos++ {
		item := ToolbarItems[pos]
		x := float64((gv.ToolBarScale + gv.ToolBarSpacing) * int(pos))

		img := item.OType.Image
		if item.OType.ImageOverlay != nil {
			img = item.OType.ImageOverlay
		}
		if item.OType.ToolbarImage != nil {
			img = item.OType.ToolbarImage
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
			x := float64(gv.ToolBarIcons / 2)
			y := float64(gv.ToolBarIcons / 2)
			op.GeoM.Translate(-x, -y)
			op.GeoM.Rotate(gv.NinetyDeg * float64(item.OType.Direction))
			op.GeoM.Translate(x, y)
		}

		vector.DrawFilledRect(toolbarCache, gv.ToolBarSpacing+float32(pos)*(gv.ToolBarScale+gv.ToolBarSpacing), gv.TbOffY, gv.ToolBarScale, gv.ToolBarScale, world.ColorToolTipBG)

		op.GeoM.Translate(x+(gv.ToolBarScale-gv.ToolBarIcons)-1, (gv.ToolBarSpacing*2)+1)

		if item.SType == gv.ObjSubGame {

			if item.OType.TypeI == SelectedItemType {
				vector.DrawFilledRect(toolbarCache, gv.ToolBarSpacing+float32(pos)*(gv.ToolBarScale+gv.ToolBarSpacing), gv.TbOffY, gv.ToolBarScale, gv.ToolBarScale, world.ColorDarkGray)
			}
		}

		if pos == index {
			if click {
				lastClick = time.Now()

				vector.DrawFilledRect(toolbarCache, gv.ToolBarSpacing+float32(pos)*(gv.ToolBarScale+gv.ToolBarSpacing), gv.TbOffY, gv.ToolBarScale, gv.ToolBarScale, world.ColorRed)
				ToolbarHover = true

				go func() {
					time.Sleep(time.Millisecond * 155)
					DrawToolbar(false, false, 0)
				}()
			} else if hover {
				vector.DrawFilledRect(toolbarCache, gv.ToolBarSpacing+float32(pos)*(gv.ToolBarScale+gv.ToolBarSpacing), gv.TbOffY, gv.ToolBarScale, gv.ToolBarScale, world.ColorAqua)
				ToolbarHover = true
			}

		}

		toolbarCache.DrawImage(img, op)

		if item.SType == gv.ObjSubGame {

			if item.OType.TypeI == SelectedItemType {
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
				aop.ColorScale.Scale(0.5, 0.5, 0.5, 0.66)
				toolbarCache.DrawImage(arrow, aop)
			}
		}
	}
}
