package main

import (
	"GameTest/gv"
	"GameTest/objects"
	"GameTest/util"
	"GameTest/world"
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
	cPreCache               = 4
	WASMTerrainDiv          = 5
	MaxBatch                = 100000
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

	lastResourceString string
	ConsoleActive      bool

	BatchTop   int
	ImageBatch [MaxBatch]*ebiten.Image
	OpBatch    [MaxBatch]*ebiten.DrawImageOptions
)

/* Setup a few images for later use */
func init() {

	world.MiniMapTile = ebiten.NewImage(1, 1)
	world.MiniMapTile.Fill(color.White)

	world.ToolBG = ebiten.NewImage(gv.ToolBarScale, gv.ToolBarScale)
	world.ToolBG.Fill(world.ColorCharcoalSemi)

	world.BeltBlock = ebiten.NewImage(1, 1)
	world.BeltBlock.Fill(world.ColorOrange)
}

/* Ebiten: Draw everything */
func (g *Game) Draw(screen *ebiten.Image) {

	/* Boot screen */
	if !world.MapGenerated.Load() ||
		!world.SpritesLoaded.Load() ||
		!world.PlayerReady.Load() {

		bootScreen(screen)
		drawDebugInfo(screen)
		drawChatLines(screen)
		time.Sleep(time.Millisecond)
		return
	}

	frameCount++

	/* If needed, calculate object visibility */
	updateVisData()

	/* WASM terrain rendering */
	if gv.WASMMode && frameCount%WASMTerrainDiv == 0 {
		objects.RenderTerrainST()
	}

	/* Draw modes */
	if world.ZoomScale > gv.MapPixelThreshold && !world.ShowResourceLayer { /* Draw icon mode */
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
	drawWorldTooltip(screen)

	/* Debug info */
	drawDebugInfo(screen)

	drawChatLines(screen)

}

func drawItemPlacement(screen *ebiten.Image) {
	/* Draw ghost for selected item */
	if SelectedItemType < gv.MaxItemType {
		mx, my := ebiten.CursorPosition()

		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		item := objects.GameObjTypes[SelectedItemType]

		/* Get mouse position on world */
		worldMouseX := int(float32(mx)/world.ZoomScale + (world.CameraX - (float32(world.ScreenWidth)/2.0)/world.ZoomScale))
		worldMouseY := int(float32(my)/world.ZoomScale + (world.CameraY - (float32(world.ScreenHeight)/2.0)/world.ZoomScale))

		/* camera + object */
		objOffX := camXPos + (float32((worldMouseX)))
		objOffY := camYPos + (float32((worldMouseY)))

		//Quick kludge for 1x3 object
		if item.Size.Y == 3 && (item.Direction == 1 || item.Direction == 3) {
			objOffX++
			objOffY--
		}

		/* camera zoom */
		x := float64(objOffX * world.ZoomScale)
		y := float64(objOffY * world.ZoomScale)

		iSize := item.Image.Bounds()
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
		wPos := world.XY{X: uint16(worldMouseX), Y: uint16(worldMouseY)}
		/* Check if object fits */
		if item.MultiTile {
			if !objects.SubObjFits(nil, item, false, wPos) {
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

		screen.DrawImage(item.Image, op)
	}
}

/* Look at camera position, make a list of visible superchunks and chunks. */
var VisObj []*world.ObjData
var VisChunk []*world.MapChunk

func updateVisData() {

	/* When needed, make a list of chunks to draw */
	if world.VisDataDirty.Load() {
		VisObj = []*world.ObjData{}
		VisChunk = []*world.MapChunk{}

		/* Calculate viewport */
		calcScreenCamera()

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

			sChunk.Visible = true

			for _, chunk := range sChunk.ChunkList {

				/* Is this chunk in the prerender area? */
				if chunk.Pos.X+cPreCache < screenStartX ||
					chunk.Pos.X-cPreCache > screenEndX ||
					chunk.Pos.Y+cPreCache < screenStartY ||
					chunk.Pos.Y-cPreCache > screenEndY {
					chunk.Precache = false
					continue
				}
				chunk.Precache = true

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
		if obj.TypeP.TypeI == gv.ObjTypeBasicBelt {

			/* Draw Input Materials */
			for _, port := range obj.Ports {
				if port.Buf.Amount > 0 {
					op, img = drawMaterials(port.Buf, obj, screen, 1.0, 1.0, nil)
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
		if obj.TypeP.TypeI == gv.ObjTypeBasicBeltOver {

			var start int32 = 32
			var middle int32 = 16
			var end int32 = 0

			found := false
			drewUnder := false

			/* Draw underpass */
			if obj.BeltOver.UnderIn != nil {

				if obj.BeltOver.UnderIn.Buf.Amount > 0 {
					/* Input */
					dir := objects.RotatePosF64(world.XYs{X: 0, Y: middle}, obj.Dir, world.XYf64{X: 16, Y: 48})
					op, img := drawMaterials(obj.BeltOver.UnderIn.Buf, obj, screen, 1.0, 1.0, &dir)
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
				dir := objects.RotatePosF64(world.XYs{X: 0, Y: middle}, obj.Dir, world.XYf64{X: 16, Y: 48})
				op, img := drawMaterials(obj.BeltOver.UnderOut.Buf, obj, screen, 1.0, 1.0, &dir)
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
				dir := objects.RotatePosF64(world.XYs{X: 0, Y: start}, obj.Dir, world.XYf64{X: 16, Y: 48})
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
				dir := objects.RotatePosF64(world.XYs{X: 0, Y: middle}, obj.Dir, world.XYf64{X: 16, Y: 48})
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
				dir := objects.RotatePosF64(world.XYs{X: 0, Y: end}, obj.Dir, world.XYf64{X: 16, Y: 48})
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
		if world.ShowInfoLayer {

			if obj.TypeP.TypeI == gv.ObjTypeBasicBox {
				for _, cont := range obj.Contents.Mats {
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

			oSize := objects.GetObjSize(obj, nil)

			/* Show objects with no fuel */
			if obj.TypeP.MaxFuelKG > 0 && obj.KGFuel < obj.TypeP.KgFuelEach {

				img := objects.ObjOverlayTypes[gv.ObjOverlayNoFuel].Image

				iSize := img.Bounds()
				var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
				op.GeoM.Scale(((float64(oSize.X))*float64(world.ZoomScale))/float64(iSize.Max.X),
					((float64(oSize.Y))*float64(world.ZoomScale))/float64(iSize.Max.Y))
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

			} else if obj.TypeP.ShowArrow {
				for _, port := range obj.Ports {
					if port.Type == gv.PORT_OUT && port.Dir == obj.Dir {

						img := objects.ObjOverlayTypes[port.Dir].Image
						iSize := img.Bounds()
						var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
						op.GeoM.Scale(((float64(oSize.X))*float64(world.ZoomScale))/float64(iSize.Max.X),
							((float64(oSize.Y))*float64(world.ZoomScale))/float64(iSize.Max.Y))
						op.GeoM.Translate(float64(objCamPosX), float64(objCamPosY))
						op.ColorScale.Scale(0.5, 0.5, 0.5, 0.66)

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
			/* Show blocked outputs */
			if obj.TypeP.ShowBlocked && obj.Blocked {

				img := objects.ObjOverlayTypes[gv.ObjOverlayBlocked].Image
				var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

				iSize := img.Bounds()
				op.GeoM.Translate(
					cBlockedIndicatorOffset,
					cBlockedIndicatorOffset)
				op.GeoM.Scale(((float64(oSize.X))*float64(world.ZoomScale))/float64(iSize.Max.X),
					((float64(oSize.Y))*float64(world.ZoomScale))/float64(iSize.Max.Y))
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
			}

		}
	}

	/* Batch render everything */
	for p := 0; p < BatchTop; p++ {
		screen.DrawImage(ImageBatch[p], OpBatch[p])
	}
}

func drawSubObjDebug(screen *ebiten.Image, b *world.BuildingData, bpos world.XY) {

	/* camera + object */
	objOffX := camXPos + (float32(bpos.X))
	objOffY := camYPos + (float32(bpos.Y))

	/* camera zoom */
	x := float64(objOffX * world.ZoomScale)
	y := float64(objOffY * world.ZoomScale)

	/* Draw sprite */
	if b.Obj.TypeP.Image == nil {
		return
	} else {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

		iSize := b.Obj.TypeP.Image.Bounds()

		op.GeoM.Scale(
			(float64(1)*float64(world.ZoomScale))/float64(iSize.Max.X),
			(float64(1)*float64(world.ZoomScale))/float64(iSize.Max.Y))

		op.GeoM.Translate(math.Floor(x), math.Floor(y))
		op.ColorScale.Scale(0.125, 0.125, 0.5, 0.33)

		screen.DrawImage(b.Obj.TypeP.Image, op)

	}
}
func drawPixmapMode(screen *ebiten.Image) {
	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

	/* Single thread render terrain for WASM */
	if gv.WASMMode && frameCount%WASMTerrainDiv == 0 {
		objects.ResourceRenderDaemonST()
		objects.PixmapRenderST()
	}
	/* Draw superchunk images (pixmap mode)*/
	world.SuperChunkListLock.RLock()
	for _, sChunk := range world.SuperChunkList {
		if !sChunk.Visible {
			continue
		}

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
		op.GeoM.Translate(8, gv.ToolBarScale)
		screen.DrawImage(gv.ResourceLegendImage, op)
	}
}

func drawDebugInfo(screen *ebiten.Image) {
	world.FPSAvr.Add(ebiten.ActualFPS())

	/* Draw debug info */
	buf := fmt.Sprintf("FPS: %.2f UPS: %0.2f Tocks: %v Ticks %v Draws: %v Arch: %v Build: v%v-%v",
		world.FPSAvr.Value(),
		1000000000.0/float64(world.MeasuredObjectUPS_ns)/2,
		humanize.SIWithDigits(float64(world.TockCount), 2, ""),
		humanize.SIWithDigits(float64(world.TickCount), 2, ""),
		humanize.SIWithDigits(float64(BatchTop), 2, ""),
		runtime.GOARCH, gv.Version, buildTime)

	DrawText(buf, world.MonoFont, color.White, world.ColorDebugBG, world.XY{X: 0, Y: world.ScreenHeight}, 11, screen, true, false, false)
}

func drawWorldTooltip(screen *ebiten.Image) {

	mx, my := ebiten.CursorPosition()
	fmx := float32(mx)
	fmy := float32(my)

	/* Get mouse position on world */
	worldMouseX := (fmx/world.ZoomScale + (world.CameraX - (float32(world.ScreenWidth)/2.0)/world.ZoomScale))
	worldMouseY := (fmy/world.ZoomScale + (world.CameraY - (float32(world.ScreenHeight)/2.0)/world.ZoomScale))

	/* Toolbar tool tip */
	uipix := float32((ToolbarMax * int(gv.ToolBarScale+gv.ToolBarScale)))

	val := int(fmx / (gv.ToolBarScale + gv.ToolBarSpacing))
	if fmx <= uipix && fmy <= gv.ToolBarScale &&
		val >= 0 && val < ToolbarMax {
		if fmx != world.PrevMouseX && fmy != world.PrevMouseY {

			pos := int(fmx / float32(gv.ToolBarScale+gv.ToolBarSpacing))
			item := ToolbarItems[pos]
			var toolTip string

			if item.OType.Info != "" {
				keyName := ""
				if item.OType.QKey != 0 {
					keyName = " ( " + item.OType.QKey.String() + " key )"
				}
				toolTip = fmt.Sprintf("%v\n%v\n%v", item.OType.Name, item.OType.Info, keyName)
			} else {
				toolTip = fmt.Sprintf("%v\n", item.OType.Name)
			}
			DrawText(toolTip, world.ToolTipFont, world.ColorWhite, world.ColorToolTipBG,
				world.XY{X: uint16(fmx) + 20, Y: uint16(fmy) + 40}, 11, screen,
				true, false, false)

			DrawToolbar(false, true, pos)
		}
	} else {
		/* Erase hover highlight when mouse moves off */
		if ToolbarHover {
			DrawToolbar(false, false, 0)
			ToolbarHover = false
		}

		/* World Obj tool tip */
		pos := util.FloatXYToPosition(worldMouseX, worldMouseY)
		chunk := util.GetChunk(pos)

		toolTip := ""
		if gv.Debug {
			toolTip = "(Debug Mode):\n"
		}
		found := false

		if chunk != nil {
			b := util.GetObj(pos, chunk)
			if b != nil {
				o := b.Obj
				found = true
				toolTip = fmt.Sprintf("%v: %v\n",
					o.TypeP.Name,
					util.PosToString(world.XY{X: uint16(worldMouseX), Y: uint16(worldMouseY)}))
				if o.Contents != nil {
					for z := 0; z < gv.MAT_MAX; z++ {
						if o.Contents.Mats[z] != nil {
							toolTip = toolTip + fmt.Sprintf("Contents: %v: %0.2f%v\n",
								o.Contents.Mats[z].TypeP.Name, o.Contents.Mats[z].Amount, o.Contents.Mats[z].TypeP.UnitName)
						}
					}
				}
				if o.TypeP.MaxFuelKG > 0 {
					toolTip = toolTip + fmt.Sprintf("Max Fuel: %0.2f kg\n", o.TypeP.MaxFuelKG)
					if o.KGFuel > o.TypeP.KgFuelEach {
						toolTip = toolTip + fmt.Sprintf("Fuel: %0.2f kg\n", o.KGFuel)
					} else {
						toolTip = toolTip + "NO FUEL\n"
					}
				}

				if o.SingleContent != nil && o.SingleContent.Amount > 0 {
					toolTip = toolTip + fmt.Sprintf("Contains: %0.2f%v %v\n", o.SingleContent.Amount, o.SingleContent.TypeP.UnitName, o.SingleContent.TypeP.Name)
				}

				if o.Blocked {
					toolTip = toolTip + "BLOCKED\n"
				}
				if o.MinerData != nil && o.MinerData.ResourcesCount == 0 {
					toolTip = toolTip + "NOTHING TO MINE.\n"
				}

				if gv.Debug {
					if o.TypeP.KgFuelEach > 0 {
						toolTip = toolTip + fmt.Sprintf("Fuel per tock: %0.2f kg\n", o.TypeP.KgFuelEach)
					}
					if o.TypeP.KgMineEach > 0 {
						toolTip = toolTip + fmt.Sprintf("Mine per tock: %0.2f kg\n", o.TypeP.KgMineEach)
					}
					if o.TypeP.MaxContainKG > 0 {
						toolTip = toolTip + fmt.Sprintf("Max contents: %0.2f kg\n", o.TypeP.MaxContainKG)
					}

					for _, p := range o.Ports {
						if p.Obj == nil {
							continue
						}
						var tstring = "None"
						var tunit = "?"
						if p.Buf.TypeP != nil {
							tstring = p.Buf.TypeP.Name
							tunit = p.Buf.TypeP.UnitName
						}

						if p.Type == gv.PORT_IN {
							toolTip = toolTip + fmt.Sprintf("Input: %v: %v: %v: %0.2f %v\n",
								util.DirToName(uint8(p.Dir)),
								p.Obj.TypeP.Name,
								tstring,
								p.Buf.Amount,
								tunit)
						} else if p.Type == gv.PORT_OUT {
							toolTip = toolTip + fmt.Sprintf("Output: %v: %v: %v: %0.2f %v\n",
								util.DirToName(uint8(p.Dir)),
								p.Obj.TypeP.Name,
								tstring,
								p.Buf.Amount,
								tunit)
						} else if p.Type == gv.PORT_FOUT {
							toolTip = toolTip + fmt.Sprintf("FuelOut: %v: %v: %v: %0.2f %v\n",
								util.DirToName(uint8(p.Dir)),
								p.Obj.TypeP.Name,
								tstring,
								p.Buf.Amount,
								tunit)
						} else if p.Type == gv.PORT_FIN {
							toolTip = toolTip + fmt.Sprintf("FuelIn: %v: %v: %v: %0.2f %v\n",
								util.DirToName(uint8(p.Dir)),
								p.Obj.TypeP.Name,
								tstring,
								p.Buf.Amount,
								tunit)
						}
					}
					if o.TypeP != nil {
						if o.TypeP.MaxContainKG > 0 {
							toolTip = toolTip + fmt.Sprintf("MaxContainKG: %v\n", o.TypeP.MaxContainKG)
						}
						if o.TypeP.KW > 0 {
							toolTip = toolTip + fmt.Sprintf("KW: %v\n", o.TypeP.KW)
						}
						if o.HasTock {
							toolTip = toolTip + "Tocking\n"
						}
						if o.HasTick {
							toolTip = toolTip + "Ticking\n"
						}
					}

				}

				if o.TypeP.Info != "" {
					toolTip = toolTip + o.TypeP.Info + "\n"
				}
			}
		}

		/* No object contents found, just show x/y */
		if !found {
			toolTip = fmt.Sprintf("(%v, %v)",
				humanize.Comma(int64((worldMouseX - gv.XYCenter))),
				humanize.Comma(int64((worldMouseY - gv.XYCenter))))
		}

		/* Tooltip for resources */
		if world.ShowResourceLayer {
			buf := ""
			if fmx != world.PrevMouseX && fmy != world.PrevMouseY {
				for p := 1; p < len(objects.NoiseLayers); p++ {
					var h float32 = float32(math.Abs(float64(objects.NoiseMap(worldMouseX, worldMouseY, p))))

					if h > 0 {
						buf = buf + fmt.Sprintf("%v: %0.2f%%\n", objects.NoiseLayers[p].Name, util.Min(h*100.0, 100.0))
					}
				}
			} else {
				/* save a bit of processing */
				buf = lastResourceString
			}
			if buf != "" {
				lastResourceString = buf
				DrawText("Yields:\n"+buf, world.ToolTipFont, world.ColorAqua, world.ColorToolTipBG,
					world.XY{X: uint16(fmx + 20), Y: uint16(fmy + 20)}, 11, screen, true, false, false)
			}
		}
		DrawText(toolTip, world.ToolTipFont, color.White, world.ColorToolTipBG, world.XY{X: world.ScreenWidth, Y: world.ScreenHeight}, 11, screen, false, true, false)

	}
}

func DrawText(input string, face font.Face, color color.Color, bgcolor color.Color, pos world.XY,
	pad float32, screen *ebiten.Image, justLeft bool, justUp bool, justCenter bool) {
	var mx, my float32

	halfPad := (pad / 2)
	tRect := text.BoundString(face, input)
	if justCenter {
		mx = float32(int(pos.X) - (tRect.Dx() / 2))
		my = float32(int(pos.Y) - (tRect.Dy() / 2))
	} else {
		if justLeft {
			mx = float32(pos.X) + halfPad
		} else {
			mx = float32(int(pos.X)-tRect.Dx()) - halfPad
		}

		if justUp {
			my = float32(int(pos.Y)-tRect.Dy()) + halfPad
		} else {
			my = float32(pos.Y) - halfPad
		}
	}
	_, _, _, alpha := bgcolor.RGBA()
	if alpha > 0 {
		vector.DrawFilledRect(screen, mx-halfPad, my-11-halfPad, float32(tRect.Dx())+pad, float32(tRect.Dy())+pad, bgcolor)
	}
	text.Draw(screen, input, face, int(mx), int(my), color)
}

/* Draw world objects */
func drawObject(screen *ebiten.Image, obj *world.ObjData, maskOnly bool) (op *ebiten.DrawImageOptions, img *ebiten.Image) {

	/* camera + object */
	objOffX := camXPos + (float32(obj.Pos.X))
	objOffY := camYPos + (float32(obj.Pos.Y))

	//Quick kludge for 1x3 object
	if obj.TypeP.Size.Y == 3 {
		if obj.Dir == 1 || obj.Dir == 3 {
			objOffX++
			objOffY--
		}
	}

	/* camera zoom */
	x := float64(objOffX * world.ZoomScale)
	y := float64(objOffY * world.ZoomScale)

	/* Draw sprite */
	if obj.TypeP.Image == nil {
		return nil, nil
	} else {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

		iSize := obj.TypeP.Image.Bounds()

		if obj.TypeP.Rotatable {
			xx := float64(iSize.Size().X / 2)
			yy := float64(iSize.Size().Y / 2)
			op.GeoM.Translate(-xx, -yy)
			op.GeoM.Rotate(gv.NinetyDeg * float64(int(obj.Dir)))
			op.GeoM.Translate(xx, yy)
		}

		op.GeoM.Scale(
			(float64(obj.TypeP.Size.X)*float64(world.ZoomScale))/float64(iSize.Max.X),
			(float64(obj.TypeP.Size.Y)*float64(world.ZoomScale))/float64(iSize.Max.Y))

		op.GeoM.Translate(math.Floor(x), math.Floor(y))

		if obj.TypeP.ImagePathActive != "" && obj.Active {
			return op, obj.TypeP.ImageActive
		} else if maskOnly {
			return op, obj.TypeP.ImageMask
		} else {
			return op, obj.TypeP.Image
		}

	}
}

/* Update local vars with camera position calculations */
func calcScreenCamera() {
	/* Adjust cam position for zoom */
	camXPos = float32(-world.CameraX) + ((float32(world.ScreenWidth) / 2.0) / world.ZoomScale)
	camYPos = float32(-world.CameraY) + ((float32(world.ScreenHeight) / 2.0) / world.ZoomScale)

	/* Get camera bounds */
	camStartX = uint16((1/world.ZoomScale + (world.CameraX - (float32(world.ScreenWidth)/2.0)/world.ZoomScale)))
	camStartY = uint16((1/world.ZoomScale + (world.CameraY - (float32(world.ScreenHeight)/2.0)/world.ZoomScale)))
	camEndX = uint16((float32(world.ScreenWidth)/world.ZoomScale + (world.CameraX - (float32(world.ScreenWidth)/2.0)/world.ZoomScale)))
	camEndY = uint16((float32(world.ScreenHeight)/world.ZoomScale + (world.CameraY - (float32(world.ScreenHeight)/2.0)/world.ZoomScale)))

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

	var lineNum uint16
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
		DrawText(line.Text, world.ToolTipFont, color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: byte(newAlpha)}, tBgColor, world.XY{X: 0, Y: (world.ScreenHeight - 25) - (lineNum * 26)}, 11, screen, true, false, false)
	}
}
