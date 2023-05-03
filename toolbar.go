package main

import (
	"Facility38/def"
	"Facility38/util"
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
	SelectedItemType uint8 = def.MaxItemType
	ToolbarItems           = []world.ToolbarItem{}

	ToolbarHover bool
)

/* Make default toolbar list */
func InitToolbar() {
	defer util.ReportPanic("InitToolbar")
	ToolbarMax = 0
	for spos, stype := range SubTypes {
		if spos == def.ObjSubUI || spos == def.ObjSubGame {
			for _, otype := range stype.List {
				/* Skips some items for WASM */
				if world.WASMMode && otype.ExcludeWASM {
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
	defer util.ReportPanic("DrawToolbar")
	ToolBarIconSize := float32(world.UIScale * def.ToolBarIconSize)
	ToolBarSpacing := float32(def.ToolBarIconSize / def.ToolBarSpaceRatio)

	toolbarCacheLock.Lock()
	defer toolbarCacheLock.Unlock()

	/* If needed, init image */
	if toolbarCache == nil {
		toolbarCache = ebiten.NewImage(int(ToolBarIconSize+ToolBarSpacing)*ToolbarMax+4, int(ToolBarIconSize+ToolBarSpacing))
	}
	/* Clear, full with semi-transparent */
	toolbarCache.Clear()
	toolbarCache.Fill(world.ColorToolTipBG)

	/* Loop through all toolbar items */
	for pos := 0; pos < ToolbarMax; pos++ {
		item := ToolbarItems[pos]
		x := float64(int(ToolBarIconSize+ToolBarSpacing) * int(pos))

		/* Get main image */
		img := item.OType.Images.Main

		/* If there is an overlay mode version, use that */
		if item.OType.Images.Overlay != nil {
			img = item.OType.Images.Overlay
		}
		/* If there is a toolbar-specific sprite, use that */
		if item.OType.Images.Toolbar != nil {
			img = item.OType.Images.Toolbar
		}
		/* Something went wrong, exit */
		if img == nil {
			return
		}

		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

		op.GeoM.Reset()
		iSize := img.Bounds()

		/* Handle non-square sprites */
		/* TODO: Get rid of this, just make toolbar sprites instead */
		var largerDim int
		if iSize.Size().X > largerDim {
			largerDim = iSize.Size().X
		}
		if iSize.Size().Y > largerDim {
			largerDim = iSize.Size().Y
		}

		/* Adjust image to toolbar size */
		op.GeoM.Scale(
			world.UIScale/(float64(largerDim)/float64(def.ToolBarIconSize)),
			world.UIScale/(float64(largerDim)/float64(def.ToolBarIconSize)))

		/* If set to, rotate sprite to direction */
		if item.OType.Rotatable && item.OType.Direction > 0 {
			x := float64(ToolBarIconSize / 2)
			y := float64(ToolBarIconSize / 2)

			/* center, rotate and move back... or we rotate on TL corner */
			op.GeoM.Translate(-x, -y)
			op.GeoM.Rotate(def.NinetyDeg * float64(item.OType.Direction))
			op.GeoM.Translate(x, y)
		}
		/* Move to correct location in toolbar image */
		op.GeoM.Translate((float64(ToolBarIconSize+(ToolBarSpacing))*float64(pos))+float64(ToolBarSpacing/2), float64(ToolBarSpacing/2))

		/* hovered/clicked icon highlight */
		if pos == index {
			if click {
				vector.DrawFilledRect(toolbarCache, float32(pos)*(ToolBarIconSize+ToolBarSpacing),
					0, ToolBarIconSize+ToolBarSpacing, ToolBarIconSize+ToolBarSpacing, world.ColorRed, false)
				ToolbarHover = true

				go func() {
					time.Sleep(time.Millisecond * 155)
					DrawToolbar(false, false, 0)
				}()
			} else if hover {
				vector.DrawFilledRect(toolbarCache, float32(pos)*(ToolBarIconSize+ToolBarSpacing),
					0, ToolBarIconSize+ToolBarSpacing, ToolBarIconSize+ToolBarSpacing, world.ColorAqua, false)
				ToolbarHover = true
			}

		}

		/* Draw to image */
		toolbarCache.DrawImage(img, op)

		/* Draw selection frame for selected game object */
		if item.SType == def.ObjSubGame {

			if item.OType.TypeI == SelectedItemType {
				/* Left */
				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(ToolBarIconSize+ToolBarSpacing),
					0,

					(def.TBSelThick),
					(ToolBarIconSize+ToolBarSpacing)-def.TBSelThick,
					world.ColorTBSelected, false)

				/* Top */
				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(ToolBarIconSize+ToolBarSpacing)+def.TBSelThick,
					0,

					(ToolBarIconSize+ToolBarSpacing)-def.TBSelThick,
					(def.TBSelThick),
					world.ColorTBSelected, false)

				/* Bottom */
				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(ToolBarIconSize+ToolBarSpacing)+def.TBSelThick,
					(ToolBarSpacing)+ToolBarIconSize-def.TBSelThick,

					(ToolBarIconSize+ToolBarSpacing)-def.TBSelThick,
					(def.TBSelThick),
					world.ColorTBSelected, false)

				/* Right */
				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(ToolBarIconSize+ToolBarSpacing)+ToolBarIconSize+ToolBarSpacing-def.TBSelThick,
					0,

					(def.TBSelThick),
					(ToolBarIconSize+ToolBarSpacing)-def.TBSelThick,
					world.ColorTBSelected, false)

			}
		}

		/* Show direction arrow, if this is a sprite we do not want to rotate */
		if item.OType.ToolBarArrow {
			var aop *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

			arrow := WorldOverlays[item.OType.Direction].Images.Main
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
