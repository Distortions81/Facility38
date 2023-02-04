package main

import (
	"GameTest/consts"
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/objects"
	"GameTest/terrain"
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

	glob.ToolBG = ebiten.NewImage(consts.ToolBarScale, consts.ToolBarScale)
	glob.ToolBG.Fill(glob.ColorCharcol)
}

/* Look at camera position, make a list of visible superchunks and chunks. Saves to VisChunks, checks glob.CameraDirty */
func updateVisData() {

	/* When needed, make a list of chunks to draw */
	if glob.VisDataDirty.Load() {

		glob.SuperChunkListLock.RLock()
		for _, sChunk := range glob.SuperChunkList {

			if sChunk.NumChunks == 0 {
				sChunk.Visible = false
				continue
			}

			/* Is this super chunk on the screen? */
			if sChunk.Pos.X < screenStartX/consts.SuperChunkSize ||
				sChunk.Pos.X > screenEndX/consts.SuperChunkSize ||
				sChunk.Pos.Y < screenStartY/consts.SuperChunkSize ||
				sChunk.Pos.Y > screenEndY/consts.SuperChunkSize {
				sChunk.Visible = false
				continue
			}

			sChunk.Visible = true

			for _, chunk := range sChunk.ChunkList {

				if chunk.NumObjects == 0 {
					chunk.Visible = false
					chunk.Precache = false
					continue
				}

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

/* Ebiten: Draw everything */
func (g *Game) Draw(screen *ebiten.Image) {

	if !glob.MapGenerated.Load() ||
		!glob.SpritesLoaded.Load() ||
		!glob.PlayerReady.Load() {
		bootScreen(screen)
		if glob.WASMMode {
			time.Sleep(time.Millisecond * 250) //4 fps
		}
		return
	}

	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

	frameCount++

	/* Draw start */
	drawStart := time.Now()
	calcScreenCamera()
	updateVisData()

	chunksDrawn := 0

	if glob.ZoomScale > consts.MapPixelThreshold { /* Draw icon mode */

		glob.SuperChunkListLock.RLock()
		for _, sChunk := range glob.SuperChunkList {

			for _, chunk := range sChunk.ChunkList {
				if !chunk.Visible {
					continue
				}
				chunksDrawn++

				/* Draw ground */
				chunk.TerrainLock.RLock()
				cTmp := chunk.TerrainImg
				if chunk.TerrainImg == nil {
					cTmp = glob.TempChunkImage
				}

				iSize := cTmp.Bounds().Size()
				op.GeoM.Reset()
				op.GeoM.Scale((consts.ChunkSize*glob.ZoomScale)/float64(iSize.X), (consts.ChunkSize*glob.ZoomScale)/float64(iSize.Y))
				op.GeoM.Translate((camXPos+float64(chunk.Pos.X*consts.ChunkSize))*glob.ZoomScale, (camYPos+float64(chunk.Pos.Y*consts.ChunkSize))*glob.ZoomScale)
				screen.DrawImage(cTmp, op)
				chunk.TerrainLock.RUnlock()

				/* Draw objects in chunk */
				for _, obj := range chunk.ObjList {
					if obj == nil {
						continue
					}
					objPos := obj.Pos

					/* Is this object on the screen? */
					if objPos.X < camStartX || objPos.X > camEndX || objPos.Y < camStartY || objPos.Y > camEndY {
						continue
					}

					/* Time to draw it */
					drawObject(screen, objPos, obj)

					/* Overlays */
					/* Draw belt overlays */
					if obj.TypeP.TypeI == consts.ObjTypeBasicBelt {

						/* camera + object */
						objOffX := camXPos + (float64(objPos.X))
						objOffY := camYPos + (float64(objPos.Y))

						/* camera zoom */
						objCamPosX := objOffX * glob.ZoomScale
						objCamPosY := objOffY * glob.ZoomScale

						iSize := obj.TypeP.Image.Bounds()
						op.GeoM.Reset()
						op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X), ((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
						op.GeoM.Translate(objCamPosX, objCamPosY)

						/* Draw Input Materials */
						if obj.OutputBuffer.Amount > 0 {
							drawMaterials(obj.OutputBuffer, obj, op, screen)
						} else {
							for _, m := range obj.InputBuffer {
								if m != nil && m.Amount > 0 {
									drawMaterials(m, obj, op, screen)
									break
								}
							}
						}

					}
					if glob.ShowInfoLayer {
						/* Info Overlays, such as arrows and blocked indicator */

						/* camera + object */
						objOffX := camXPos + (float64(objPos.X))
						objOffY := camYPos + (float64(objPos.Y))

						/* camera zoom */
						objCamPosX := objOffX * glob.ZoomScale
						objCamPosY := objOffY * glob.ZoomScale

						iSize := obj.TypeP.Image.Bounds()
						op.GeoM.Reset()
						op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X), ((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
						op.GeoM.Translate(objCamPosX, objCamPosY)

						if obj.TypeP.HasMatOutput {
							iSize := obj.TypeP.Image.Bounds()
							op.GeoM.Reset()
							op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X), ((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
							op.GeoM.Translate(objCamPosX, objCamPosY)
							/* Draw Arrow */
							img := objects.ObjOverlayTypes[obj.Direction].Image
							if img != nil {
								screen.DrawImage(img, op)
							}
							/* Show blocked outputs */
							img = objects.ObjOverlayTypes[consts.ObjOverlayBlocked].Image
							if obj.OutputBuffer.Amount > 0 {

								var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

								iSize := obj.TypeP.Image.Bounds()
								op.GeoM.Reset()
								op.GeoM.Translate(float64(iSize.Max.X)-float64(objects.ObjOverlayTypes[consts.ObjOverlayBlocked].Image.Bounds().Max.X)-cBlockedIndicatorOffset, cBlockedIndicatorOffset)
								op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X), ((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
								op.GeoM.Translate(objCamPosX, objCamPosY)
								screen.DrawImage(img, op)
							}
						}
					}
				}
			}
		}
		glob.SuperChunkListLock.RUnlock()

	} else {

		/* Single thread render terrain for WASM */
		if glob.WASMMode {
			terrain.PixmapRenderST()
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
				(consts.MaxSuperChunk*glob.ZoomScale)/float64(consts.MaxSuperChunk),
				(consts.MaxSuperChunk*glob.ZoomScale)/float64(consts.MaxSuperChunk))

			op.GeoM.Translate(
				((camXPos+float64((sChunk.Pos.X))*consts.MaxSuperChunk)*glob.ZoomScale)-1,
				((camYPos+float64((sChunk.Pos.Y))*consts.MaxSuperChunk)*glob.ZoomScale)-1)

			screen.DrawImage(sChunk.PixMap, op)
			sChunk.PixLock.Unlock()
		}
		glob.SuperChunkListLock.RUnlock()
	}

	/* Get mouse position on world */
	worldMouseX := (glob.MouseX/glob.ZoomScale + (glob.CameraX - (float64(glob.ScreenWidth)/2.0)/glob.ZoomScale))
	worldMouseY := (glob.MouseY/glob.ZoomScale + (glob.CameraY - (float64(glob.ScreenHeight)/2.0)/glob.ZoomScale))

	/* Draw debug info */
	dbuf := fmt.Sprintf("FPS: %.2f UPS: %.2f Active Objects: %v Arch: %v Build: %v",
		ebiten.ActualFPS(),
		1000000000.0/float64(glob.MeasuredObjectUPS_ns),
		humanize.SIWithDigits(float64(objects.TockWorkSize*objects.NumWorkers), 2, ""),
		runtime.GOARCH, buildTime)

	tRect := text.BoundString(glob.ToolTipFont, dbuf)
	mx := 0.0
	my := float64(glob.ScreenHeight) - 4.0
	ebitenutil.DrawRect(screen, mx-1, my-(float64(tRect.Dy()-1)), float64(tRect.Dx()+4), float64(tRect.Dy()+3), glob.ColorToolTipBG)
	text.Draw(screen, dbuf, glob.ToolTipFont, int(mx), int(my), glob.ColorAqua)

	/* Draw toolbar */
	screen.DrawImage(toolbarCache, nil)

	/* Toolbar tool tip */
	uipix := float64(ToolbarMax * int(consts.ToolBarScale))
	if glob.MouseX <= uipix+consts.ToolBarOffsetX && glob.MouseY <= consts.ToolBarScale+consts.ToolBarOffsetY {
		val := int(glob.MouseX / consts.ToolBarScale)
		if val >= 0 && val < ToolbarMax {

			item := ToolbarItems[int(glob.MouseX/consts.ToolBarScale)]

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
				toolTip = fmt.Sprintf("(%v,%v) %v", humanize.Comma(int64(math.Floor(worldMouseX-consts.XYCenter))), humanize.Comma(int64(math.Floor(worldMouseY-consts.XYCenter))), o.TypeP.Name)
				for z := consts.DIR_NORTH; z < consts.DIR_NONE; z++ {
					if o.Contents[z] != nil {
						toolTip = toolTip + fmt.Sprintf(" (Contents: %v: %v)",
							o.Contents[z].TypeP.Name, o.Contents[z].Amount)
					}
				}
				if o.OutputBuffer != nil && consts.Debug {
					toolTip = toolTip + fmt.Sprintf(" (OutputBuf: %v: %v: %v)",
						util.DirToName(o.Direction),
						o.OutputBuffer.TypeP.Name,
						o.OutputBuffer.Amount)
				}
				if o.OutputObj != nil && o.OutputObj.InputBuffer[util.ReverseDirection(o.Direction)] != nil && consts.Debug {
					toolTip = toolTip + fmt.Sprintf(" (OutputObj: %v: %v)",
						util.DirToName(o.Direction), o.OutputObj.TypeP.Name)
				}

				if consts.Debug {
					for z := consts.DIR_NORTH; z < consts.DIR_NONE; z++ {
						if o.InputBuffer[z] != nil {
							toolTip = toolTip + fmt.Sprintf(" (InputBuf: %v: %v: %v)",
								util.DirToName(z),
								o.InputBuffer[z].TypeP.Name,
								o.InputBuffer[z].Amount)
						}
						if o.InputObjs[z] != nil {
							toolTip = toolTip + fmt.Sprintf(" (InputObj: %v: %v)",
								o.InputObjs[z].TypeP.Name,
								util.DirToName(util.ReverseDirection(z)))
						}
					}
				}
			}
		}

		/* No object contents found, just show x/y */
		if !found {
			toolTip = fmt.Sprintf("(%v, %v)", humanize.Comma(int64(math.Floor(worldMouseX-consts.XYCenter))), humanize.Comma(int64(math.Floor(worldMouseY-consts.XYCenter))))
		}

		tRect := text.BoundString(glob.ToolTipFont, toolTip)
		mx := glob.MouseX + 20
		my := glob.MouseY + 20
		ebitenutil.DrawRect(screen, mx-1, my-(float64(tRect.Dy()-1)), float64(tRect.Dx()+4), float64(tRect.Dy()+3), glob.ColorToolTipBG)
		text.Draw(screen, toolTip, glob.ToolTipFont, int(mx), int(my), glob.ColorAqua)

	}

	/* Limit frame rate */
	if glob.WASMMode && frameCount%WASMTerrtainDiv == 0 {
		terrain.RenderTerrainST()
	}

	/* UI: WIP */
	if g.ui != nil {
		g.ui.Draw(screen)
	}

	/* Throttle if needed */
	sleepFor := cMAX_RENDER_NS - time.Since(drawStart)
	time.Sleep(sleepFor)

}

/* Draw world objects */
func drawObject(screen *ebiten.Image, objPos glob.XY, obj *glob.ObjData) {

	/* camera + object */
	objOffX := camXPos + (float64(objPos.X))
	objOffY := camYPos + (float64(objPos.Y))

	/* camera zoom */
	x := objOffX * glob.ZoomScale
	y := objOffY * glob.ZoomScale

	/* Draw sprite */
	if obj.TypeP.Image == nil {
		return
	} else {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

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

		if obj.TypeP.Rotatable && obj.Direction > 0 {
			x := float64(iSize.Size().X / 2)
			y := float64(iSize.Size().Y / 2)
			op.GeoM.Translate(-x, -y)
			op.GeoM.Rotate(cNinetyDeg * float64(obj.Direction))
			op.GeoM.Translate(x, y)
		}

		op.GeoM.Scale(
			(float64(obj.TypeP.Size.X)*glob.ZoomScale)/float64(iSize.Max.X),
			(float64(obj.TypeP.Size.Y)*glob.ZoomScale)/float64(iSize.Max.Y))

		op.GeoM.Translate(math.Floor(x), math.Floor(y))

		if consts.Verbose {
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
	screenStartX = camStartX / consts.ChunkSize
	screenStartY = camStartY / consts.ChunkSize
	screenEndX = camEndX / consts.ChunkSize
	screenEndY = camEndY / consts.ChunkSize
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
