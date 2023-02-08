package main

import (
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/gv"
	"GameTest/objects"
	"GameTest/util"
	"fmt"
	"image/color"
	"math"
	"runtime"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

const (
	cNinetyDeg              = math.Pi / 2
	cBlockedIndicatorOffset = 0
	cMAX_RENDER_NS          = 1000000000 / 360 /* 360 FPS */
	cMAX_RENDER_NS_BOOT     = 1000000000 / 30  /* 30 FPS */
	cPreCache               = 4
	WASMTerrtainDiv         = 5
)

var (
	camXPos float64
	camYPos float64

	camStartX int
	camStartY int
	camEndX   int
	camEndY   int

	screenStartX int
	screenStartY int
	screenEndX   int
	screenEndY   int
	frameCount   uint64
)

/* Setup a few images for later use */
func init() {
	glob.MiniMapTile = ebiten.NewImage(1, 1)
	glob.MiniMapTile.Fill(color.White)

	glob.ToolBG = ebiten.NewImage(gv.ToolBarScale, gv.ToolBarScale)
	glob.ToolBG.Fill(glob.ColorCharcolSemi)
}

/* Ebiten: Draw everything */
func (g *Game) Draw(screen *ebiten.Image) {

	if !glob.MapGenerated.Load() ||
		!glob.SpritesLoaded.Load() ||
		!glob.PlayerReady.Load() {

		bootScreen(screen)
		time.Sleep(time.Microsecond)
		return
	}

	frameCount++

	calcScreenCamera()
	updateVisData()

	if gv.WASMMode && frameCount%WASMTerrtainDiv == 0 {
		objects.RenderTerrainST()
	}
	if glob.ZoomScale > gv.MapPixelThreshold { /* Draw icon mode */
		drawIconMode(screen)
	} else {
		drawPixmapMode(screen)
	}

	/* Draw toolbar */
	screen.DrawImage(toolbarCache, nil)

	drawWorldTooltip(screen)
	drawDebugInfo(screen)

}

/* Look at camera position, make a list of visible superchunks and chunks. Saves to VisChunks, checks glob.CameraDirty */
func updateVisData() {

	/* When needed, make a list of chunks to draw */
	if glob.VisDataDirty.Load() {

		glob.SuperChunkListLock.RLock()
		for _, sChunk := range glob.SuperChunkList {

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

				glob.VisDataDirty.Store(false)

			}
		}
		glob.SuperChunkListLock.RUnlock()
	}
}

func drawTerrain(chunk *glob.MapChunk, screen *ebiten.Image) {
	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

	/* Draw ground */
	chunk.TerrainLock.RLock()
	cTmp := chunk.TerrainImg
	if chunk.TerrainImg == nil {
		cTmp = glob.TempChunkImage
	}

	iSize := cTmp.Bounds().Size()
	op.GeoM.Reset()
	op.GeoM.Scale((gv.ChunkSize*glob.ZoomScale)/float64(iSize.X),
		(gv.ChunkSize*glob.ZoomScale)/float64(iSize.Y))
	op.GeoM.Translate((camXPos+float64(chunk.Pos.X*gv.ChunkSize))*glob.ZoomScale,
		(camYPos+float64(chunk.Pos.Y*gv.ChunkSize))*glob.ZoomScale)
	screen.DrawImage(cTmp, op)
	chunk.TerrainLock.RUnlock()
}

func drawIconMode(screen *ebiten.Image) {
	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

	glob.SuperChunkListLock.RLock()
	for _, sChunk := range glob.SuperChunkList {
		for _, chunk := range sChunk.ChunkList {
			if !chunk.Visible {
				continue
			}

			drawTerrain(chunk, screen)
			if gv.ShowMineralLayer {
				continue
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
				drawObject(screen, obj)

				/* Overlays */
				/* Draw belt overlays */
				if obj.TypeP.TypeI == gv.ObjTypeBasicBelt {

					/* camera + object */
					objOffX := camXPos + (float64(obj.Pos.X))
					objOffY := camYPos + (float64(obj.Pos.Y))

					/* camera zoom */
					objCamPosX := objOffX * glob.ZoomScale
					objCamPosY := objOffY * glob.ZoomScale

					iSize := obj.TypeP.Image.Bounds()
					op.GeoM.Reset()
					op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X),
						((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
					op.GeoM.Translate(objCamPosX, objCamPosY)

					/* Draw Input Materials */
					for p := range obj.Ports {
						drawMaterials(&obj.Ports[p].Buf, obj, op, screen)
					}

				}
				if glob.ShowInfoLayer {
					/* Info Overlays, such as arrows and blocked indicator */

					/* camera + object */
					objOffX := camXPos + (float64(obj.Pos.X))
					objOffY := camYPos + (float64(obj.Pos.Y))

					/* camera zoom */
					objCamPosX := objOffX * glob.ZoomScale
					objCamPosY := objOffY * glob.ZoomScale

					iSize := obj.TypeP.Image.Bounds()
					op.GeoM.Reset()
					op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X),
						((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
					op.GeoM.Translate(objCamPosX, objCamPosY)

					if obj.TypeP.ShowArrow {
						for p, port := range obj.Ports {
							if port.PortDir != gv.PORT_OUTPUT {
								continue
							}

							iSize := obj.TypeP.Image.Bounds()
							op.GeoM.Reset()
							op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X),
								((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
							op.GeoM.Translate(objCamPosX, objCamPosY)

							/* Draw Arrow */
							img := objects.ObjOverlayTypes[p].Image
							if img != nil {
								screen.DrawImage(img, op)
							}

							/* Show blocked outputs */
							img = objects.ObjOverlayTypes[gv.ObjOverlayBlocked].Image
							if obj.TypeP.ShowBlocked && obj.Blocked {

								var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

								iSize := obj.TypeP.Image.Bounds()
								op.GeoM.Reset()
								op.GeoM.Translate(
									float64(iSize.Max.X)-float64(objects.ObjOverlayTypes[gv.ObjOverlayBlocked].Image.Bounds().Max.X)-cBlockedIndicatorOffset,
									cBlockedIndicatorOffset)
								op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X),
									((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
								op.GeoM.Translate(objCamPosX, objCamPosY)
								screen.DrawImage(img, op)
							}
						}
					}
				}
			}
		}
	}
	glob.SuperChunkListLock.RUnlock()
}

func drawPixmapMode(screen *ebiten.Image) {
	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

	/* Single thread render terrain for WASM */
	if gv.WASMMode {
		objects.PixmapRenderST()
	}
	/* Draw superchunk images (pixmap mode)*/
	glob.SuperChunkListLock.RLock()
	for _, sChunk := range glob.SuperChunkList {
		if !sChunk.Visible {
			continue
		}

		sChunk.PixLock.Lock()
		if sChunk.PixMap == nil {
			sChunk.PixLock.Unlock()
			continue
		}

		op.GeoM.Reset()
		op.GeoM.Scale(
			(gv.MaxSuperChunk*glob.ZoomScale)/float64(gv.MaxSuperChunk),
			(gv.MaxSuperChunk*glob.ZoomScale)/float64(gv.MaxSuperChunk))

		op.GeoM.Translate(
			((camXPos+float64((sChunk.Pos.X))*gv.MaxSuperChunk)*glob.ZoomScale)-1,
			((camYPos+float64((sChunk.Pos.Y))*gv.MaxSuperChunk)*glob.ZoomScale)-1)

		screen.DrawImage(sChunk.PixMap, op)
		sChunk.PixLock.Unlock()
	}
	glob.SuperChunkListLock.RUnlock()
}

func drawDebugInfo(screen *ebiten.Image) {
	/* Draw debug info */
	dbuf := fmt.Sprintf("FPS: %.2f UPS: %.2f Active Objects: %v Arch: %v Build: %v",
		ebiten.ActualFPS(),
		1000000000.0/float64(glob.MeasuredObjectUPS_ns),
		humanize.SIWithDigits(float64(objects.TockWorkSize*(objects.NumWorkers*objects.WorkChunks)), 2, ""),
		runtime.GOARCH, buildTime)

	tRect := text.BoundString(glob.ToolTipFont, dbuf)
	mx := 0.0
	my := float64(glob.ScreenHeight) - 4.0
	ebitenutil.DrawRect(screen, mx-1, my-(float64(tRect.Dy()-1)), float64(tRect.Dx()+4), float64(tRect.Dy()+3), glob.ColorToolTipBG)
	text.Draw(screen, dbuf, glob.ToolTipFont, int(mx), int(my), glob.ColorAqua)
}

func drawWorldTooltip(screen *ebiten.Image) {
	/* Get mouse position on world */
	worldMouseX := (glob.MouseX/glob.ZoomScale + (glob.CameraX - (float64(glob.ScreenWidth)/2.0)/glob.ZoomScale))
	worldMouseY := (glob.MouseY/glob.ZoomScale + (glob.CameraY - (float64(glob.ScreenHeight)/2.0)/glob.ZoomScale))

	/* Toolbar tool tip */
	uipix := float64(ToolbarMax * int(gv.ToolBarScale))
	if glob.MouseX <= uipix+gv.ToolBarOffsetX && glob.MouseY <= gv.ToolBarScale+gv.ToolBarOffsetY {
		val := int(glob.MouseX / gv.ToolBarScale)
		if val >= 0 && val < ToolbarMax {

			item := ToolbarItems[int(glob.MouseX/gv.ToolBarScale)]

			toolTip := fmt.Sprintf("%v", item.OType.Name)
			tRect := text.BoundString(glob.ToolTipFont, toolTip)
			mx := glob.MouseX + 20
			my := glob.MouseY + 20
			ebitenutil.DrawRect(screen, mx-1, my-(float64(tRect.Dy()-1)), float64(tRect.Dx()+4), float64(tRect.Dy()+3), glob.ColorToolTipBG)
			text.Draw(screen, toolTip, glob.ToolTipFont, int(mx), int(my), glob.ColorAqua)
		}
	} else {
		/* World Obj tool tip */
		pos := util.FloatXYToPosition(worldMouseX, worldMouseY)
		chunk := util.GetChunk(pos)

		toolTip := ""
		found := false

		if chunk != nil {
			o := util.GetObj(pos, chunk)
			if o != nil {
				found = true
				toolTip = fmt.Sprintf("%v: (%v,%v)\n",
					o.TypeP.Name,
					humanize.Comma(int64(math.Floor(worldMouseX-gv.XYCenter))),
					humanize.Comma(int64(math.Floor(worldMouseY-gv.XYCenter))))
				for z := 0; z < gv.MAT_MAX; z++ {
					if o.Contents[z] != nil {
						toolTip = toolTip + fmt.Sprintf("(Contents: %v: %v%v)\n",
							o.Contents[z].TypeP.Name, o.Contents[z].Amount, o.Contents[z].TypeP.UnitName)
					}
				}
				if gv.Debug {
					for z := 0; z < gv.DIR_MAX; z++ {
						if o.Ports[z].PortDir == gv.PORT_INPUT && o.Ports[z].Obj != nil {
							toolTip = toolTip + fmt.Sprintf("(Input: %v: %v: %v: %v)\n",
								util.DirToName(uint8(z)),
								o.Ports[z].Obj.TypeP.Name,
								o.Ports[z].Buf.TypeP.Name, o.Ports[z].Buf.Amount)
						}
						if o.Ports[z].PortDir == gv.PORT_OUTPUT && o.Ports[z].Obj != nil {
							toolTip = toolTip + fmt.Sprintf("(Output: %v: %v: %v: %v)\n",
								util.DirToName(uint8(z)),
								o.Ports[z].Obj.TypeP.Name,
								o.Ports[z].Buf.TypeP.Name, o.Ports[z].Buf.Amount)
						}
					}
				}
			}
		}

		/* No object contents found, just show x/y */
		if !found {
			toolTip = fmt.Sprintf("(%v, %v)",
				humanize.Comma(int64(math.Floor(worldMouseX-gv.XYCenter))),
				humanize.Comma(int64(math.Floor(worldMouseY-gv.XYCenter))))
		}

		tRect := text.BoundString(glob.ToolTipFont, toolTip)
		mx := glob.MouseX + 20
		my := glob.MouseY + 20
		ebitenutil.DrawRect(screen, mx-1, my-15, float64(tRect.Dx()+4), float64(tRect.Dy()+3), glob.ColorToolTipBG)
		text.Draw(screen, toolTip, glob.ToolTipFont, int(mx), int(my), glob.ColorAqua)
	}
}

/* Draw world objects */
func drawObject(screen *ebiten.Image, obj *glob.ObjData) {

	/* camera + object */
	objOffX := camXPos + (float64(obj.Pos.X))
	objOffY := camYPos + (float64(obj.Pos.Y))

	/* camera zoom */
	x := objOffX * glob.ZoomScale
	y := objOffY * glob.ZoomScale

	/* Draw sprite */
	if obj.TypeP.Image == nil {
		return
	} else {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

		iSize := obj.TypeP.Image.Bounds()

		/*
			if consts.Debug {
				op.ColorM.Reset()
				if obj.BlinkRed > 0 {
					op.ColorM.Scale(1, 0, 0, 1)
					obj.BlinkRed--
				}
				if obj.BlinkGreen > 0 {
					op.ColorM.Scale(0, 1, 0, 1)
					obj.BlinkGreen--
				}
			} */

		if obj.TypeP.Rotatable {
			xx := float64(iSize.Size().X / 2)
			yy := float64(iSize.Size().Y / 2)
			op.GeoM.Translate(-xx, -yy)
			op.GeoM.Rotate(cNinetyDeg * float64(int(obj.Dir)))
			op.GeoM.Translate(xx, yy)
		}

		op.GeoM.Scale(
			(float64(obj.TypeP.Size.X)*glob.ZoomScale)/float64(iSize.Max.X),
			(float64(obj.TypeP.Size.Y)*glob.ZoomScale)/float64(iSize.Max.Y))

		op.GeoM.Translate(math.Floor(x), math.Floor(y))

		if gv.Verbose {
			cwlog.DoLog("%v,%v (%v)", x, y, (float64(obj.TypeP.Size.X)*glob.ZoomScale)/float64(iSize.Max.X))
		}
		screen.DrawImage(obj.TypeP.Image, op)

	}

}

/* Update local vars with camera position calculations */
func calcScreenCamera() {
	/* Adjust cam position for zoom */
	camXPos = float64(-glob.CameraX) + ((float64(glob.ScreenWidth) / 2.0) / glob.ZoomScale)
	camYPos = float64(-glob.CameraY) + ((float64(glob.ScreenHeight) / 2.0) / glob.ZoomScale)

	/* Get camera bounds */
	camStartX = int((1/glob.ZoomScale + (glob.CameraX - (float64(glob.ScreenWidth)/2.0)/glob.ZoomScale)))
	camStartY = int((1/glob.ZoomScale + (glob.CameraY - (float64(glob.ScreenHeight)/2.0)/glob.ZoomScale)))
	camEndX = int((float64(glob.ScreenWidth)/glob.ZoomScale + (glob.CameraX - (float64(glob.ScreenWidth)/2.0)/glob.ZoomScale)))
	camEndY = int((float64(glob.ScreenHeight)/glob.ZoomScale + (glob.CameraY - (float64(glob.ScreenHeight)/2.0)/glob.ZoomScale)))

	/* Pre-calc camera chunk position */
	screenStartX = camStartX / gv.ChunkSize
	screenStartY = camStartY / gv.ChunkSize
	screenEndX = camEndX / gv.ChunkSize
	screenEndY = camEndY / gv.ChunkSize
}

/* Draw materials on belts */
func drawMaterials(m *glob.MatData, obj *glob.ObjData, op *ebiten.DrawImageOptions, screen *ebiten.Image) {

	if m.Amount > 0 {
		img := m.TypeP.Image
		if img != nil {
			screen.DrawImage(img, op)
		}
	}
}
