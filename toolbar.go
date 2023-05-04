package main

import (
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	toolbarCache     *ebiten.Image
	toolbarCacheLock sync.RWMutex
	toolbarMax       int
	selectedItemType uint8 = maxItemType
	toolbarItems           = []toolbarItem{}

	toolbarHover bool
)

/* Make default toolbar list */
func initToolbar() {
	defer reportPanic("InitToolbar")
	toolbarMax = 0
	for spos, stype := range subTypes {
		if spos == objSubUI || spos == objSubGame {
			for _, otype := range stype.list {
				/* Skips some items for WASM */
				if wasmMode && otype.excludeWASM {
					continue
				}
				toolbarMax++
				toolbarItems = append(toolbarItems, toolbarItem{sType: spos, oType: otype})

			}
		}
	}
}

/* Draw toolbar to an image */
func drawToolbar(click, hover bool, index int) {
	defer reportPanic("drawToolbar")
	iconSize := float32(uiScale * toolBarIconSize)
	spacing := float32(toolBarIconSize / toolBarSpaceRatio)

	toolbarCacheLock.Lock()
	defer toolbarCacheLock.Unlock()

	/* If needed, init image */
	if toolbarCache == nil {
		toolbarCache = ebiten.NewImage(int(iconSize+spacing)*toolbarMax+4, int(iconSize+spacing))
	}
	/* Clear, full with semi-transparent */
	toolbarCache.Clear()
	toolbarCache.Fill(ColorToolTipBG)

	/* Loop through all toolbar items */
	for pos := 0; pos < toolbarMax; pos++ {
		item := toolbarItems[pos]
		x := float64(int(iconSize+spacing) * int(pos))

		/* Get main image */
		img := item.oType.images.main

		/* If there is an overlay mode version, use that */
		if item.oType.images.overlay != nil {
			img = item.oType.images.overlay
		}
		/* If there is a toolbar-specific sprite, use that */
		if item.oType.images.toolbar != nil {
			img = item.oType.images.toolbar
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
			uiScale/(float64(largerDim)/float64(toolBarIconSize)),
			uiScale/(float64(largerDim)/float64(toolBarIconSize)))

		/* If set to, rotate sprite to direction */
		if item.oType.rotatable && item.oType.direction > 0 {
			x := float64(iconSize / 2)
			y := float64(iconSize / 2)

			/* center, rotate and move back... or we rotate on TL corner */
			op.GeoM.Translate(-x, -y)
			op.GeoM.Rotate(ninetyDeg * float64(item.oType.direction))
			op.GeoM.Translate(x, y)
		}
		/* Move to correct location in toolbar image */
		op.GeoM.Translate((float64(iconSize+(spacing))*float64(pos))+float64(spacing/2), float64(spacing/2))

		/* hovered/clicked icon highlight */
		if pos == index {
			if click {
				vector.DrawFilledRect(toolbarCache, float32(pos)*(iconSize+spacing),
					0, iconSize+spacing, iconSize+spacing, ColorRed, false)
				toolbarHover = true

				go func() {
					time.Sleep(time.Millisecond * 155)
					drawToolbar(false, false, 0)
				}()
			} else if hover {
				vector.DrawFilledRect(toolbarCache, float32(pos)*(iconSize+spacing),
					0, iconSize+spacing, iconSize+spacing, ColorAqua, false)
				toolbarHover = true
			}

		}

		/* Draw to image */
		toolbarCache.DrawImage(img, op)

		/* Draw selection frame for selected game object */
		if item.sType == objSubGame {

			if item.oType.typeI == selectedItemType {
				/* Left */
				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(iconSize+spacing),
					0,

					(tbSelThick),
					(iconSize+spacing)-tbSelThick,
					ColorTBSelected, false)

				/* Top */
				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(iconSize+spacing)+tbSelThick,
					0,

					(iconSize+spacing)-tbSelThick,
					(tbSelThick),
					ColorTBSelected, false)

				/* Bottom */
				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(iconSize+spacing)+tbSelThick,
					(spacing)+iconSize-tbSelThick,

					(iconSize+spacing)-tbSelThick,
					(tbSelThick),
					ColorTBSelected, false)

				/* Right */
				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(iconSize+spacing)+iconSize+spacing-tbSelThick,
					0,

					(tbSelThick),
					(iconSize+spacing)-tbSelThick,
					ColorTBSelected, false)

			}
		}

		/* Show direction arrow, if this is a sprite we do not want to rotate */
		if item.oType.toolBarArrow {
			var aop *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

			arrow := worldOverlays[item.oType.direction].images.main
			if arrow != nil {
				if arrow.Bounds().Max.X != int(iconSize) {
					aop.GeoM.Scale(1.0/(float64(arrow.Bounds().Max.X)/float64(iconSize)),
						1.0/(float64(arrow.Bounds().Max.Y)/float64(iconSize)))
				}
				aop.GeoM.Translate(x, 0)
				aop.ColorScale.Scale(0.5, 0.5, 0.5, 0.66)
				toolbarCache.DrawImage(arrow, aop)
			}
		}
	}
}
