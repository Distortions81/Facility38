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
				if WASMMode && otype.excludeWASM {
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
	toolBarIconSize := float32(UIScale * ToolBarIconSize)
	toolBarSpacing := float32(ToolBarIconSize / toolBarSpaceRatio)

	toolbarCacheLock.Lock()
	defer toolbarCacheLock.Unlock()

	/* If needed, init image */
	if toolbarCache == nil {
		toolbarCache = ebiten.NewImage(int(toolBarIconSize+toolBarSpacing)*toolbarMax+4, int(toolBarIconSize+toolBarSpacing))
	}
	/* Clear, full with semi-transparent */
	toolbarCache.Clear()
	toolbarCache.Fill(ColorToolTipBG)

	/* Loop through all toolbar items */
	for pos := 0; pos < toolbarMax; pos++ {
		item := toolbarItems[pos]
		x := float64(int(toolBarIconSize+toolBarSpacing) * int(pos))

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
			UIScale/(float64(largerDim)/float64(ToolBarIconSize)),
			UIScale/(float64(largerDim)/float64(ToolBarIconSize)))

		/* If set to, rotate sprite to direction */
		if item.oType.rotatable && item.oType.direction > 0 {
			x := float64(toolBarIconSize / 2)
			y := float64(toolBarIconSize / 2)

			/* center, rotate and move back... or we rotate on TL corner */
			op.GeoM.Translate(-x, -y)
			op.GeoM.Rotate(ninetyDeg * float64(item.oType.direction))
			op.GeoM.Translate(x, y)
		}
		/* Move to correct location in toolbar image */
		op.GeoM.Translate((float64(toolBarIconSize+(toolBarSpacing))*float64(pos))+float64(toolBarSpacing/2), float64(toolBarSpacing/2))

		/* hovered/clicked icon highlight */
		if pos == index {
			if click {
				vector.DrawFilledRect(toolbarCache, float32(pos)*(toolBarIconSize+toolBarSpacing),
					0, toolBarIconSize+toolBarSpacing, toolBarIconSize+toolBarSpacing, ColorRed, false)
				toolbarHover = true

				go func() {
					time.Sleep(time.Millisecond * 155)
					drawToolbar(false, false, 0)
				}()
			} else if hover {
				vector.DrawFilledRect(toolbarCache, float32(pos)*(toolBarIconSize+toolBarSpacing),
					0, toolBarIconSize+toolBarSpacing, toolBarIconSize+toolBarSpacing, ColorAqua, false)
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
					float32(pos)*(toolBarIconSize+toolBarSpacing),
					0,

					(tbSelThick),
					(toolBarIconSize+toolBarSpacing)-tbSelThick,
					ColorTBSelected, false)

				/* Top */
				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(toolBarIconSize+toolBarSpacing)+tbSelThick,
					0,

					(toolBarIconSize+toolBarSpacing)-tbSelThick,
					(tbSelThick),
					ColorTBSelected, false)

				/* Bottom */
				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(toolBarIconSize+toolBarSpacing)+tbSelThick,
					(toolBarSpacing)+toolBarIconSize-tbSelThick,

					(toolBarIconSize+toolBarSpacing)-tbSelThick,
					(tbSelThick),
					ColorTBSelected, false)

				/* Right */
				vector.DrawFilledRect(toolbarCache,
					float32(pos)*(toolBarIconSize+toolBarSpacing)+toolBarIconSize+toolBarSpacing-tbSelThick,
					0,

					(tbSelThick),
					(toolBarIconSize+toolBarSpacing)-tbSelThick,
					ColorTBSelected, false)

			}
		}

		/* Show direction arrow, if this is a sprite we do not want to rotate */
		if item.oType.toolBarArrow {
			var aop *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

			arrow := worldOverlays[item.oType.direction].images.main
			if arrow != nil {
				if arrow.Bounds().Max.X != int(toolBarIconSize) {
					aop.GeoM.Scale(1.0/(float64(arrow.Bounds().Max.X)/float64(toolBarIconSize)),
						1.0/(float64(arrow.Bounds().Max.Y)/float64(toolBarIconSize)))
				}
				aop.GeoM.Translate(x, 0)
				aop.ColorScale.Scale(0.5, 0.5, 0.5, 0.66)
				toolbarCache.DrawImage(arrow, aop)
			}
		}
	}
}
