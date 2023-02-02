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
	cPreCache               = 3
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

	superChunksDrawn int
	frameCount       uint64
)

func init() {
	glob.MiniMapTile = ebiten.NewImage(1, 1)
	glob.MiniMapTile.Fill(color.White)

	glob.ToolBG = ebiten.NewImage(consts.ToolBarScale, consts.ToolBarScale)
	glob.ToolBG.Fill(glob.ColorCharcol)
}

func makeVisList() {

	/* When needed, make a list of chunks to draw */
	if glob.CameraDirty {

		glob.VisChunkLock.Lock()
		defer glob.VisChunkLock.Unlock()

		glob.VisChunkTop = 0
		glob.VisSChunkTop = 0

		superChunksDrawn = 0
		SChunkTmp := glob.SuperChunkList
		for _, sChunk := range SChunkTmp {
			scPos := sChunk.Pos

			if sChunk.NumChunks == 0 {
				continue
			}

			/* Is this super chunk on the screen? */
			if scPos.X < screenStartX/consts.SuperChunkSize ||
				scPos.X > screenEndX/consts.SuperChunkSize ||
				scPos.Y < screenStartY/consts.SuperChunkSize ||
				scPos.Y > screenEndY/consts.SuperChunkSize {
				sChunk.Visible = false
				continue
			}

			superChunksDrawn++
			sChunk.Visible = true

			if glob.VisSChunkTop < consts.MAX_DRAW_CHUNKS {
				glob.VisSChunks[glob.VisSChunkTop] = sChunk
				glob.VisSChunkPos[glob.VisSChunkTop] = scPos
				glob.VisSChunkTop++
			} else {
				break
			}

			sChunkTmp := sChunk.ChunkList
			for _, chunk := range sChunkTmp {
				chunkPos := chunk.Pos

				if sChunk.NumChunks == 0 {
					continue
				}

				/* Is this chunk in the prerender area? */
				if chunkPos.X+cPreCache < screenStartX ||
					chunkPos.X-cPreCache > screenEndX ||
					chunkPos.Y+cPreCache < screenStartY ||
					chunkPos.Y-cPreCache > screenEndY {
					chunk.Precache = false
					continue
				}
				chunk.Precache = true

				/* Is this chunk on the screen? */
				if chunkPos.X < screenStartX ||
					chunkPos.X > screenEndX ||
					chunkPos.Y < screenStartY ||
					chunkPos.Y > screenEndY {
					chunk.Visible = false
					continue
				}
				chunk.Visible = true

				if glob.VisChunkTop < consts.MAX_DRAW_CHUNKS {
					glob.VisChunks[glob.VisChunkTop] = chunk
					glob.VisChunkPos[glob.VisChunkTop] = chunkPos
					glob.VisChunkTop++
				} else {
					break
				}
				glob.CameraDirty = false
			}
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {

	if !glob.MapGenerated ||
		!glob.SpritesLoaded ||
		!glob.PlayerReady {
		bootScreen(screen)
		time.Sleep(time.Millisecond * 125) //8 fps
		return
	}

	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

	frameCount++

	/* Draw start */
	drawStart := time.Now()
	calcScreenCamera()
	makeVisList()

	chunksDrawn := 0

	if glob.ZoomScale > consts.MapPixelThreshold { /* Draw icon mode */

		glob.VisChunkLock.RLock()
		VisChunkTmp := glob.VisChunks
		VisChunkPosTmp := glob.VisChunkPos
		glob.VisChunkLock.RUnlock()

		for index, chunk := range VisChunkTmp {
			if chunk == nil || (chunk.Precache && !chunk.Visible) {
				continue
			}

			chunkPos := VisChunkPosTmp[index]
			chunksDrawn++

			/* Draw ground */
			/* No image, use a temporary texture while it draws */
			if chunk.TerrainImg == nil {
				chunk.TerrainLock.Lock()
				chunk.TerrainImg = glob.TempChunkImage
				chunk.UsingTemporary = true
				chunk.TerrainLock.Unlock()
			}

			chunk.TerrainLock.Lock()
			iSize := chunk.TerrainImg.Bounds().Size()
			op.GeoM.Reset()
			op.GeoM.Scale((consts.ChunkSize*glob.ZoomScale)/float64(iSize.X), (consts.ChunkSize*glob.ZoomScale)/float64(iSize.Y))
			op.GeoM.Translate((camXPos+float64(chunkPos.X*consts.ChunkSize))*glob.ZoomScale, (camYPos+float64(chunkPos.Y*consts.ChunkSize))*glob.ZoomScale)
			screen.DrawImage(chunk.TerrainImg, op)
			chunk.TerrainLock.Unlock()

			/* Draw objects in chunk */
			ObjListTmp := chunk.ObjList
			for _, obj := range ObjListTmp {
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
	} else {

		if glob.FixWASM {
			terrain.PixmapRenderST()
		}

		/* Draw superchunk images */

		glob.VisSChunkLock.RLock()
		VisSChunkTmp := glob.VisSChunks
		visSChunkPosTmp := glob.VisSChunkPos
		glob.VisSChunkLock.RUnlock()

		for index, sChunk := range VisSChunkTmp {
			if sChunk == nil {
				continue
			}

			sChunk.PixLock.Lock()
			if !sChunk.Visible || sChunk.PixMap == nil {
				sChunk.PixLock.Unlock()
				continue
			}
			sChunk.PixLock.Unlock()

			cPos := visSChunkPosTmp[index]

			op.GeoM.Reset()
			op.GeoM.Scale(
				(consts.SuperChunkPixels*glob.ZoomScale)/float64(consts.SuperChunkPixels),
				(consts.SuperChunkPixels*glob.ZoomScale)/float64(consts.SuperChunkPixels))

			op.GeoM.Translate(
				((camXPos+float64((cPos.X))*consts.SuperChunkPixels)*glob.ZoomScale)-1,
				((camYPos+float64((cPos.Y))*consts.SuperChunkPixels)*glob.ZoomScale)-1)

			sChunk.PixLock.Lock()
			screen.DrawImage(sChunk.PixMap, op)
			sChunk.PixLock.Unlock()
		}
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

		/* No object contents found */
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
	if glob.FixWASM && frameCount%WASMTerrtainDiv == 0 {
		terrain.RenderTerrainST()
	}
	if g.ui != nil {
		g.ui.Draw(screen)
	}

	sleepFor := cMAX_RENDER_NS - time.Since(drawStart)
	time.Sleep(sleepFor)

}

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
		}

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

func drawMaterials(m *glob.MatData, obj *glob.ObjData, op *ebiten.DrawImageOptions, screen *ebiten.Image) {

	if m.Amount > 0 {
		img := m.TypeP.Image
		if img != nil {
			screen.DrawImage(img, op)
		}
	}
}
