package main

import (
	"Facility38/gv"
	"Facility38/util"
	"Facility38/world"
	"fmt"
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
	cBlockedIndicatorOffset = 0
	WASMTerrainDiv          = 5
	MaxBatch                = 10000
	infoWidth               = 128
	infoHeight              = 128
	infoSpaceRight          = 8
	infoSpaceTop            = 8
	infoPad                 = 4
)

var (
	camXPos float32
	camYPos float32

	camStartX uint16
	camStartY uint16
	camEndX   uint16
	camEndY   uint16

	screenStartX uint16
	screenStartY uint16
	screenEndX   uint16
	screenEndY   uint16
	frameCount   uint64

	WorldMouseX float32
	WorldMouseY float32

	lastResourceString string
	ConsoleActive      bool

	lastCamX float32
	lastCamY float32

	BatchTop   int
	ImageBatch [MaxBatch]*ebiten.Image
	OpBatch    [MaxBatch]*ebiten.DrawImageOptions
	UILayer    *ebiten.Image
	UIScale    float32 = 0.5
)

/* Setup a few images for later use */
func init() {
	world.MiniMapTile = ebiten.NewImage(1, 1)
	world.MiniMapTile.Fill(color.White)
}

/* Ebiten: Draw everything */
func (g *Game) Draw(screen *ebiten.Image) {

	/* Boot screen */
	if !world.MapGenerated.Load() ||
		!world.SpritesLoaded.Load() ||
		!world.PlayerReady.Load() {

		bootScreen(screen)
		drawChatLines(screen)
		time.Sleep(time.Millisecond)
		return
	}

	frameCount++

	/* If needed, calculate object visibility */
	updateVisData()

	/* WASM terrain rendering */
	if gv.WASMMode {
		if frameCount%WASMTerrainDiv == 0 {
			RenderTerrainST()
		}
	} else {
		RenderTerrainST()
	}

	/* Draw modes */
	if world.ZoomScale > gv.MapPixelThreshold { /* Draw icon mode */
		drawIconMode(screen)
	} else {
		drawPixmapMode(screen)
	}

	drawItemPlacement(screen)

	/* Draw toolbar */
	toolbarCacheLock.RLock()
	screen.DrawImage(toolbarCache, nil)
	toolbarCacheLock.RUnlock()

	/* Tooltips */
	drawItemInfo(screen)

	/* Debug info */
	if world.InfoLine {
		drawDebugInfo(screen)
	}

	drawChatLines(screen)
	DrawOpenWindows(screen)
}

var lastVal int

func toolBarTooltip(screen *ebiten.Image, fmx int, fmy int) bool {
	ToolBarIconSize := float32(gv.UIScale * gv.ToolBarIconSize)
	ToolBarSpacing := float32(gv.ToolBarIconSize / gv.ToolBarSpaceRatio)

	/* Calculate item */
	val := int(fmx / int(ToolBarIconSize+ToolBarSpacing))

	/* Check if mouse is on top of the toolbar */
	if fmy <= int(ToolBarIconSize) &&
		val < ToolbarMax && val >= 0 {

		/* Calculate toolbar item */
		item := ToolbarItems[val]
		var toolTip string

		/* Show item description if it exists */
		if item.OType.Description != "" {

			/* Show item hot key if found */
			keyName := ""
			if item.OType.QKey != 0 {
				keyName = " ( " + item.OType.QKey.String() + " key )"
			}

			toolTip = fmt.Sprintf("%v\n%v\n%v", item.OType.Name, item.OType.Description, keyName)
		} else {
			/* Otherwise, just display item name */
			toolTip = fmt.Sprintf("%v\n", item.OType.Name)
		}

		/* Draw text */
		DrawText(toolTip, world.ToolTipFont, world.ColorWhite, world.ColorToolTipBG,
			world.XYf32{X: float32(fmx) + 10, Y: float32(fmy) + 10}, 11, screen,
			true, false, false)

		/* Don't redraw if item has not changed */
		if lastVal != val {
			lastVal = val
			DrawToolbar(false, true, val)
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

	/* World Obj tool tip */
	pos := util.FloatXYToPosition(WorldMouseX, WorldMouseY)

	/* Handle toolbar */
	if toolBarTooltip(screen, MouseX, MouseY) {
		/* Mouse on toolbar, stop here */
		return
	}

	/* Erase toolbar hover highlight when mouse moves off */
	if ToolbarHover {
		DrawToolbar(false, false, 0)
		ToolbarHover = false
	}

	/* Look for object */
	chunk := util.GetChunk(pos)

	toolTip := ""

	/* Mouse on world */
	if chunk != nil {
		b := util.GetObj(pos, chunk)
		/* Object found */
		if b != nil {
			o := b.Obj

			/* Object name and position */
			toolTip = fmt.Sprintf("%v: %v\n",
				o.Unique.TypeP.Name,
				util.PosToString(world.XY{X: uint16(WorldMouseX), Y: uint16(WorldMouseY)}))

			/* Show contents */
			if o.Unique.Contents != nil {
				for z := 0; z < gv.MAT_MAX; z++ {
					if o.Unique.Contents.Mats[z] != nil {
						toolTip = toolTip + fmt.Sprintf("Contents: %v: %v\n",
							o.Unique.Contents.Mats[z].TypeP.Name, PrintUnit(o.Unique.Contents.Mats[z]))
					}
				}
			}

			/* Show fuel */
			if o.Unique.TypeP.MachineSettings.MaxFuelKG > 0 {
				toolTip = toolTip + fmt.Sprintf("Max Fuel: %v\n", PrintWeight(o.Unique.TypeP.MachineSettings.MaxFuelKG))
				if o.Unique.KGFuel > o.Unique.TypeP.MachineSettings.KgFuelPerCycle {
					toolTip = toolTip + fmt.Sprintf("Fuel: %v\n", PrintWeight(o.Unique.KGFuel))
				} else {
					toolTip = toolTip + "NO FUEL\n"
				}
			}

			/* Show single contents */
			if o.Unique.SingleContent != nil && o.Unique.SingleContent.Amount > 0 {
				toolTip = toolTip + fmt.Sprintf("Contains: %v %v\n", PrintUnit(o.Unique.SingleContent), o.Unique.SingleContent.TypeP.Name)
			}

			/* Show if blocked */
			if o.Blocked {
				toolTip = toolTip + "BLOCKED\n"
			}

			/* Show miner on empty tile */
			if o.MinerData != nil && o.MinerData.ResourcesCount == 0 {
				toolTip = toolTip + "NOTHING TO MINE.\n"
			}

			/* Debug info */
			if gv.Debug {
				if o.Unique.TypeP.MachineSettings.KgFuelPerCycle > 0 {
					toolTip = toolTip + fmt.Sprintf("Fuel per tock: %v\n", PrintWeight(o.Unique.TypeP.MachineSettings.KgFuelPerCycle))
				}
				if o.Unique.TypeP.MachineSettings.KgPerCycle > 0 {
					toolTip = toolTip + fmt.Sprintf("Per Cycle: %v\n", PrintWeight(o.Unique.TypeP.MachineSettings.KgPerCycle))
				}
				if o.Unique.TypeP.MachineSettings.MaxContainKG > 0 {
					toolTip = toolTip + fmt.Sprintf("Max contents: %v\n", PrintWeight(o.Unique.TypeP.MachineSettings.MaxContainKG))
				}

				for _, p := range o.Ports {
					if p.Obj == nil {
						continue
					}
					var tstring = "None"
					if p.Buf.TypeP != nil {
						tstring = p.Buf.TypeP.Name
					}

					if p.Type == gv.PORT_IN {
						toolTip = toolTip + fmt.Sprintf("Input: %v: %v: %v: %v (%v)\n",
							util.DirToName(uint8(p.Dir)),
							p.Obj.Unique.TypeP.Name,
							tstring,
							PrintUnit(p.Buf),
							CalcVolume(p.Buf))
					} else if p.Type == gv.PORT_OUT {
						toolTip = toolTip + fmt.Sprintf("Output: %v: %v: %v: %v (%v)\n",
							util.DirToName(uint8(p.Dir)),
							p.Obj.Unique.TypeP.Name,
							tstring,
							PrintUnit(p.Buf),
							CalcVolume(p.Buf))
					} else if p.Type == gv.PORT_FOUT {
						toolTip = toolTip + fmt.Sprintf("FuelOut: %v: %v: %v: %v (%v)\n",
							util.DirToName(uint8(p.Dir)),
							p.Obj.Unique.TypeP.Name,
							tstring,
							PrintUnit(p.Buf),
							CalcVolume(p.Buf))
					} else if p.Type == gv.PORT_FIN {
						toolTip = toolTip + fmt.Sprintf("FuelIn: %v: %v: %v: %v (%v)\n",
							util.DirToName(uint8(p.Dir)),
							p.Obj.Unique.TypeP.Name,
							tstring,
							PrintUnit(p.Buf),
							CalcVolume(p.Buf))
					}
				}
				if o.Unique.TypeP != nil {
					if o.Unique.TypeP.MachineSettings.MaxContainKG > 0 {
						toolTip = toolTip + fmt.Sprintf("MaxContain: %v\n",
							PrintWeight(o.Unique.TypeP.MachineSettings.MaxContainKG))
					}
					if o.Unique.TypeP.MachineSettings.KW > 0 {
						toolTip = toolTip + fmt.Sprintf("KW: %v\n", o.Unique.TypeP.MachineSettings.KW)
					}
					if o.HasTock {
						toolTip = toolTip + "Tocking\n"
					}
					if o.HasTick {
						toolTip = toolTip + "Ticking\n"
					}
				}

			}

			if o.Unique.TypeP.Description != "" {
				toolTip = toolTip + o.Unique.TypeP.Description + "\n"
			}

			vector.DrawFilledRect(screen, float32(world.ScreenWidth)-(infoWidth)-infoSpaceRight-infoPad, infoSpaceTop, infoWidth+infoPad, infoHeight+infoPad, world.ColorToolTipBG, false)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale((1.0/float64(o.Unique.TypeP.Size.X))*8.0, (1.0/float64(o.Unique.TypeP.Size.Y))*8.0)
			op.GeoM.Translate(float64(world.ScreenWidth)-(infoWidth)-infoSpaceRight, infoSpaceTop+(infoPad/2))
			screen.DrawImage(o.Unique.TypeP.Images.Main, op)
		} else {
			/* Otherwise, just show x/y location */
			toolTip = fmt.Sprintf("(%v, %v)",
				humanize.Comma(int64((WorldMouseX - gv.XYCenter))),
				humanize.Comma(int64((WorldMouseY - gv.XYCenter))))
		}
		DrawText(toolTip, world.ToolTipFont, color.White, world.ColorToolTipBG,
			world.XYf32{X: float32(world.ScreenWidth), Y: float32(world.ScreenHeight)},
			11, screen, false, true, false)
	}
	/* Tooltip for resources */
	if world.ShowResourceLayer {
		buf := ""
		/* Only recalculate if mouse moves */
		if MouseX != LastMouseX || MouseY != LastMouseY || camXPos != lastCamX || camYPos != lastCamY {

			/* Get info for all layers */
			for p := 1; p < len(NoiseLayers); p++ {
				var h float32 = float32(math.Abs(float64(NoiseMap(WorldMouseX, WorldMouseY, p))))

				if h >= 0.0001 {
					buf = buf + fmt.Sprintf("%v: %0.2f%%\n", NoiseLayers[p].Name, util.Min(h*100.0, 100.0))
				}
			}
		} else {
			/* save a bit of processing */
			buf = lastResourceString
		}
		if buf != "" {
			DrawText("Yields:\n"+buf, world.ToolTipFont, world.ColorAqua, world.ColorToolTipBG,
				world.XYf32{X: (float32(MouseX) + 20), Y: (float32(MouseY) + 20)}, 11, screen, true, true, false)
		}
		lastResourceString = buf
	}
}

func drawItemPlacement(screen *ebiten.Image) {
	/* Draw ghost for selected item */
	if SelectedItemType < gv.MaxItemType {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		item := WorldObjs[SelectedItemType]

		/* camera + object */
		objOffX := camXPos + (float32(int(WorldMouseX)))
		objOffY := camYPos + (float32(int(WorldMouseY)))

		//Quick kludge for 1x3 object
		if item.Size.Y == 3 && (item.Direction == 1 || item.Direction == 3) {
			objOffX++
			objOffY--
		}

		/* camera zoom */
		x := float64(objOffX * world.ZoomScale)
		y := float64(objOffY * world.ZoomScale)

		iSize := item.Images.Main.Bounds()
		if item.Rotatable {
			xx := float64(iSize.Size().X / 2)
			yy := float64(iSize.Size().Y / 2)
			op.GeoM.Translate(-xx, -yy)
			op.GeoM.Rotate(gv.NinetyDeg * float64(int(item.Direction)))
			op.GeoM.Translate(xx, yy)
		}

		op.GeoM.Scale(
			(float64(item.Size.X)*float64(world.ZoomScale))/float64(iSize.Max.X),
			(float64(item.Size.Y)*float64(world.ZoomScale))/float64(iSize.Max.Y))

		op.GeoM.Translate(math.Floor(x), math.Floor(y))

		/* Tint red if we can't place item */
		blocked := false
		wPos := world.XY{X: uint16(WorldMouseX), Y: uint16(WorldMouseY)}
		/* Check if object fits */
		if item.MultiTile {
			if !SubObjFits(nil, item, false, wPos) {
				blocked = true
			}
		} else {
			tchunk := util.GetChunk(wPos)
			if util.GetObj(wPos, tchunk) != nil {
				blocked = true
			}
		}

		if blocked {
			op.ColorScale.Scale(0.5, 0.125, 0.125, 0.5)
		} else {
			op.ColorScale.Scale(0.5, 0.5, 0.5, 0.5)
		}

		img := item.Images.Main
		if item.Images.Overlay != nil {
			img = item.Images.Overlay
		}
		screen.DrawImage(img, op)
	}
}

/* Look at camera position, make a list of visible superchunks and chunks. */
var VisObj []*world.ObjData
var VisChunk []*world.MapChunk
var VisSChunk []*world.MapSuperChunk

func updateVisData() {

	/* When needed, make a list of chunks to draw */
	if world.VisDataDirty.Load() {
		VisObj = []*world.ObjData{}
		VisChunk = []*world.MapChunk{}
		VisSChunk = []*world.MapSuperChunk{}

		/*
		* Calculate viewport
		* Moved to Update()
		 */

		world.SuperChunkListLock.RLock()
		for _, sChunk := range world.SuperChunkList {

			/* Is this super chunk on the screen? */
			if sChunk.Pos.X < screenStartX/gv.SuperChunkSize ||
				sChunk.Pos.X > screenEndX/gv.SuperChunkSize ||
				sChunk.Pos.Y < screenStartY/gv.SuperChunkSize ||
				sChunk.Pos.Y > screenEndY/gv.SuperChunkSize {
				sChunk.Visible = false
				continue
			}

			VisSChunk = append(VisSChunk, sChunk)

			sChunk.Visible = true

			for _, chunk := range sChunk.ChunkList {

				/* Is this chunk on the screen? */
				if chunk.Pos.X < screenStartX ||
					chunk.Pos.X > screenEndX ||
					chunk.Pos.Y < screenStartY ||
					chunk.Pos.Y > screenEndY {
					chunk.Visible = false
					continue
				}

				VisChunk = append(VisChunk, chunk)

				chunk.Visible = true
				for _, obj := range chunk.ObjList {
					/* Is this object on the screen? */
					if obj.Pos.X < camStartX || obj.Pos.X > camEndX || obj.Pos.Y < camStartY || obj.Pos.Y > camEndY {
						continue
					}
					VisObj = append(VisObj, obj)
				}
			}
		}
		world.VisDataDirty.Store(false)
		world.SuperChunkListLock.RUnlock()
	}
}

func drawTerrain(chunk *world.MapChunk) (*ebiten.DrawImageOptions, *ebiten.Image) {
	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

	/* Draw ground */
	chunk.TerrainLock.RLock()
	cTmp := chunk.TerrainImage
	if chunk.TerrainImage == nil {
		cTmp = world.TempChunkImage
	}

	iSize := cTmp.Bounds().Size()
	op.GeoM.Reset()
	op.GeoM.Scale((gv.ChunkSize*float64(world.ZoomScale))/float64(iSize.X),
		(gv.ChunkSize*float64(world.ZoomScale))/float64(iSize.Y))
	op.GeoM.Translate((float64(camXPos)+float64(chunk.Pos.X*gv.ChunkSize))*float64(world.ZoomScale),
		(float64(camYPos)+float64(chunk.Pos.Y*gv.ChunkSize))*float64(world.ZoomScale))
	chunk.TerrainLock.RUnlock()

	return op, cTmp
}

func drawIconMode(screen *ebiten.Image) {

	BatchTop = 0

	/* Draw ground */
	if world.ShowResourceLayer {
		drawPixmapMode(screen)
	} else {
		for _, chunk := range VisChunk {
			op, img := drawTerrain(chunk)
			if img != nil {
				OpBatch[BatchTop] = op
				ImageBatch[BatchTop] = img
				if BatchTop < MaxBatch {
					BatchTop++
				} else {
					break
				}
			}
		}
	}

	/* Draw objects */
	for _, obj := range VisObj {

		op, img := drawObject(screen, obj, false)
		if img != nil {
			OpBatch[BatchTop] = op
			ImageBatch[BatchTop] = img
			if BatchTop < MaxBatch {
				BatchTop++
			} else {
				break
			}
		}

		/* Overlays */
		/* Draw belt overlays */
		if obj.Unique.TypeP.TypeI == gv.ObjTypeBasicBelt {

			/* Draw Input Materials */
			for _, port := range obj.Ports {
				if port.Buf.Amount > 0 {
					op, img = drawMaterials(port.Buf, obj, screen, 0.8, 1.0, nil)
					if img != nil {
						OpBatch[BatchTop] = op
						ImageBatch[BatchTop] = img
						if BatchTop < MaxBatch {
							BatchTop++
						} else {
							break
						}
					}
					break
				}
			}
		} else if obj.Unique.TypeP.TypeI == gv.ObjTypeBasicBeltOver {
			/* Overpass belts */

			var start int32 = 32
			var middle int32 = 16
			var end int32 = 0

			found := false
			drewUnder := false

			/* Draw underpass */
			if obj.BeltOver.UnderIn != nil {

				if obj.BeltOver.UnderIn.Buf.Amount > 0 {
					/* Input */
					dir := RotatePosF64(world.XYs{X: 0, Y: middle}, obj.Dir, world.XYf64{X: 16, Y: 48})
					op, img := drawMaterials(obj.BeltOver.UnderIn.Buf, obj, screen, 0.8, 1.0, &dir)
					if img != nil {
						OpBatch[BatchTop] = op
						ImageBatch[BatchTop] = img

						found = true
						drewUnder = true

						if BatchTop < MaxBatch {
							BatchTop++
						} else {
							break
						}
					}
				}
			}
			if !found && obj.BeltOver.UnderOut != nil {

				/* Output */
				dir := RotatePosF64(world.XYs{X: 0, Y: middle}, obj.Dir, world.XYf64{X: 16, Y: 48})
				op, img := drawMaterials(obj.BeltOver.UnderOut.Buf, obj, screen, 0.8, 1.0, &dir)
				if img != nil {
					OpBatch[BatchTop] = op
					ImageBatch[BatchTop] = img

					drewUnder = true

					if BatchTop < MaxBatch {
						BatchTop++
					} else {
						break
					}
				}
			}

			/* Draw mask for underpass */
			if drewUnder {
				opb, imgb := drawObject(screen, obj, true)
				if img != nil {
					OpBatch[BatchTop] = opb
					ImageBatch[BatchTop] = imgb
					if BatchTop < MaxBatch {
						BatchTop++
					} else {
						break
					}
				}
			}

			/* Draw overpass input */
			if obj.BeltOver.OverIn != nil {
				dir := RotatePosF64(world.XYs{X: 0, Y: start}, obj.Dir, world.XYf64{X: 16, Y: 48})
				op, img := drawMaterials(obj.BeltOver.OverIn.Buf, obj, screen, 0.75, 1.0, &dir)
				if img != nil {
					OpBatch[BatchTop] = op
					ImageBatch[BatchTop] = img
					if BatchTop < MaxBatch {
						BatchTop++
					} else {
						break
					}
				}
			}

			/* Draw overpass middle */
			if obj.BeltOver.Middle != nil {
				dir := RotatePosF64(world.XYs{X: 0, Y: middle}, obj.Dir, world.XYf64{X: 16, Y: 48})
				op, img := drawMaterials(obj.BeltOver.Middle, obj, screen, 0.5, 1.0, &dir)
				if img != nil {
					OpBatch[BatchTop] = op
					ImageBatch[BatchTop] = img
					if BatchTop < MaxBatch {
						BatchTop++
					} else {
						break
					}
				}
			}

			/* Draw overpass output */
			if obj.BeltOver.OverOut != nil {
				dir := RotatePosF64(world.XYs{X: 0, Y: end}, obj.Dir, world.XYf64{X: 16, Y: 48})
				op, img := drawMaterials(obj.BeltOver.OverOut.Buf, obj, screen, 0.75, 1.0, &dir)
				if img != nil {
					OpBatch[BatchTop] = op
					ImageBatch[BatchTop] = img
					if BatchTop < MaxBatch {
						BatchTop++
					} else {
						break
					}
				}
			}
		}
		/* Draw overlays */
		if world.OverlayMode {

			/* Show box conents in overylay mode */
			if obj.Unique.TypeP.TypeI == gv.ObjTypeBasicBox {
				for _, cont := range obj.Unique.Contents.Mats {
					if cont == nil {
						continue
					}
					op, img = drawMaterials(cont, obj, screen, 0.5, 0.75, nil)
					if img != nil {
						OpBatch[BatchTop] = op
						ImageBatch[BatchTop] = img
						if BatchTop < MaxBatch {
							BatchTop++
						} else {
							break
						}
					}
					break
				}
			}
			/* Info Overlays, such as arrows and blocked indicator */

			/* camera + object */
			var objOffX float32 = camXPos + (float32(obj.Pos.X))
			var objOffY float32 = camYPos + (float32(obj.Pos.Y))

			/* camera zoom */
			objCamPosX := objOffX * world.ZoomScale
			objCamPosY := objOffY * world.ZoomScale

			/* Show objects with no fuel */
			if obj.Unique.TypeP.MachineSettings.MaxFuelKG > 0 && obj.Unique.KGFuel < obj.Unique.TypeP.MachineSettings.KgFuelPerCycle {

				img := WorldOverlays[gv.ObjOverlayNoFuel].Images.Main

				iSize := img.Bounds()
				var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
				op.GeoM.Scale(((float64(1))*float64(world.ZoomScale))/float64(iSize.Max.X),
					((float64(1))*float64(world.ZoomScale))/float64(iSize.Max.Y))
				op.GeoM.Translate(float64(objCamPosX), float64(objCamPosY))

				if img != nil {
					OpBatch[BatchTop] = op
					ImageBatch[BatchTop] = img
					if BatchTop < MaxBatch {
						BatchTop++
					} else {
						break
					}

				}
				/*
					} else if obj.Unique.TypeP.ShowBlocked && obj.Blocked {

						img := ObjOverlayTypes[gv.ObjOverlayBlocked].Image
						var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

						iSize := img.Bounds()
						op.GeoM.Translate(
							cBlockedIndicatorOffset,
							cBlockedIndicatorOffset)

						op.GeoM.Scale((1*float64(world.ZoomScale))/float64(iSize.Max.X),
							(1*float64(world.ZoomScale))/float64(iSize.Max.Y))
						op.GeoM.Translate(float64(objCamPosX), float64(objCamPosY))
						op.ColorScale.ScaleAlpha(0.5)

						if img != nil {
							OpBatch[BatchTop] = op
							ImageBatch[BatchTop] = img
							if BatchTop < MaxBatch {
								BatchTop++
							} else {
								break
							}
						}
				*/
			} else if obj.Unique.TypeP.ShowArrow {
				/* Output arrows */
				for _, port := range obj.Ports {
					if port.Type == gv.PORT_OUT && port.Dir == obj.Dir {

						img := WorldOverlays[port.Dir].Images.Main
						iSize := img.Bounds()
						var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
						op.GeoM.Scale((1*float64(world.ZoomScale))/float64(iSize.Max.X),
							((1)*float64(world.ZoomScale))/float64(iSize.Max.Y))
						op.GeoM.Translate(float64(objCamPosX), float64(objCamPosY))
						op.ColorScale.Scale(0.5, 0.5, 0.5, 0.7)

						/* Draw Arrow */
						if img != nil {
							OpBatch[BatchTop] = op
							ImageBatch[BatchTop] = img
							if BatchTop < MaxBatch {
								BatchTop++
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

	/* Batch render everything */
	for p := 0; p < BatchTop; p++ {
		screen.DrawImage(ImageBatch[p], OpBatch[p])
	}
}

/*
	func drawSubObjDebug(screen *ebiten.Image, b *world.BuildingData, bpos world.XY) {

		objOffX := camXPos + (float32(bpos.X))
		objOffY := camYPos + (float32(bpos.Y))

		x := float64(objOffX * world.ZoomScale)
		y := float64(objOffY * world.ZoomScale)

		if b.Obj.Unique.TypeP.Image == nil {
			return
		} else {
			var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

			iSize := b.Obj.Unique.TypeP.Image.Bounds()

			op.GeoM.Scale(
				(float64(1)*float64(world.ZoomScale))/float64(iSize.Max.X),
				(float64(1)*float64(world.ZoomScale))/float64(iSize.Max.Y))

			op.GeoM.Translate(math.Floor(x), math.Floor(y))
			op.ColorScale.Scale(0.125, 0.125, 0.5, 0.33)

			screen.DrawImage(b.Obj.Unique.TypeP.Image, op)

		}
	}
*/

func drawPixmapMode(screen *ebiten.Image) {
	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

	/* Single thread render terrain for WASM */
	if gv.WASMMode && frameCount%WASMTerrainDiv == 0 {
		ResourceRenderDaemonST()
		PixmapRenderST()
	}
	/* Draw superchunk images (pixmap mode)*/
	world.SuperChunkListLock.RLock()
	for _, sChunk := range VisSChunk {
		sChunk.PixelMapLock.Lock()
		if sChunk.PixelMap == nil {
			sChunk.PixelMapLock.Unlock()
			continue
		}

		op.GeoM.Reset()
		op.GeoM.Scale(
			(gv.MaxSuperChunk*float64(world.ZoomScale))/float64(gv.MaxSuperChunk),
			(gv.MaxSuperChunk*float64(world.ZoomScale))/float64(gv.MaxSuperChunk))

		op.GeoM.Translate(
			((float64(camXPos)+float64((sChunk.Pos.X))*gv.MaxSuperChunk)*float64(world.ZoomScale))-1,
			((float64(camYPos)+float64((sChunk.Pos.Y))*gv.MaxSuperChunk)*float64(world.ZoomScale))-1)

		screen.DrawImage(sChunk.PixelMap, op)
		sChunk.PixelMapLock.Unlock()
	}
	world.SuperChunkListLock.RUnlock()

	if world.ShowResourceLayer && gv.ResourceLegendImage != nil {
		op.GeoM.Reset()
		op.GeoM.Scale(2, 2)
		op.GeoM.Translate(8, float64(gv.ToolBarIconSize))
		screen.DrawImage(gv.ResourceLegendImage, op)
	}
}

func drawDebugInfo(screen *ebiten.Image) {

	/* Draw debug info */
	buf := fmt.Sprintf("FPS: %-4v UPS: %4.2f Objects: %8v, %-8v/%8v, %-8v Arch: %v Build: v%v-%v",
		int(world.FPSAvr.Value()),
		(world.UPSAvr.Value()),
		humanize.SIWithDigits(float64(world.TockCount), 2, ""),
		humanize.SIWithDigits(float64(world.ActiveTockCount), 2, ""),
		humanize.SIWithDigits(float64(world.TickCount), 2, ""),
		humanize.SIWithDigits(float64(world.ActiveTickCount), 2, ""),
		runtime.GOARCH, gv.Version, buildTime,
	)

	if gv.Debug {
		buf = buf + fmt.Sprintf(" (%v,%v)", MouseX, MouseY)
	}
	var pad float32 = 4
	DrawText(buf, world.MonoFont, color.White, world.ColorDebugBG,
		world.XYf32{X: 0, Y: float32(world.ScreenHeight) - pad},
		pad, screen, true, true, false)

	world.FPSAvr.Add(ebiten.ActualFPS())
}

func DrawText(input string, face font.Face, color color.Color, bgcolor color.Color, pos world.XYf32,
	pad float32, screen *ebiten.Image, justLeft bool, justUp bool, justCenter bool) world.XYf32 {
	var tmx, tmy float32
	halfPad := pad / 2

	tRect := text.BoundString(face, input)
	fHeight := text.BoundString(face, "1")

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
			tmy = float32(int(pos.Y))
		} else {
			tmy = float32(pos.Y + float32(tRect.Dy()))
		}
	}
	_, _, _, alpha := bgcolor.RGBA()

	if alpha > 0 {
		vector.DrawFilledRect(
			screen, tmx-halfPad, tmy-float32(fHeight.Dy())-halfPad,
			float32(tRect.Dx())+pad, float32(tRect.Dy())+pad, bgcolor, false,
		)
	}
	text.Draw(screen, input, face, int(tmx), int(tmy), color)

	return world.XYf32{X: float32(tRect.Dx()) + pad, Y: float32(tRect.Dy()) + pad}
}

/* Draw world objects */
func drawObject(screen *ebiten.Image, obj *world.ObjData, maskOnly bool) (op *ebiten.DrawImageOptions, img *ebiten.Image) {

	/* camera + object */
	objOffX := camXPos + (float32(obj.Pos.X))
	objOffY := camYPos + (float32(obj.Pos.Y))

	//Quick kludge for 1x3 object
	if obj.Unique.TypeP.Size.Y == 3 {
		if obj.Dir == 1 || obj.Dir == 3 {
			objOffX++
			objOffY--
		}
	}

	/* camera zoom */
	x := float64(objOffX * world.ZoomScale)
	y := float64(objOffY * world.ZoomScale)

	/* Draw sprite */
	if obj.Unique.TypeP.Images.Main == nil {
		return nil, nil
	} else {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

		iSize := obj.Unique.TypeP.Images.Main.Bounds()

		if obj.IsCorner {
			xx := float64(iSize.Size().X / 2)
			yy := float64(iSize.Size().Y / 2)
			op.GeoM.Translate(-xx, -yy)
			op.GeoM.Rotate(gv.NinetyDeg * float64(int(obj.CornerDir)))
			op.GeoM.Translate(xx, yy)
		} else if obj.Unique.TypeP.Rotatable {
			xx := float64(iSize.Size().X / 2)
			yy := float64(iSize.Size().Y / 2)
			op.GeoM.Translate(-xx, -yy)
			op.GeoM.Rotate(gv.NinetyDeg * float64(int(obj.Dir)))
			op.GeoM.Translate(xx, yy)
		}

		op.GeoM.Scale(
			(float64(obj.Unique.TypeP.Size.X)*float64(world.ZoomScale))/float64(iSize.Max.X),
			(float64(obj.Unique.TypeP.Size.Y)*float64(world.ZoomScale))/float64(iSize.Max.Y))

		op.GeoM.Translate(math.Floor(x), math.Floor(y))

		if obj.IsCorner {
			return op, obj.Unique.TypeP.Images.Corner
		} else if obj.Active && obj.Unique.TypeP.Images.Active != nil {
			return op, obj.Unique.TypeP.Images.Active
		} else if world.OverlayMode && obj.Unique.TypeP.Images.Overlay != nil {
			return op, obj.Unique.TypeP.Images.Overlay
		} else if maskOnly {
			return op, obj.Unique.TypeP.Images.Mask
		} else {
			return op, obj.Unique.TypeP.Images.Main
		}

	}
}

/* Update local vars with camera position calculations */
func calcScreenCamera() {
	var padding float32 = 3 /* Set to max item size */

	lastCamX = camXPos
	lastCamY = camYPos

	/* Adjust cam position for zoom */
	camXPos = float32(-world.CameraX) + ((float32(world.ScreenWidth) / 2.0) / world.ZoomScale)
	camYPos = float32(-world.CameraY) + ((float32(world.ScreenHeight) / 2.0) / world.ZoomScale)

	/* Get camera bounds */
	camStartX = uint16((1/world.ZoomScale + (world.CameraX - padding - (float32(world.ScreenWidth)/2.0)/world.ZoomScale)))
	camStartY = uint16((1/world.ZoomScale + (world.CameraY - padding - (float32(world.ScreenHeight)/2.0)/world.ZoomScale)))
	camEndX = uint16((float32(world.ScreenWidth)/world.ZoomScale + (world.CameraX + padding - (float32(world.ScreenWidth)/2.0)/world.ZoomScale)))
	camEndY = uint16((float32(world.ScreenHeight)/world.ZoomScale + (world.CameraY + padding - (float32(world.ScreenHeight)/2.0)/world.ZoomScale)))

	/* Pre-calc camera chunk position */
	screenStartX = camStartX / gv.ChunkSize
	screenStartY = camStartY / gv.ChunkSize
	screenEndX = camEndX / gv.ChunkSize
	screenEndY = camEndY / gv.ChunkSize
}

/* Draw materials on belts */
func drawMaterials(m *world.MatData, obj *world.ObjData, screen *ebiten.Image, scale float64, alpha float32, pos *world.XYf64) (op *ebiten.DrawImageOptions, img *ebiten.Image) {

	if m.Amount > 0 {
		img := m.TypeP.Image
		if img != nil {

			/* camera + object */
			objOffX := camXPos + (float32(obj.Pos.X))
			objOffY := camYPos + (float32(obj.Pos.Y))

			/* camera zoom */
			objCamPosX := float64(objOffX * world.ZoomScale)
			objCamPosY := float64(objOffY * world.ZoomScale)

			iSize := img.Bounds()
			var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
			xx := float64(iSize.Dx()) / 2.0
			yy := float64(iSize.Dy()) / 2.0
			op.GeoM.Translate(-xx, -yy)
			op.GeoM.Rotate(float64(m.Rot) * gv.NinetyDeg)
			if scale != 1 {
				op.GeoM.Scale(scale, scale)
			}
			op.GeoM.Translate(xx, yy)
			if pos != nil {
				op.GeoM.Translate(pos.X, pos.Y)
			}

			op.GeoM.Scale(
				((float64(1))*float64(world.ZoomScale))/float64(iSize.Max.X),
				((float64(1))*float64(world.ZoomScale))/float64(iSize.Max.Y))
			op.GeoM.Translate(objCamPosX, objCamPosY)
			op.ColorScale.ScaleAlpha(alpha)
			return op, img
		}
	}
	return nil, nil
}

func drawChatLines(screen *ebiten.Image) {

	var lineNum int
	util.ChatLinesLock.Lock()
	defer util.ChatLinesLock.Unlock()

	for x := util.ChatLinesTop; x > 0 && lineNum < gv.ChatHeightLines; x-- {
		line := util.ChatLines[x-1]
		/* Ignore old chat lines */
		since := time.Since(line.Timestamp)
		if !ConsoleActive && since > line.Lifetime {
			continue
		}
		lineNum++

		tBgColor := world.ColorToolTipBG
		r, g, b, _ := line.Color.RGBA()
		var blend float64 = 0
		if line.Lifetime-since < gv.ChatFadeTime {
			blend = (float64(gv.ChatFadeTime-(line.Lifetime-since)) / float64(gv.ChatFadeTime) * 100.0)
		}
		newAlpha := (254.0 - (blend * 2.55))

		oldAlpha := tBgColor.A
		faded := newAlpha - float64(253.0-int(oldAlpha))
		if faded <= 0 {
			faded = 0
		} else if faded > 254 {
			faded = 254
		}
		tBgColor.A = byte(faded)

		tRect := text.BoundString(world.ToolTipFont, line.Text)

		var pad int = int(world.FontDPI / 10.0)
		DrawText(line.Text, world.ToolTipFont,
			color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: byte(newAlpha)},
			tBgColor, world.XYf32{X: 0, Y: float32(world.ScreenHeight) - float32(lineNum*(tRect.Dy()+pad))},
			float32(pad), screen, true, true, false)
	}
}
