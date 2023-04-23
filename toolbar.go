package main

import (
	"Facility38/gv"
	"Facility38/objects"
	"Facility38/world"
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

	ToolbarHover bool
)

/* Make default toolbar list */
func InitToolbar() {

	ToolbarMax = 0
	for spos, stype := range objects.SubTypes {
		if spos == gv.ObjSubUI || spos == gv.ObjSubGame {
			for _, otype := range stype.List {
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
	ToolBarIconSize := float32(gv.ToolBarIconSize)
	ToolBarSpacing := float32(gv.ToolBarIconSize / 8)

	toolbarCacheLock.Lock()
	defer toolbarCacheLock.Unlock()

	if toolbarCache == nil {
		toolbarCache = ebiten.NewImage(int(ToolBarIconSize+ToolBarSpacing)*ToolbarMax, int(ToolBarIconSize+ToolBarSpacing))
	}
	toolbarCache.Clear()
	toolbarCache.Fill(world.ColorToolTipBG)

	for pos := 0; pos < ToolbarMax; pos++ {
		item := ToolbarItems[pos]
		x := float64(int(ToolBarIconSize+ToolBarSpacing) * int(pos))

		img := item.OType.Images.Main
		if item.OType.Images.Overlay != nil {
			img = item.OType.Images.Overlay
		}
		if item.OType.Images.Toolbar != nil {
			img = item.OType.Images.Toolbar
		}
		if img == nil {
			return
		}

		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

		op.GeoM.Reset()
		iSize := img.Bounds()
		op.GeoM.Scale(gv.UIScale/(float64(iSize.Max.X)/float64(ToolBarIconSize)), gv.UIScale/(float64(iSize.Max.Y)/float64(ToolBarIconSize)))

		if item.OType.Rotatable && item.OType.Direction > 0 {
			x := float64(ToolBarIconSize / 2)
			y := float64(ToolBarIconSize / 2)
			op.GeoM.Translate(-x, -y)
			op.GeoM.Rotate(gv.NinetyDeg * float64(item.OType.Direction))
			op.GeoM.Translate(x, y)
		}

		op.GeoM.Translate(x+(float64(ToolBarIconSize)-float64(ToolBarIconSize))-1, float64(ToolBarSpacing*2)+1)

		if item.SType == gv.ObjSubGame {

			if item.OType.TypeI == SelectedItemType {
				vector.DrawFilledRect(toolbarCache, ToolBarSpacing+float32(pos)*(ToolBarIconSize+ToolBarSpacing),
					gv.TbOffY, ToolBarIconSize, ToolBarIconSize, world.ColorDarkGray, false)
			}
		}

		if pos == index {
			if click {
				vector.DrawFilledRect(toolbarCache, ToolBarSpacing+float32(pos)*(ToolBarIconSize+ToolBarSpacing),
					gv.TbOffY, ToolBarIconSize, ToolBarIconSize, world.ColorRed, false)
				ToolbarHover = true

				go func() {
					time.Sleep(time.Millisecond * 155)
					DrawToolbar(false, false, 0)
				}()
			} else if hover {
				vector.DrawFilledRect(toolbarCache, ToolBarSpacing+float32(pos)*(ToolBarIconSize+ToolBarSpacing),
					gv.TbOffY, ToolBarIconSize, ToolBarIconSize, world.ColorAqua, false)
				ToolbarHover = true
			}

		}

		toolbarCache.DrawImage(img, op)

		if item.SType == gv.ObjSubGame {

			if item.OType.TypeI == SelectedItemType {
				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(ToolBarIconSize+ToolBarSpacing)+1,
					gv.TbOffY,

					(gv.TBSelThick),
					ToolBarIconSize,
					world.ColorTBSelected, false)

				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(ToolBarIconSize+ToolBarSpacing)+1,
					gv.TbOffY,

					ToolBarIconSize,
					(gv.TBSelThick),
					world.ColorTBSelected, false)

				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(ToolBarIconSize+ToolBarSpacing)+1,
					gv.TbOffY+ToolBarIconSize-(gv.TBSelThick),

					ToolBarIconSize,
					(gv.TBSelThick),
					world.ColorTBSelected, false)

				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(ToolBarIconSize+ToolBarSpacing)+gv.TbOffY+ToolBarIconSize-(gv.TBSelThick)+1,
					2,

					(gv.TBSelThick),
					ToolBarIconSize,
					world.ColorTBSelected, false)

			}
		}

		if item.OType.ToolBarArrow {
			var aop *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

			arrow := objects.WorldOverlays[item.OType.Direction].Images.Main
			if arrow != nil {
				if arrow.Bounds().Max.X != int(ToolBarIconSize) {
					aop.GeoM.Scale(1.0/(float64(arrow.Bounds().Max.X)/float64(ToolBarIconSize)),
						1.0/(float64(arrow.Bounds().Max.Y)/float64(ToolBarIconSize)))
				}
				aop.GeoM.Translate(x, 0)
				aop.ColorScale.Scale(0.5, 0.5, 0.5, 0.66)
				toolbarCache.DrawImage(arrow, aop)
			}
		}
	}
}
