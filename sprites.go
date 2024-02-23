package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

/* Load all sprites, sub missing ones */
func loadSprites(dark bool) {
	defer reportPanic("loadSprites")
	darkStr := ""
	if dark {
		darkStr = "-dark"
	}

	for _, oType := range subTypes {
		for key, item := range oType.list {

			/* Main */
			img, err := getSpriteImage(oType.folder+"/"+item.base+darkStr+".png", false)

			/* If not found, check subfolder */
			if err != nil {
				img, err = getSpriteImage(oType.folder+"/"+item.base+"/"+item.base+darkStr+".png", false)
				if err != nil && !dark {
					/* If not found, fill texture with text */
					img = ebiten.NewImage(int(spriteScale), int(spriteScale))
					img.Fill(ColorVeryDarkGray)
					text.Draw(img, item.symbol, objectFont, placeholdOffX, placeholdOffY, color.White)
				}
			}
			if dark {
				oType.list[key].images.darkMain = img
			} else {
				oType.list[key].images.lightMain = img
			}

			/* Corner pieces */
			cornerImg, err := getSpriteImage(oType.folder+"/"+item.base+"/"+item.base+"-corner"+darkStr+".png", false)
			if err == nil {
				if dark {
					oType.list[key].images.darkCorner = cornerImg
				} else {
					oType.list[key].images.lightCorner = cornerImg
				}
			}

			/* Active*/
			activeImage, err := getSpriteImage(oType.folder+"/"+item.base+"/"+item.base+"-active"+darkStr+".png", false)
			if err == nil {
				if dark {
					oType.list[key].images.darkActive = activeImage
				} else {
					oType.list[key].images.lightActive = activeImage
				}
			}

			/* Overlays */
			overlayImg, err := getSpriteImage(oType.folder+"/"+item.base+"/"+item.base+"-overlay"+darkStr+".png", false)
			if err == nil {
				if dark {
					oType.list[key].images.darkOverlay = overlayImg
				} else {
					oType.list[key].images.lightOverlay = overlayImg
				}
			}
			/* Toolbar */
			toolbarImg, err := getSpriteImage(oType.folder+"/"+item.base+"/"+item.base+"-toolbar"+darkStr+".png", false)
			if err == nil {
				if dark {
					oType.list[key].images.darkToolbar = toolbarImg
				} else {
					oType.list[key].images.lightToolbar = toolbarImg
				}
			}

			/* Masks */
			maskImg, err := getSpriteImage(oType.folder+"/"+item.base+"/"+"-mask"+darkStr+".png", false)
			if err == nil {
				if dark {
					oType.list[key].images.lightMask = maskImg
				} else {
					oType.list[key].images.darkMask = maskImg
				}
			}

			wasmSleep()
		}
	}

	for m, item := range matTypes {
		if !dark {
			img, err := getSpriteImage("belt-obj/"+item.base+".png", false)
			if err != nil {
				/* If not found, fill texture with text */
				img = ebiten.NewImage(int(spriteScale), int(spriteScale))
				img.Fill(ColorVeryDarkGray)
				text.Draw(img, item.symbol, objectFont, placeholdOffX, placeholdOffY, color.White)
			}
			matTypes[m].lightImage = img
		} else {

			darkMatImg, err := getSpriteImage("belt-obj/"+item.base+"-dark.png", false)
			if err == nil {
				matTypes[m].darkImage = darkMatImg
				doLog(true, "loaded dark: %v", item.base)
			}
		}
		wasmSleep()
	}

	img, err := getSpriteImage("ui/resource-legend.png", true)
	if err == nil {
		resourceLegendImage = img
	}

	/* Loads boot screen assets */
	titleImg, err := getSpriteImage("title.png", true)
	if err == nil {
		TitleImage = titleImg
	}
	ebitenImg, err := getSpriteImage("ebiten.png", true)
	if err == nil {
		ebitenLogo = ebitenImg
	}

	linkSprites(false)
	linkSprites(true)

	setupTerrainCache()
	drawToolbar(false, false, 255)
	spritesLoaded.Store(true)
}

func linkSprites(dark bool) {
	defer reportPanic("LinkSprites")
	for _, oType := range subTypes {
		for key, item := range oType.list {
			if dark {
				if item.images.darkMain != nil {
					oType.list[key].images.main = item.images.darkMain
				}
				if item.images.darkToolbar != nil {
					oType.list[key].images.toolbar = item.images.darkToolbar
				}
				if item.images.darkMask != nil {
					oType.list[key].images.mask = item.images.darkMask
				}
				if item.images.darkActive != nil {
					oType.list[key].images.active = item.images.darkActive
				}
				if item.images.darkCorner != nil {
					oType.list[key].images.corner = item.images.darkCorner
				}
				if item.images.darkOverlay != nil {
					oType.list[key].images.overlay = item.images.darkOverlay
				}
				for m, item := range matTypes {
					if item.darkImage != nil {
						matTypes[m].image = matTypes[m].darkImage
					}
				}
			} else {
				if item.images.lightMain != nil {
					oType.list[key].images.main = item.images.lightMain
				}
				if item.images.lightToolbar != nil {
					oType.list[key].images.toolbar = item.images.lightToolbar
				}
				if item.images.lightMask != nil {
					oType.list[key].images.mask = item.images.lightMask
				}
				if item.images.lightActive != nil {
					oType.list[key].images.active = item.images.lightActive
				}
				if item.images.lightCorner != nil {
					oType.list[key].images.corner = item.images.lightCorner
				}
				if item.images.lightOverlay != nil {
					oType.list[key].images.overlay = item.images.lightOverlay
				}
				for m, item := range matTypes {
					if item.lightImage != nil {
						matTypes[m].image = matTypes[m].lightImage
					}
				}
			}
		}
	}
}
