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

	lastClick    time.Time
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
	ToolBarScale := float32(gv.ToolBarScale * gv.UIScale)
	ToolBarIconSize := float32(gv.ToolBarIconSize)
	ToolBarSpacing := float32(gv.ToolBarSpacing)

	toolbarCacheLock.Lock()
	defer toolbarCacheLock.Unlock()

	if toolbarCache == nil {
		toolbarCache = ebiten.NewImage(int(ToolBarScale+ToolBarSpacing)*ToolbarMax, int(ToolBarScale+ToolBarSpacing))
	}
	toolbarCache.Clear()

	for pos := 0; pos < ToolbarMax; pos++ {
		item := ToolbarItems[pos]
		x := float64(int(ToolBarScale+ToolBarSpacing) * int(pos))

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

		vector.DrawFilledRect(toolbarCache, ToolBarSpacing+float32(pos)*(ToolBarScale+ToolBarSpacing),
			float32(gv.TbOffY), float32(ToolBarScale), float32(ToolBarScale), world.ColorToolTipBG, false)

		op.GeoM.Translate(x+(float64(ToolBarScale)-float64(ToolBarIconSize))-1, float64(ToolBarSpacing*2)+1)

		if item.SType == gv.ObjSubGame {

			if item.OType.TypeI == SelectedItemType {
				vector.DrawFilledRect(toolbarCache, ToolBarSpacing+float32(pos)*(ToolBarScale+ToolBarSpacing), gv.TbOffY, ToolBarScale, ToolBarScale, world.ColorDarkGray, false)
			}
		}

		if pos == index {
			if click {
				lastClick = time.Now()

				vector.DrawFilledRect(toolbarCache, ToolBarSpacing+float32(pos)*(ToolBarScale+ToolBarSpacing), gv.TbOffY, ToolBarScale, ToolBarScale, world.ColorRed, false)
				ToolbarHover = true

				go func() {
					time.Sleep(time.Millisecond * 155)
					DrawToolbar(false, false, 0)
				}()
			} else if hover {
				vector.DrawFilledRect(toolbarCache, ToolBarSpacing+float32(pos)*(ToolBarScale+ToolBarSpacing), gv.TbOffY, ToolBarScale, ToolBarScale, world.ColorAqua, false)
				ToolbarHover = true
			}

		}

		toolbarCache.DrawImage(img, op)

		if item.SType == gv.ObjSubGame {

			if item.OType.TypeI == SelectedItemType {
				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(ToolBarScale+ToolBarSpacing)+1,
					gv.TbOffY,

					(gv.TBSelThick),
					ToolBarScale,
					world.ColorTBSelected, false)

				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(ToolBarScale+ToolBarSpacing)+1,
					gv.TbOffY,

					ToolBarScale,
					(gv.TBSelThick),
					world.ColorTBSelected, false)

				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(ToolBarScale+ToolBarSpacing)+1,
					gv.TbOffY+ToolBarScale-(gv.TBSelThick),

					ToolBarScale,
					(gv.TBSelThick),
					world.ColorTBSelected, false)

				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(ToolBarScale+ToolBarSpacing)+gv.TbOffY+ToolBarScale-(gv.TBSelThick)+1,
					2,

					(gv.TBSelThick),
					ToolBarScale,
					world.ColorTBSelected, false)

			}
		}

		if item.OType.ToolBarArrow {
			var aop *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

			arrow := objects.WorldOverlays[item.OType.Direction].Images.Main
			if arrow != nil {
				if arrow.Bounds().Max.X != int(ToolBarScale) {
					aop.GeoM.Scale(1.0/(float64(arrow.Bounds().Max.X)/float64(ToolBarScale)),
						1.0/(float64(arrow.Bounds().Max.Y)/float64(ToolBarScale)))
				}
				aop.GeoM.Translate(x, 0)
				aop.ColorScale.Scale(0.5, 0.5, 0.5, 0.66)
				toolbarCache.DrawImage(arrow, aop)
			}
		}
	}
}
