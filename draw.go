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
	cMAX_RENDER_NS          = 1000000000 / 360 /* 360 FPS */
	cMAX_RENDER_NS_BOOT     = 1000000000 / 30  /* 30 FPS */
	cPreCache               = 4
	WASMTerrainDiv          = 5
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
	ImageBatch [gv.ChunkSizeTotal * 3]*ebiten.Image
	OpBatch    [gv.ChunkSizeTotal * 3]*ebiten.DrawImageOptions
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

	/* Calculate viewport */
	calcScreenCamera()

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

	/* Draw toolbar */
	toolbarCacheLock.RLock()
	screen.DrawImage(toolbarCache, nil)
	toolbarCacheLock.RUnlock()

	if SelectedItemType > 0 && SelectedItemType < 255 {
		op := &ebiten.DrawImageOptions{}
		item := objects.GameObjTypes[SelectedItemType]
		img := item.Image

		op.GeoM.Scale(float64(world.ZoomScale)/gv.SpriteScale, float64(world.ZoomScale)/gv.SpriteScale)
		op.GeoM.Translate(float64(gMouseX), float64(gMouseY))
		op.ColorScale.Scale(0.5, 0.5, 0.5, 0.4)
		screen.DrawImage(img, op)
	}

	/* Tooltips */
	drawWorldTooltip(screen)

	/* Debug info */
	drawDebugInfo(screen)

	drawChatLines(screen)

}

/* Look at camera position, make a list of visible superchunks and chunks. Saves to VisChunks, checks world.CameraDirty */
func updateVisData() {

	/* When needed, make a list of chunks to draw */
	if world.VisDataDirty.Load() {

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
				chunk.Visible = true

				world.VisDataDirty.Store(false)

			}
		}
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

	world.SuperChunkListLock.RLock()
	for _, sChunk := range world.SuperChunkList {
		for _, chunk := range sChunk.ChunkList {
			if !chunk.Visible {
				continue
			}

			BatchTop = 0

			/* Draw ground*/
			op, img := drawTerrain(chunk)
			if img != nil {
				OpBatch[BatchTop] = op
				ImageBatch[BatchTop] = img
				BatchTop++
			}

			/* Draw objects in chunk */
			for _, obj := range chunk.ObjList {
				if obj == nil {
					continue
				}
				/* Is this object on the screen? */
				if obj.Pos.X < camStartX || obj.Pos.X > camEndX || obj.Pos.Y < camStartY || obj.Pos.Y > camEndY {
					continue
				}

				/* Time to draw it */
				op, img = drawObject(screen, obj)
				if img != nil {
					OpBatch[BatchTop] = op
					ImageBatch[BatchTop] = img
					BatchTop++
				}

				/* Overlays */
				/* Draw belt overlays */
				if obj.TypeP.TypeI == gv.ObjTypeBasicBelt {

					/* Draw Input Materials */
					for _, port := range obj.Inputs {
						if port.Buf.Amount > 0 {
							op, img = drawMaterials(port.Buf, obj, screen, 1.0)
							if img != nil {
								OpBatch[BatchTop] = op
								ImageBatch[BatchTop] = img
								BatchTop++
							}
							break
						}
					}
				}
				if obj.TypeP.TypeI == gv.ObjTypeBasicBeltInterRight {
					op, img = drawMaterials(obj.Ports[2].Buf, obj, screen, 0.5)
					if img != nil {
						OpBatch[BatchTop] = op
						ImageBatch[BatchTop] = img
						BatchTop++
					}
				}
				if world.ShowInfoLayer {

					if obj.TypeP.TypeI == gv.ObjTypeBasicBox {
						for _, cont := range obj.Contents {
							if cont == nil {
								continue
							}
							op, img = drawMaterials(cont, obj, screen, 1.0)
							if img != nil {
								OpBatch[BatchTop] = op
								ImageBatch[BatchTop] = img
								BatchTop++
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
					if obj.TypeP.MaxFuelKG > 0 && obj.KGFuel < obj.TypeP.KgFuelEach {

						img := objects.ObjOverlayTypes[gv.ObjOverlayNoFuel].Image

						iSize := img.Bounds()
						var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
						op.GeoM.Scale(((float64(obj.TypeP.Size.X))*float64(world.ZoomScale))/float64(iSize.Max.X),
							((float64(obj.TypeP.Size.Y))*float64(world.ZoomScale))/float64(iSize.Max.Y))
						op.GeoM.Translate(float64(objCamPosX), float64(objCamPosY))

						if img != nil {
							OpBatch[BatchTop] = op
							ImageBatch[BatchTop] = img
							BatchTop++
						}

					} else if obj.TypeP.ShowArrow {
						for p := range obj.Ports {
							img := objects.ObjOverlayTypes[p].Image
							iSize := img.Bounds()
							var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
							op.GeoM.Scale(((float64(obj.TypeP.Size.X))*float64(world.ZoomScale))/float64(iSize.Max.X),
								((float64(obj.TypeP.Size.Y))*float64(world.ZoomScale))/float64(iSize.Max.Y))
							op.GeoM.Translate(float64(objCamPosX), float64(objCamPosY))

							/* Draw Arrow */
							if img != nil {
								OpBatch[BatchTop] = op
								ImageBatch[BatchTop] = img
								BatchTop++
							}
						}
					}
					/* Show blocked outputs */
					if (obj.TypeP.ShowBlocked && obj.Blocked) ||
						obj.MinerData != nil && obj.MinerData.ResourcesCount == 0 {

						img := objects.ObjOverlayTypes[gv.ObjOverlayBlocked].Image
						var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

						iSize := img.Bounds()
						op.GeoM.Translate(
							cBlockedIndicatorOffset,
							cBlockedIndicatorOffset)
						op.GeoM.Scale(((float64(obj.TypeP.Size.X))*float64(world.ZoomScale))/float64(iSize.Max.X),
							((float64(obj.TypeP.Size.Y))*float64(world.ZoomScale))/float64(iSize.Max.Y))
						op.GeoM.Translate(float64(objCamPosX), float64(objCamPosY))

						if img != nil {
							OpBatch[BatchTop] = op
							ImageBatch[BatchTop] = img
							BatchTop++
						}
					}

				}
			}

			for p := 0; p < BatchTop; p++ {
				screen.DrawImage(ImageBatch[p], OpBatch[p])
			}
		}
	}
	world.SuperChunkListLock.RUnlock()
}

func drawPixmapMode(screen *ebiten.Image) {
	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

	/* Single thread render terrain for WASM */
	if gv.WASMMode {
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
	/* Draw debug info */
	buf := fmt.Sprintf("FPS: %.2f UPS: %.2f Active Objects: %v Arch: %v Build: %v",
		ebiten.ActualFPS(),
		1000000000.0/float32(world.MeasuredObjectUPS_ns),
		humanize.SIWithDigits(float64(world.TockCount), 2, ""),
		runtime.GOARCH, buildTime)

	DrawText(buf, world.ToolTipFont, color.White, world.ColorDebugBG, world.XY{X: 0, Y: world.ScreenHeight}, 11, screen, true, false, false)
}

func drawWorldTooltip(screen *ebiten.Image) {
	/* Get mouse position on world */
	worldMouseX := (world.MouseX/world.ZoomScale + (world.CameraX - (float32(world.ScreenWidth)/2.0)/world.ZoomScale))
	worldMouseY := (world.MouseY/world.ZoomScale + (world.CameraY - (float32(world.ScreenHeight)/2.0)/world.ZoomScale))

	/* Toolbar tool tip */
	uipix := float32((ToolbarMax * int(gv.ToolBarScale+gv.ToolBarScale)))

	val := int(world.MouseX / (gv.ToolBarScale + gv.ToolBarSpacing))
	if world.MouseX <= uipix && world.MouseY <= gv.ToolBarScale &&
		val >= 0 && val < ToolbarMax {
		if gMouseX != gPrevMouseX && gMouseY != gPrevMouseY {

			pos := int(world.MouseX / float32(gv.ToolBarScale+gv.ToolBarSpacing))
			item := ToolbarItems[pos]
			var toolTip string

			if item.OType.Info != "" {
				toolTip = fmt.Sprintf("%v\n%v\n", item.OType.Name, item.OType.Info)
			} else {
				toolTip = fmt.Sprintf("%v\n", item.OType.Name)
			}
			DrawText(toolTip, world.ToolTipFont, world.ColorWhite, world.ColorToolTipBG,
				world.XY{X: uint16(world.MouseX) + 20, Y: uint16(world.MouseY) + 40}, 11, screen,
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
		found := false

		if chunk != nil {
			b := util.GetObj(pos, chunk)
			if b != nil {
				o := b.Obj
				found = true
				toolTip = fmt.Sprintf("%v: (%v,%v)\n",
					o.TypeP.Name,
					humanize.Comma(int64((worldMouseX - gv.XYCenter))),
					humanize.Comma(int64((worldMouseY - gv.XYCenter))))
				for z := 0; z < gv.MAT_MAX; z++ {
					if o.Contents[z] != nil {
						toolTip = toolTip + fmt.Sprintf("Contents: %v: %0.2f%v\n",
							o.Contents[z].TypeP.Name, o.Contents[z].Amount, o.Contents[z].TypeP.UnitName)
					}
				}
				if o.TypeP.MaxFuelKG > 0 {
					if o.KGFuel > o.TypeP.KgFuelEach {
						toolTip = toolTip + fmt.Sprintf("Fuel: %0.2f kg\n", o.KGFuel)
					} else {
						toolTip = toolTip + "NO FUEL\n"
					}
				}

				if o.Blocked {
					toolTip = toolTip + "BLOCKED\n"
				}
				if o.MinerData != nil && o.MinerData.ResourcesCount == 0 {
					toolTip = toolTip + "NOTHING TO MINE.\n"
				}

				if gv.Debug {

					for z, p := range o.Ports {
						if p.Obj == nil {
							continue
						}
						if p.Type == gv.PORT_IN && p.Obj != nil {
							toolTip = toolTip + fmt.Sprintf("(Input: %v: %v: %0.2f)\n",
								util.DirToName(uint8(z)),
								p.Obj.TypeP.Name,
								//p.Buf.TypeP.Name,
								p.Buf.Amount)
						}
						if p.Type == gv.PORT_OUT && p.Obj != nil {
							toolTip = toolTip + fmt.Sprintf("(Output: %v: %v: %0.2f)\n",
								util.DirToName(uint8(z)),
								p.Obj.TypeP.Name,
								//p.Buf.TypeP.Name,
								p.Buf.Amount)
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
			if gMouseX != gPrevMouseX && gMouseY != gPrevMouseY {
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
					world.XY{X: uint16(gMouseX + 20), Y: uint16(gMouseY + 20)}, 11, screen, true, false, false)
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
func drawObject(screen *ebiten.Image, obj *world.ObjData) (op *ebiten.DrawImageOptions, img *ebiten.Image) {

	/* camera + object */
	objOffX := camXPos + (float32(obj.Pos.X))
	objOffY := camYPos + (float32(obj.Pos.Y))

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
func drawMaterials(m *world.MatData, obj *world.ObjData, screen *ebiten.Image, scale float64) (op *ebiten.DrawImageOptions, img *ebiten.Image) {

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
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(xx, yy)

			op.GeoM.Scale(
				((float64(obj.TypeP.Size.X))*float64(world.ZoomScale))/float64(iSize.Max.X),
				((float64(obj.TypeP.Size.Y))*float64(world.ZoomScale))/float64(iSize.Max.Y))
			op.GeoM.Translate(objCamPosX, objCamPosY)
			return op, img
		}
	}
	return nil, nil
}
