package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"runtime"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

const (
	wasmTerrainDiv  = 5
	maxBatch        = 15000
	infoWidth       = 128
	infoHeight      = 128
	infoSpaceRight  = 8
	infoSpaceTop    = 8
	infoPad         = 4
	batchGCInterval = time.Second * 30
)

var (
	/* Camera position */
	camXPos float32
	camYPos float32

	/* Previous camera position */
	lastCamX float32
	lastCamY float32

	/* Camera rect */
	camStartX uint16
	camStartY uint16
	camEndX   uint16
	camEndY   uint16

	/* Screen rect */
	screenStartX uint16
	screenStartY uint16
	screenEndX   uint16
	screenEndY   uint16
	frameCount   uint64

	/* Mouse position in world coords */
	worldMouseX float32
	worldMouseY float32

	/* Batch draw data */
	batchTop       int
	batchWatermark int
	batchGC        time.Time
	imageBatch     [maxBatch]*ebiten.Image
	opBatch        [maxBatch]*ebiten.DrawImageOptions

	consoleActive bool
)

/* Ebiten: Draw everything */
func (g *Game) Draw(screen *ebiten.Image) {
	defer reportPanic("Draw")
	frameCount++

	/* Boot/load/auth screen */
	if !authorized.Load() || !mapGenerated.Load() ||
		!spritesLoaded.Load() ||
		playerReady.Load() == 0 {

		bootScreen(screen)
		time.Sleep(time.Millisecond)
		return
	}

	/* Clear batch render items for GC */
	if frameCount%30 == 0 {
		if time.Since(batchGC) > batchGCInterval {
			batchGC = time.Now()

			for o := 0; o <= batchWatermark; o++ {
				imageBatch[o] = nil
				opBatch[o] = nil
			}
			batchWatermark = 0

		}
		if batchTop > batchWatermark {
			batchWatermark = batchTop
		}
	}
	batchTop = 0

	/* If needed, calculate object visibility */
	updateVisData()

	/* WASM terrain rendering */
	if wasmMode {
		if frameCount%wasmTerrainDiv == 0 {
			renderTerrainST()
		}
	} else {
		/* Standard rendering */
		renderTerrainST()
	}

	/* Draw modes */
	if zoomScale > mapPixelThreshold {
		/* Draw icon mode */
		drawIconMode()
	} else {
		/* Pixmap mode */
		drawPixmapMode()
	}

	/* Batch render everything */
	screen.Clear()
	for p := 0; p < batchTop; p++ {
		if imageBatch[p] == nil || opBatch[p] == nil {
			continue
		}
		screen.DrawImage(imageBatch[p], opBatch[p])
	}

	/* Not batched, draws on screen */
	drawTime(screen)
	drawItemPlacement(screen)

	/* Draw toolbar */
	toolbarCacheLock.RLock()
	screen.DrawImage(toolbarCache, nil)
	toolbarCacheLock.RUnlock()

	/* Tooltips */
	drawItemInfo(screen)

	/* Debug info */
	if infoLine {
		drawDebugInfo(screen)
	}

	/* Draw chat */
	drawChatLines(screen)

	/* Draw windows */
	drawOpenWindows(screen)

	/* Boot screen fade-out */
	if playerReady.Load() < 60 {
		bootScreen(screen)
	}
}

var lastVal int

func toolBarTooltip(screen *ebiten.Image, fmx int, fmy int) bool {
	defer reportPanic("toolBarTooltip")

	iconSize := float32(uiScale * toolBarIconSize)
	spacing := float32(iconSize / toolBarSpaceRatio)

	/* Calculate item */
	val := int(fmx / int(iconSize+spacing))

	/* Check if mouse is on top of the toolbar */
	if fmy <= int(iconSize) &&
		val < toolbarMax && val >= 0 {

		/* Calculate toolbar item */
		item := toolbarItems[val]
		var toolTip string

		/* Show item description if it exists */
		if item.oType.description != "" {

			/* Show item hot key if found */
			keyName := ""
			if item.oType.qKey != 0 {
				keyName = " ( " + item.oType.qKey.String() + " key )"
			}

			toolTip = fmt.Sprintf("%v\n%v\n%v", item.oType.name, item.oType.description, keyName)
		} else {
			/* Otherwise, just display item name */
			toolTip = fmt.Sprintf("%v\n", item.oType.name)
		}

		/* Draw text */
		drawText(toolTip, toolTipFont, color.White, ColorToolTipBG,
			XYf32{X: float32(fmx) + 10, Y: float32(fmy) + 10}, 11, screen,
			true, false, false)

		/* Don't redraw if item has not changed */
		if lastVal != val {
			lastVal = val
			drawToolbar(false, true, val)
		}
		return true
	}

	/* Not on toolbar, reset lastVal if needed */
	if lastVal != -1 {
		lastVal = -1
	}

	return false
}

func drawItemInfo(screen *ebiten.Image) {
	defer reportPanic("drawItemInfo")

	/* World Obj tool tip */
	pos := FloatXYToPosition(worldMouseX, worldMouseY)

	/* Handle toolbar */
	if toolBarTooltip(screen, MouseX, MouseY) {
		/* Mouse on toolbar, stop here */
		return
	}

	/* Erase toolbar hover highlight when mouse moves off */
	if toolbarHover {
		drawToolbar(false, false, 0)
		toolbarHover = false
	}

	/* Look for object */
	chunk := GetChunk(pos)

	toolTip := ""

	/* Mouse on world */
	if chunk != nil {
		b := GetObj(pos, chunk)
		/* Object found */
		if b != nil {
			o := b.obj

			/* Object name and position */
			toolTip = fmt.Sprintf("%v: %v\n",
				o.Unique.typeP.name,
				posToString(XY{X: uint16(worldMouseX), Y: uint16(worldMouseY)}))

			/* Show contents */
			if o.Unique.Contents != nil {
				for z := 0; z < MAT_MAX; z++ {
					if o.Unique.Contents.mats[z] != nil {
						toolTip = toolTip + fmt.Sprintf("Contents: %v: %v\n",
							o.Unique.Contents.mats[z].typeP.name, printUnit(o.Unique.Contents.mats[z]))
					}
				}
			}

			/* Show fuel */
			if o.Unique.typeP.machineSettings.maxFuelKG > 0 {
				toolTip = toolTip + fmt.Sprintf("Max Fuel: %v\n", printWeight(o.Unique.typeP.machineSettings.maxFuelKG))
				if o.Unique.KGFuel > o.Unique.typeP.machineSettings.kgFuelPerCycle {
					toolTip = toolTip + fmt.Sprintf("Fuel: %v\n", printWeight(o.Unique.KGFuel))
				} else {
					toolTip = toolTip + "NO FUEL\n"
				}
			}

			/* Show single contents */
			if o.Unique.SingleContent != nil && o.Unique.SingleContent.Amount > 0 {
				toolTip = toolTip + fmt.Sprintf("Contains: %v %v\n", printUnit(o.Unique.SingleContent), o.Unique.SingleContent.typeP.name)
			}

			/* Show miner on empty tile */
			if o.MinerData != nil && o.MinerData.resourcesCount == 0 {
				toolTip = toolTip + "NO SOLID RESOURCES TO MINE HERE!\n"
			}

			/* Debug info */
			if debugMode {
				if o.Unique.typeP.machineSettings.kgFuelPerCycle > 0 {
					toolTip = toolTip + fmt.Sprintf("Fuel per tock: %v\n", printWeight(o.Unique.typeP.machineSettings.kgFuelPerCycle))
				}
				if o.Unique.typeP.machineSettings.kgPerCycle > 0 {
					toolTip = toolTip + fmt.Sprintf("Per Cycle: %v\n", printWeight(o.Unique.typeP.machineSettings.kgPerCycle))
				}

				for _, p := range o.Ports {
					if p.obj == nil {
						continue
					}
					var typeString = "None"
					if p.Buf.typeP != nil {
						typeString = p.Buf.typeP.name
					}

					if p.Type == PORT_IN {
						toolTip = toolTip + fmt.Sprintf("Input: %v: %v: %v: %v (%v)\n",
							dirToName(uint8(p.Dir)),
							p.obj.Unique.typeP.name,
							typeString,
							printUnit(p.Buf),
							calcVolume(p.Buf))
					} else if p.Type == PORT_OUT {
						toolTip = toolTip + fmt.Sprintf("Output: %v: %v: %v: %v (%v)\n",
							dirToName(uint8(p.Dir)),
							p.obj.Unique.typeP.name,
							typeString,
							printUnit(p.Buf),
							calcVolume(p.Buf))
					} else if p.Type == PORT_FOUT {
						toolTip = toolTip + fmt.Sprintf("FuelOut: %v: %v: %v: %v (%v)\n",
							dirToName(uint8(p.Dir)),
							p.obj.Unique.typeP.name,
							typeString,
							printUnit(p.Buf),
							calcVolume(p.Buf))
					} else if p.Type == PORT_FIN {
						toolTip = toolTip + fmt.Sprintf("FuelIn: %v: %v: %v: %v (%v)\n",
							dirToName(uint8(p.Dir)),
							p.obj.Unique.typeP.name,
							typeString,
							printUnit(p.Buf),
							calcVolume(p.Buf))
					}
				}
				if o.Unique.typeP != nil {
					if o.Unique.typeP.machineSettings.maxContainKG > 0 {
						toolTip = toolTip + fmt.Sprintf("MaxContain: %v\n",
							printWeight(o.Unique.typeP.machineSettings.maxContainKG))
					}
					if o.Unique.typeP.machineSettings.kw > 0 {
						toolTip = toolTip + fmt.Sprintf("KW: %v\n", o.Unique.typeP.machineSettings.kw)
					}
					if o.hasTock {
						toolTip = toolTip + "Tocking\n"
					}
					if o.hasTick {
						toolTip = toolTip + "Ticking\n"
					}
				}

			}

			if o.Unique.typeP.description != "" {
				toolTip = toolTip + o.Unique.typeP.description
			}

			/* Draw item preview */
			if zoomScale < mapPixelThreshold || showResourceLayer {
				vector.DrawFilledRect(screen, float32(ScreenWidth)-(infoWidth)-infoSpaceRight-infoPad, infoSpaceTop, infoWidth+infoPad, infoHeight+infoPad, ColorToolTipBG, false)
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale((1.0/float64(o.Unique.typeP.size.X))*8.0, (1.0/float64(o.Unique.typeP.size.Y))*8.0)
				op.GeoM.Translate(float64(ScreenWidth)-(infoWidth)-infoSpaceRight, infoSpaceTop+(infoPad/2))
				screen.DrawImage(o.Unique.typeP.images.main, op)
			}
		} else {
			/* Otherwise, just show x/y location */
			toolTip = fmt.Sprintf("(%v, %v)",
				humanize.Comma(int64((worldMouseX - xyCenter))),
				humanize.Comma(int64((worldMouseY - xyCenter))))
		}
		drawText(toolTip, generalFont, color.White, ColorToolTipBG,
			XYf32{X: float32(ScreenWidth) - 8, Y: float32(ScreenHeight)},
			4, screen, false, true, false)
	}
	/* Tooltip for resources */
	if showResourceLayer {
		buf := ""
		/* Only recalculate if mouse moves */

		/* Get info for all layers */
		for p := 1; p < len(noiseLayers); p++ {
			var h float32 = float32(math.Abs(float64(noiseMap(float32(uint16(worldMouseX)), float32(uint16(worldMouseY)), p))))

			if h >= 0.01 {
				mat := MatData{Amount: h * KGPerTile, typeP: noiseLayers[p].typeP}
				if chunk != nil {
					tile := chunk.tileMap[XY{X: uint16(worldMouseX), Y: uint16(worldMouseY)}]
					if tile != nil {
						mat.Amount -= tile.minerData.mined[p] / KGPerTile
						if mat.Amount < 0 {
							mat.Amount = 0
						}
					}
				}
				buf = buf + fmt.Sprintf("%v: %v\n", noiseLayers[p].name, printUnit(&mat))

			}
		}

		if buf == "" {
			buf = "No resources found."
		}
		drawText("Yields:\n"+buf, toolTipFont, ColorAqua, ColorToolTipBG,
			XYf32{X: (float32(MouseX) + 20), Y: (float32(MouseY) + 20)}, 11, screen, true, true, false)
	}
}

func drawItemPlacement(screen *ebiten.Image) {
	defer reportPanic("drawItemPlacement")

	/* Draw ghost for selected item */
	if selectedItemType < maxItemType {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
		item := worldObjs[selectedItemType]

		/* camera + object */
		objOffX := camXPos + (float32(int(worldMouseX)))
		objOffY := camYPos + (float32(int(worldMouseY)))

		//Quick kludge for 1x3 object
		if item.size.Y == 3 && (item.direction == 1 || item.direction == 3) {
			objOffX++
			objOffY--
		}

		/* camera zoom */
		x := float64(objOffX * zoomScale)
		y := float64(objOffY * zoomScale)

		iSize := item.images.main.Bounds()
		if item.rotatable {
			xx := float64(iSize.Size().X / 2)
			yy := float64(iSize.Size().Y / 2)
			op.GeoM.Translate(-xx, -yy)
			op.GeoM.Rotate(ninetyDeg * float64(int(item.direction)))
			op.GeoM.Translate(xx, yy)
		}

		op.GeoM.Scale(
			(float64(item.size.X)*float64(zoomScale))/float64(iSize.Max.X),
			(float64(item.size.Y)*float64(zoomScale))/float64(iSize.Max.Y))

		op.GeoM.Translate(math.Floor(x), math.Floor(y))

		/* Tint red if we can't place item */
		blocked := false
		wPos := XY{X: uint16(worldMouseX), Y: uint16(worldMouseY)}
		/* Check if object fits */
		if item.multiTile {
			if !subObjFits(nil, item, false, wPos) {
				blocked = true
			}
		} else {
			tmpChunk := GetChunk(wPos)
			if GetObj(wPos, tmpChunk) != nil {
				blocked = true
			}
		}

		if blocked {
			op.ColorScale.Scale(0.5, 0.125, 0.125, 0.5)
		} else {
			op.ColorScale.Scale(0.5, 0.5, 0.5, 0.5)
		}

		img := item.images.main
		if item.images.overlay != nil {
			img = item.images.overlay
		}
		screen.DrawImage(img, op)
	}
}

/* Look at camera position, make a list of visible super-chunks and chunks. */
var visObj []*ObjData
var visChunk []*mapChunkData
var visSChunk []*mapSuperChunkData

func updateVisData() {
	defer reportPanic("updateVisData")

	/* When needed, make a list of chunks to draw */
	if visDataDirty.Load() {
		screenSizeLock.Lock()
		visDataDirty.Store(false)
		screenSizeLock.Unlock()

		visObj = []*ObjData{}
		visChunk = []*mapChunkData{}
		visSChunk = []*mapSuperChunkData{}

		/*
		* Calculate viewport
		* Moved to Update()
		 */

		superChunkListLock.RLock()
		for _, sChunk := range superChunkList {

			/* Is this super chunk on the screen? */
			if sChunk.pos.X < screenStartX/superChunkSize ||
				sChunk.pos.X > screenEndX/superChunkSize ||
				sChunk.pos.Y < screenStartY/superChunkSize ||
				sChunk.pos.Y > screenEndY/superChunkSize {
				sChunk.visible = false
				continue
			}

			visSChunk = append(visSChunk, sChunk)

			sChunk.visible = true

			for _, chunk := range sChunk.chunkList {

				/* Is this chunk on the screen? */
				if chunk.pos.X < screenStartX ||
					chunk.pos.X > screenEndX ||
					chunk.pos.Y < screenStartY ||
					chunk.pos.Y > screenEndY {
					chunk.visible = false
					continue
				}

				visChunk = append(visChunk, chunk)

				chunk.visible = true
				for _, obj := range chunk.objList {
					/* Is this object on the screen? */
					if obj.Pos.X < camStartX || obj.Pos.X > camEndX || obj.Pos.Y < camStartY || obj.Pos.Y > camEndY {
						continue
					}
					visObj = append(visObj, obj)
				}
			}
		}

		superChunkListLock.RUnlock()
	}
}

func drawTerrain(chunk *mapChunkData) (*ebiten.DrawImageOptions, *ebiten.Image) {
	defer reportPanic("drawTerrain")
	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

	/* Draw ground */
	chunk.terrainLock.RLock()
	cTmp := chunk.terrainImage
	if chunk.terrainImage == nil {
		cTmp = TempChunkImage
	}

	iSize := cTmp.Bounds().Size()
	op.GeoM.Reset()
	op.GeoM.Scale((chunkSize*float64(zoomScale))/float64(iSize.X),
		(chunkSize*float64(zoomScale))/float64(iSize.Y))
	op.GeoM.Translate((float64(camXPos)+float64(chunk.pos.X*chunkSize))*float64(zoomScale),
		(float64(camYPos)+float64(chunk.pos.Y*chunkSize))*float64(zoomScale))
	chunk.terrainLock.RUnlock()

	return op, cTmp
}

func drawIconMode() {
	defer reportPanic("drawIconMode")

	/* Draw ground */
	if showResourceLayer {
		drawPixmapMode()
	} else {
		for _, chunk := range visChunk {
			op, img := drawTerrain(chunk)
			if img != nil {
				opBatch[batchTop] = op
				imageBatch[batchTop] = img
				if batchTop < maxBatch {
					batchTop++
				} else {
					break
				}
			}
		}
	}

	/* Draw objects */
	for _, obj := range visObj {

		op, img := drawObject(obj, false)
		if img != nil {
			opBatch[batchTop] = op
			imageBatch[batchTop] = img
			if batchTop < maxBatch {
				batchTop++
			} else {
				break
			}
		}

		/* Overlays */
		/* Draw belt overlays */
		if obj.Unique.typeP.typeI == objTypeBasicBelt {

			/* Draw Input Materials */
			for _, port := range obj.Ports {
				if port.Buf.Amount > 0 {
					op, img = drawMaterials(port.Buf, obj, 0.8, 1.0, nil)
					if img != nil {
						opBatch[batchTop] = op
						imageBatch[batchTop] = img
						if batchTop < maxBatch {
							batchTop++
						} else {
							break
						}
					}
					break
				}
			}
		} else if obj.Unique.typeP.typeI == objTypeBasicBeltOver {
			/* Overpass belts */

			var start int32 = 32
			var middle int32 = 16
			var end int32 = 0

			found := false
			drewUnder := false

			/* Draw underpass */
			if obj.beltOver.underIn != nil {

				if obj.beltOver.underIn.Buf.Amount > 0 {
					/* Input */
					dir := rotatePosF64(XYs{X: 0, Y: middle}, obj.Dir, XYf64{X: 16, Y: 48})
					op, img := drawMaterials(obj.beltOver.underIn.Buf, obj, 0.8, 1.0, &dir)
					if img != nil {
						opBatch[batchTop] = op
						imageBatch[batchTop] = img

						found = true
						drewUnder = true

						if batchTop < maxBatch {
							batchTop++
						} else {
							break
						}
					}
				}
			}
			if !found && obj.beltOver.underOut != nil {

				/* Output */
				dir := rotatePosF64(XYs{X: 0, Y: middle}, obj.Dir, XYf64{X: 16, Y: 48})
				op, img := drawMaterials(obj.beltOver.underOut.Buf, obj, 0.8, 1.0, &dir)
				if img != nil {
					opBatch[batchTop] = op
					imageBatch[batchTop] = img

					drewUnder = true

					if batchTop < maxBatch {
						batchTop++
					} else {
						break
					}
				}
			}

			/* Draw mask for underpass */
			if drewUnder {
				opb, imgb := drawObject(obj, true)
				if img != nil {
					opBatch[batchTop] = opb
					imageBatch[batchTop] = imgb
					if batchTop < maxBatch {
						batchTop++
					} else {
						break
					}
				}
			}

			/* Draw overpass input */
			if obj.beltOver.overIn != nil {
				dir := rotatePosF64(XYs{X: 0, Y: start}, obj.Dir, XYf64{X: 16, Y: 48})
				op, img := drawMaterials(obj.beltOver.overIn.Buf, obj, 0.75, 1.0, &dir)
				if img != nil {
					opBatch[batchTop] = op
					imageBatch[batchTop] = img
					if batchTop < maxBatch {
						batchTop++
					} else {
						break
					}
				}
			}

			/* Draw overpass middle */
			if obj.beltOver.middle != nil {
				dir := rotatePosF64(XYs{X: 0, Y: middle}, obj.Dir, XYf64{X: 16, Y: 48})
				op, img := drawMaterials(obj.beltOver.middle, obj, 0.5, 1.0, &dir)
				if img != nil {
					opBatch[batchTop] = op
					imageBatch[batchTop] = img
					if batchTop < maxBatch {
						batchTop++
					} else {
						break
					}
				}
			}

			/* Draw overpass output */
			if obj.beltOver.overOut != nil {
				dir := rotatePosF64(XYs{X: 0, Y: end}, obj.Dir, XYf64{X: 16, Y: 48})
				op, img := drawMaterials(obj.beltOver.overOut.Buf, obj, 0.75, 1.0, &dir)
				if img != nil {
					opBatch[batchTop] = op
					imageBatch[batchTop] = img
					if batchTop < maxBatch {
						batchTop++
					} else {
						break
					}
				}
			}
		}

		/* camera + object */
		var objOffX float32 = camXPos + (float32(obj.Pos.X))
		var objOffY float32 = camYPos + (float32(obj.Pos.Y))

		/* camera zoom */
		objCamPosX := objOffX * zoomScale
		objCamPosY := objOffY * zoomScale

		/* Show blocked miners */
		if obj.Unique.typeP.typeI == objTypeBasicMiner && obj.blocked {
			img := worldOverlays[objOverlayBlocked].images.main

			iSize := img.Bounds()
			var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
			op.GeoM.Scale(((float64(1))*float64(zoomScale))/float64(iSize.Max.X),
				((float64(1))*float64(zoomScale))/float64(iSize.Max.Y))
			op.GeoM.Translate(float64(objCamPosX), float64(objCamPosY))

			if img != nil {
				opBatch[batchTop] = op
				imageBatch[batchTop] = img
				if batchTop < maxBatch {
					batchTop++
				} else {
					break
				}
			}
		}

		/* Draw overlays */
		if overlayMode {

			/* Show box contents in overlay mode */
			if obj.Unique.typeP.typeI == objTypeBasicBox {
				for _, cont := range obj.Unique.Contents.mats {
					if cont == nil {
						continue
					}
					op, img = drawMaterials(cont, obj, 0.5, 0.75, nil)
					if img != nil {
						opBatch[batchTop] = op
						imageBatch[batchTop] = img
						if batchTop < maxBatch {
							batchTop++
						} else {
							break
						}
					}
					break
				}
			}
			/* Show selected objects */
			if obj.selected {

				img := worldOverlays[objOverlaySel].images.main

				iSize := img.Bounds()
				var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
				op.GeoM.Scale(((float64(1))*float64(zoomScale))/float64(iSize.Max.X),
					((float64(1))*float64(zoomScale))/float64(iSize.Max.Y))
				op.GeoM.Translate(float64(objCamPosX), float64(objCamPosY))

				if img != nil {
					opBatch[batchTop] = op
					imageBatch[batchTop] = img
					if batchTop < maxBatch {
						batchTop++
					} else {
						break
					}

				}

			} else if obj.Unique.typeP.machineSettings.maxFuelKG > 0 && obj.Unique.KGFuel < obj.Unique.typeP.machineSettings.kgFuelPerCycle {
				/* Show objects with no fuel */

				img := worldOverlays[objOverlayNoFuel].images.main

				iSize := img.Bounds()
				var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
				op.GeoM.Scale(((float64(1))*float64(zoomScale))/float64(iSize.Max.X),
					((float64(1))*float64(zoomScale))/float64(iSize.Max.Y))
				op.GeoM.Translate(float64(objCamPosX), float64(objCamPosY))

				if img != nil {
					opBatch[batchTop] = op
					imageBatch[batchTop] = img
					if batchTop < maxBatch {
						batchTop++
					} else {
						break
					}

				}

			} else if obj.Unique.typeP.showArrow {
				/* Output arrows */
				for _, port := range obj.Ports {
					if port.Type == PORT_OUT && port.Dir == obj.Dir {

						img := worldOverlays[port.Dir].images.main
						iSize := img.Bounds()
						var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
						op.GeoM.Scale((1*float64(zoomScale))/float64(iSize.Max.X),
							((1)*float64(zoomScale))/float64(iSize.Max.Y))
						op.GeoM.Translate(float64(objCamPosX), float64(objCamPosY))
						op.ColorScale.Scale(0.5, 0.5, 0.5, 0.7)

						/* Draw Arrow */
						if img != nil {
							opBatch[batchTop] = op
							imageBatch[batchTop] = img
							if batchTop < maxBatch {
								batchTop++
							} else {
								break
							}
						}
						break
					}
				}
			}
		}
	}
}

func drawPixmapMode() {
	defer reportPanic("drawPixmapMode")

	/* Single thread render terrain for WASM */
	if wasmMode && frameCount%wasmTerrainDiv == 0 {
		resourceRenderDaemonST()
		pixmapRenderST()
	}
	/* Draw superChunk images (pixmap mode)*/
	superChunkListLock.RLock()
	for _, sChunk := range visSChunk {
		sChunk.pixelMapLock.Lock()
		if sChunk.pixelMap == nil {
			sChunk.pixelMapLock.Unlock()
			continue
		}

		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
		op.GeoM.Scale(
			(maxSuperChunk*float64(zoomScale))/float64(maxSuperChunk),
			(maxSuperChunk*float64(zoomScale))/float64(maxSuperChunk))

		op.GeoM.Translate(
			((float64(camXPos)+float64((sChunk.pos.X))*maxSuperChunk)*float64(zoomScale))-1,
			((float64(camYPos)+float64((sChunk.pos.Y))*maxSuperChunk)*float64(zoomScale))-1)

		if sChunk.pixelMap != nil {
			opBatch[batchTop] = op
			imageBatch[batchTop] = sChunk.pixelMap
			if batchTop < maxBatch {
				batchTop++
			} else {
				break
			}
		}
		sChunk.pixelMapLock.Unlock()
	}
	superChunkListLock.RUnlock()

	if showResourceLayer && resourceLegendImage != nil {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
		op.GeoM.Scale(2, 2)
		op.GeoM.Translate(8, float64(toolBarIconSize*uiScale)*2)
		if resourceLegendImage != nil {
			opBatch[batchTop] = op
			imageBatch[batchTop] = resourceLegendImage
			if batchTop < maxBatch {
				batchTop++
			}
		}
	}
}

func drawDebugInfo(screen *ebiten.Image) {
	defer reportPanic("drawDebugInfo")

	/* Draw debug info */
	buf := fmt.Sprintf("FPS: %-4v UPS: %4.2f Updates: %8v/%-8v/%8v/%-8v Blocks: %-8v/%-8v Draws: %-5v(%-5v) Arch: %v Build: v%v-%v",
		int(ebiten.ActualFPS()),
		(actualUPS),
		humanize.SIWithDigits(float64(tockCount), 2, ""),
		humanize.SIWithDigits(float64(activeTockCount), 2, ""),
		humanize.SIWithDigits(float64(tickCount), 2, ""),
		humanize.SIWithDigits(float64(activeTickCount), 2, ""),
		tickBlocks, tockBlocks,
		batchTop, batchWatermark,
		runtime.GOARCH, version, buildTime,
	)

	var pad float32 = 2 * float32(uiScale)
	drawText(buf, monoFont, color.White, ColorDebugBG,
		XYf32{X: 0, Y: float32(ScreenHeight) + (pad * 4.5)},
		pad, screen, true, true, false)

}

func drawTime(screen *ebiten.Image) {
	defer reportPanic("drawTime")

	gameClock := time.Duration((GameTick / uint64(objectUPS/2))) * time.Second
	/* Draw debug info */
	buf := fmt.Sprintf("Map Time: %v\nTime: %v",
		gameClock.String(),
		time.Now().Format("3:04PM"),
	)

	if debugMode {
		buf = buf + fmt.Sprintf("\n(%-4v,%-4v)", MouseX, MouseY)
	}
	var pad float32 = 8
	drawText(buf, monoFont, color.White, ColorDebugBG,
		XYf32{X: float32(ScreenWidth) - pad, Y: -pad},
		pad, screen, false, false, false)

}

func rectDrawText(input string, face font.Face, color color.Color, bgcolor color.Color, pos XYf32,
	pad float32, screen *ebiten.Image, justLeft bool, justUp bool, justCenter bool) image.Rectangle {
	defer reportPanic("DrawText")
	var tmx, tmy float32

	tRect := text.BoundString(face, input)

	if justCenter {
		tmx = float32(int(pos.X) - (tRect.Dx() / 2))
		tmy = float32(int(pos.Y) - (tRect.Dy() / 2))
	} else {
		if justLeft {
			tmx = float32(pos.X)
		} else {
			tmx = float32(int(pos.X) - tRect.Dx())
		}

		if justUp {
			tmy = float32(int(pos.Y) - tRect.Dy())
		} else {
			tmy = float32(pos.Y + float32(tRect.Dy()))
		}
	}

	fHeight := text.BoundString(face, "gpqabcABC!|_,;^*`")

	xPos := tmx - pad
	yPos := tmy - float32(fHeight.Dy()) - (float32(pad) / 2.0)
	xWidth := float32(tRect.Dx()) + pad*2
	yWidth := float32(tRect.Dy()) + pad*2
	vector.DrawFilledRect(
		screen, xPos, yPos,
		xWidth, yWidth, bgcolor, false,
	)
	text.Draw(screen, input, face, int(tmx), int(tmy), color)

	result := image.Rectangle{}
	result.Min.X = int(xPos)
	result.Min.Y = int(yPos)

	result.Max.X = int(xPos + xWidth)
	result.Max.Y = int(yPos + yWidth)
	return result

}

func drawText(input string, face font.Face, color color.Color, bgcolor color.Color, pos XYf32,
	pad float32, screen *ebiten.Image, justLeft bool, justUp bool, justCenter bool) XYf32 {
	defer reportPanic("DrawText")
	var tmx, tmy float32

	tRect := text.BoundString(face, input)

	if justCenter {
		tmx = float32(int(pos.X) - (tRect.Dx() / 2))
		tmy = float32(int(pos.Y) - (tRect.Dy() / 2))
	} else {
		if justLeft {
			tmx = float32(pos.X)
		} else {
			tmx = float32(int(pos.X) - tRect.Dx())
		}

		if justUp {
			tmy = float32(int(pos.Y) - tRect.Dy())
		} else {
			tmy = float32(pos.Y + float32(tRect.Dy()))
		}
	}
	_, _, _, alpha := bgcolor.RGBA()

	if alpha > 0 {
		fHeight := text.BoundString(face, "gpqabcABC!|_,;^*`")
		vector.DrawFilledRect(
			screen, tmx-pad, tmy-float32(fHeight.Dy())-(float32(pad)/2.0),
			float32(tRect.Dx())+pad*2, float32(tRect.Dy())+pad*2, bgcolor, false,
		)
	}
	text.Draw(screen, input, face, int(tmx), int(tmy), color)

	return XYf32{X: float32(tRect.Dx()) + pad, Y: float32(tRect.Dy()) + pad}
}

/* Draw world objects */
func drawObject(obj *ObjData, maskOnly bool) (op *ebiten.DrawImageOptions, img *ebiten.Image) {
	defer reportPanic("drawObject")
	/* camera + object */
	objOffX := camXPos + (float32(obj.Pos.X))
	objOffY := camYPos + (float32(obj.Pos.Y))

	//Quick kludge for 1x3 object
	if obj.Unique.typeP.size.Y == 3 {
		if obj.Dir == 1 || obj.Dir == 3 {
			objOffX++
			objOffY--
		}
	}

	/* camera zoom */
	x := float64(objOffX * zoomScale)
	y := float64(objOffY * zoomScale)

	/* Draw sprite */
	if obj.Unique.typeP.images.main == nil {
		return nil, nil
	} else {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

		iSize := obj.Unique.typeP.images.main.Bounds()

		if obj.isCorner {
			xx := float64(iSize.Size().X / 2)
			yy := float64(iSize.Size().Y / 2)
			op.GeoM.Translate(-xx, -yy)
			op.GeoM.Rotate(ninetyDeg * float64(int(obj.cornerDir)))
			op.GeoM.Translate(xx, yy)
		} else if obj.Unique.typeP.rotatable {
			xx := float64(iSize.Size().X / 2)
			yy := float64(iSize.Size().Y / 2)
			op.GeoM.Translate(-xx, -yy)
			op.GeoM.Rotate(ninetyDeg * float64(int(obj.Dir)))
			op.GeoM.Translate(xx, yy)
		}

		op.GeoM.Scale(
			(float64(obj.Unique.typeP.size.X)*float64(zoomScale))/float64(iSize.Max.X),
			(float64(obj.Unique.typeP.size.Y)*float64(zoomScale))/float64(iSize.Max.Y))

		op.GeoM.Translate(math.Floor(x), math.Floor(y))

		if obj.isCorner {
			return op, obj.Unique.typeP.images.corner
		} else if obj.active && obj.Unique.typeP.images.active != nil {
			return op, obj.Unique.typeP.images.active
		} else if overlayMode && obj.Unique.typeP.images.overlay != nil {
			return op, obj.Unique.typeP.images.overlay
		} else if maskOnly {
			return op, obj.Unique.typeP.images.mask
		} else {
			return op, obj.Unique.typeP.images.main
		}

	}
}

/* Update local vars with camera position calculations */
func calcScreenCamera() {
	defer reportPanic("calcScreenCamera")
	var padding float32 = 3 /* Set to max item size */

	lastCamX = camXPos
	lastCamY = camYPos

	/* Adjust cam position for zoom */
	camXPos = float32(-cameraX) + ((float32(ScreenWidth) / 2.0) / zoomScale)
	camYPos = float32(-cameraY) + ((float32(ScreenHeight) / 2.0) / zoomScale)

	/* Get camera bounds */
	camStartX = uint16((1/zoomScale + (cameraX - padding - (float32(ScreenWidth)/2.0)/zoomScale)))
	camStartY = uint16((1/zoomScale + (cameraY - padding - (float32(ScreenHeight)/2.0)/zoomScale)))
	camEndX = uint16((float32(ScreenWidth)/zoomScale + (cameraX + padding - (float32(ScreenWidth)/2.0)/zoomScale)))
	camEndY = uint16((float32(ScreenHeight)/zoomScale + (cameraY + padding - (float32(ScreenHeight)/2.0)/zoomScale)))

	/* Pre-calc camera chunk position */
	screenStartX = camStartX / chunkSize
	screenStartY = camStartY / chunkSize
	screenEndX = camEndX / chunkSize
	screenEndY = camEndY / chunkSize
}

/* Draw materials on belts */
func drawMaterials(m *MatData, obj *ObjData, scale float64, alpha float32, pos *XYf64) (op *ebiten.DrawImageOptions, img *ebiten.Image) {
	defer reportPanic("drawMaterials")
	if obj != nil && m.Amount > 0 {
		img := m.typeP.image
		if img != nil {

			/* camera + object */
			objOffX := camXPos + (float32(obj.Pos.X))
			objOffY := camYPos + (float32(obj.Pos.Y))

			/* camera zoom */
			objCamPosX := float64(objOffX * zoomScale)
			objCamPosY := float64(objOffY * zoomScale)

			iSize := img.Bounds()
			var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
			xx := float64(iSize.Dx()) / 2.0
			yy := float64(iSize.Dy()) / 2.0
			op.GeoM.Translate(-xx, -yy)
			op.GeoM.Rotate(float64(m.Rot) * ninetyDeg)
			if scale != 1 {
				op.GeoM.Scale(scale, scale)
			}
			op.GeoM.Translate(xx, yy)
			if pos != nil {
				op.GeoM.Translate(pos.X, pos.Y)
			}

			op.GeoM.Scale(
				((float64(1))*float64(zoomScale))/float64(iSize.Max.X),
				((float64(1))*float64(zoomScale))/float64(iSize.Max.Y))
			op.GeoM.Translate(objCamPosX, objCamPosY)
			op.ColorScale.ScaleAlpha(alpha)
			return op, img
		}
	}
	return nil, nil
}

var chatVertSpace float32 = 24.0 * float32(uiScale)

func drawChatLines(screen *ebiten.Image) {
	defer reportPanic("drawChatLines")
	var lineNum int
	chatLinesLock.Lock()
	defer chatLinesLock.Unlock()

	for x := chatLinesTop; x > 0 && lineNum < chatHeightLines; x-- {
		line := chatLines[x-1]
		/* Ignore old chat lines */
		since := time.Since(line.timestamp)
		if !consoleActive && since > line.lifetime {
			continue
		}
		lineNum++

		/* BG */
		tempBGColor := ColorToolTipBG
		/* Text color */
		r, g, b, _ := line.color.RGBA()

		/* Alpha + fade out */
		var blend float64 = 0
		if line.lifetime-since < chatFadeTime {
			blend = (float64(chatFadeTime-(line.lifetime-since)) / float64(chatFadeTime) * 100.0)
		}
		newAlpha := (254.0 - (blend * 2.55))
		oldAlpha := tempBGColor.A
		faded := newAlpha - float64(253.0-int(oldAlpha))
		if faded <= 0 {
			faded = 0
		} else if faded > 254 {
			faded = 254
		}
		tempBGColor.A = byte(faded)

		drawText(line.text, generalFont,
			color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: byte(newAlpha)},
			tempBGColor, XYf32{X: padding, Y: float32(ScreenHeight) - (float32(lineNum) * (float32(generalFontH) * 1.2)) - chatVertSpace},
			2, screen, true, false, false)
	}
}
