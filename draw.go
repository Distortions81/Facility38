package main

import (
	"GameTest/consts"
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
	cPreCache               = 2
)

var (
	/* Visible Chunk Cache */
	gVisChunks   [consts.MAX_DRAW_CHUNKS]*glob.MapChunk
	gVisChunkPos [consts.MAX_DRAW_CHUNKS]glob.XY
	gVisChunkTop int

	gVisSChunks   [consts.MAX_DRAW_CHUNKS]*glob.MapSuperChunk
	gVisSChunkPos [consts.MAX_DRAW_CHUNKS]glob.XY
	gVisSChunkTop int

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
)

func init() {
	glob.MiniMapTile = ebiten.NewImage(1, 1)
	glob.MiniMapTile.Fill(color.White)

	glob.ToolBG = ebiten.NewImage(consts.ToolBarScale, consts.ToolBarScale)
	glob.ToolBG.Fill(glob.ColorCharcol)
}

func calcScreenCamera() {
	/* Adjust cam position for zoom */
	camXPos = float64(-glob.CameraX) + (float64(glob.ScreenWidth/2) / glob.ZoomScale)
	camYPos = float64(-glob.CameraY) + (float64(glob.ScreenHeight/2) / glob.ZoomScale)

	/* Get camera bounds */
	camStartX = int((1/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale)))
	camStartY = int((1/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale)))
	camEndX = int((float64(glob.ScreenWidth)/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale)))
	camEndY = int((float64(glob.ScreenHeight)/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale)))

	/* Pre-calc camera chunk position */
	screenStartX = camStartX / consts.ChunkSize
	screenStartY = camStartY / consts.ChunkSize
	screenEndX = camEndX / consts.ChunkSize
	screenEndY = camEndY / consts.ChunkSize
}

func makeVisList() {
	/* When needed, make a list of chunks to draw */
	if glob.CameraDirty {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

		gVisChunkTop = 0
		gVisSChunkTop = 0

		superChunksDrawn = 0
		for scPos, sChunk := range glob.SuperChunkMap {

			if glob.ZoomScale > consts.MapPixelThreshold {
				if sChunk.MapImg != nil {
					sChunk.MapImg.Dispose()
					sChunk.MapImg = nil
				}
			}
			/* Is this super chunk on the screen? */
			if scPos.X < screenStartX/consts.SuperChunkSize ||
				scPos.X > screenEndX/consts.SuperChunkSize ||
				scPos.Y < screenStartY/consts.SuperChunkSize ||
				scPos.Y > screenEndY/consts.SuperChunkSize {
				if sChunk.MapImg != nil {
					sChunk.MapImg.Dispose()
					sChunk.MapImg = nil
				}
				sChunk.Visible = false
				continue
			}

			superChunksDrawn++
			sChunk.Visible = true

			/* Make Pixelmap images */
			if sChunk.MapImg == nil {
				sChunk.MapImg = ebiten.NewImage(consts.SuperChunkPixels, consts.SuperChunkPixels)
				sChunk.MapImg.Fill(glob.ColorCharcol)
				for _, ctmp := range sChunk.Chunks {
					if ctmp.NumObjects <= 0 {
						continue
					}

					/* Draw objects in chunk */
					for objPos, _ := range ctmp.WObject {
						scX := (((scPos.X) * (consts.SuperChunkPixels)) - consts.XYCenter)
						scY := (((scPos.Y) * (consts.SuperChunkPixels)) - consts.XYCenter)

						x := float64((objPos.X - consts.XYCenter) - scX)
						y := float64((objPos.Y - consts.XYCenter) - scY)
						op.GeoM.Reset()
						op.GeoM.Translate(x, y)
						sChunk.MapImg.DrawImage(glob.MiniMapTile, op)
					}
				}
			}
			if gVisSChunkTop < consts.MAX_DRAW_CHUNKS {
				gVisSChunks[gVisSChunkTop] = sChunk
				gVisSChunkPos[gVisSChunkTop] = scPos
				gVisSChunkTop++
			} else {
				break
			}

			for chunkPos, chunk := range sChunk.Chunks {

				/* Is this chunk in the prerender area?
				if chunkPos.X+cPreCache < screenStartX ||
					chunkPos.X-cPreCache > screenEndX ||
					chunkPos.Y+cPreCache < screenStartY ||
					chunkPos.Y-cPreCache > screenEndY {
					chunk.Visible = false
					continue
				}
				chunk.Visible = true */

				/* Is this chunk on the screen? */
				if chunkPos.X < screenStartX ||
					chunkPos.X > screenEndX ||
					chunkPos.Y < screenStartY ||
					chunkPos.Y > screenEndY {
					chunk.Visible = false
					continue
				}
				chunk.Visible = true

				if gVisChunkTop < consts.MAX_DRAW_CHUNKS {
					gVisChunks[gVisChunkTop] = chunk
					gVisChunkPos[gVisChunkTop] = chunkPos
					gVisChunkTop++
				} else {
					break
				}
				glob.CameraDirty = false
			}
		}
	}
}

var ready bool = false

func (g *Game) Draw(screen *ebiten.Image) {

	if !glob.MapGenerated ||
		!glob.SpritesLoaded ||
		!glob.PlayerReady {
		bootScreen(screen)
		time.Sleep(time.Millisecond * 125) //8 fps
		return
	}

	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

	/* Draw start */
	drawStart := time.Now()

	calcScreenCamera()

	glob.SuperChunkMapLock.Lock()

	makeVisList()

	chunksDrawn := 0

	if glob.ZoomScale > consts.MapPixelThreshold { /* Draw icon mode */
		for i := 0; i < gVisChunkTop; i++ {
			chunkPos := gVisChunkPos[i]
			chunk := gVisChunks[i]
			chunksDrawn++

			/* Draw ground */
			/* No image, use a temporary texture while it draws */
			if chunk.TerrainImg == nil {
				chunk.TerrainImg = glob.TempChunkImage
				chunk.UsingTemporary = true
			}

			iSize := chunk.TerrainImg.Bounds().Size()
			op.GeoM.Reset()
			op.GeoM.Scale((consts.ChunkSize*glob.ZoomScale)/float64(iSize.X), (consts.ChunkSize*glob.ZoomScale)/float64(iSize.Y))
			op.GeoM.Translate((camXPos+float64(chunkPos.X*consts.ChunkSize))*glob.ZoomScale, (camYPos+float64(chunkPos.Y*consts.ChunkSize))*glob.ZoomScale)
			screen.DrawImage(chunk.TerrainImg, op)

			/* Draw objects in chunk */
			for objPos, obj := range chunk.WObject {

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
						revDir := util.ReverseDirection(obj.Direction)
						if obj.OutputObj != nil && obj.OutputBuffer.Amount > 0 &&
							obj.OutputObj.InputBuffer[revDir] != nil && obj.OutputObj.InputBuffer[revDir].Amount > 0 {

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
		/* Draw superchunk images */
		for z := 0; z < gVisSChunkTop; z++ {
			sChunk := gVisSChunks[z]
			cPos := gVisSChunkPos[z]

			op.GeoM.Reset()
			op.GeoM.Scale(
				(consts.SuperChunkPixels*glob.ZoomScale)/float64(consts.SuperChunkPixels),
				(consts.SuperChunkPixels*glob.ZoomScale)/float64(consts.SuperChunkPixels))

			op.GeoM.Translate(
				((camXPos+float64((cPos.X))*consts.SuperChunkPixels)*glob.ZoomScale)-1,
				((camYPos+float64((cPos.Y))*consts.SuperChunkPixels)*glob.ZoomScale)-1)

			screen.DrawImage(sChunk.MapImg, op)
		}
	}

	glob.SuperChunkMapLock.Unlock()

	/* Get mouse position on world */
	worldMouseX := (glob.MouseX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
	worldMouseY := (glob.MouseY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))

	/* Draw debug info */
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("FPS: %.2f UPS: %.2f Workers: %v Job-size: %v Active Objects: %v CDraw: %v SCDraw: %v Arch: %v Build: %v",
			ebiten.ActualFPS(),
			1000000000.0/float64(glob.MeasuredObjectUPS_ns),
			objects.NumWorkers,
			humanize.SIWithDigits(float64(objects.TockWorkSize), 2, ""),
			humanize.SIWithDigits(float64(objects.TockWorkSize*objects.NumWorkers), 2, ""),
			chunksDrawn,
			superChunksDrawn,
			runtime.GOARCH, buildTime),
		0, glob.ScreenHeight-20)

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
		chunk := util.GetChunk(&pos)

		toolTip := ""
		found := false
		if chunk != nil {
			o := chunk.WObject[pos]
			if o != nil {
				found = true
				toolTip = fmt.Sprintf("(%v,%v) %v", humanize.Comma(int64(worldMouseX-consts.XYCenter)), humanize.Comma(int64(worldMouseY-consts.XYCenter)), o.TypeP.Name)
				toolTip = toolTip + fmt.Sprintf(" (OutputBuf: %v-%v)",
					o.OutputBuffer.TypeP.Name,
					o.OutputBuffer.Amount)
				if o.OutputObj != nil {
					rev := util.ReverseDirection(o.Direction)
					toolTip = toolTip + fmt.Sprintf(" (OutputObj: %v, Contains: %v-%v, Dir: %v)",
						o.OutputObj.TypeP.Name,
						o.OutputObj.InputBuffer[rev].Amount,
						o.OutputObj.InputBuffer[rev].TypeP.Name,
						util.DirToName(o.Direction))
				}
				for z := consts.DIR_NORTH; z < consts.DIR_NONE; z++ {
					glob.SuperChunkMapLock.Lock()
					if o.InputBuffer[z] != nil {
						zPos := util.FindObj(o.InputObjs[z])
						zX := zPos.X
						zY := zPos.Y
						toolTip = toolTip + fmt.Sprintf(" (InputBuf: %v: %v (%v,%v)), Contains: %v-%v)",
							o.InputObjs[z].TypeP.Name,
							util.DirToName(z),
							zX, zY,
							o.InputBuffer[z].Amount,
							o.InputBuffer[z].TypeP.Name)
					}
					glob.SuperChunkMapLock.Unlock()
				}

				found := false
				for _, t := range o.Contents {
					if t != nil && t.Amount > 0 {
						found = true
						toolTip += fmt.Sprintf(" Contents: %v: %v", t.TypeP.Name, humanize.SIWithDigits(float64(t.Amount), 2, ""))
					}
				}
				if !found {
					for _, t := range o.InputBuffer {
						if t != nil && t.Amount > 0 {
							//toolTip += fmt.Sprintf(" Contents: %v: %v", t.TypeP.Name, humanize.SIWithDigits(float64(t.Amount), 2, ""))
						}
					}
				}

			}
		}

		/* No object contents found */
		if !found {
			toolTip = fmt.Sprintf("(%v, %v)", humanize.Comma(int64(worldMouseX-consts.XYCenter)), humanize.Comma(int64(worldMouseY-consts.XYCenter)))
		}

		tRect := text.BoundString(glob.ToolTipFont, toolTip)
		mx := glob.MouseX + 20
		my := glob.MouseY + 20
		ebitenutil.DrawRect(screen, mx-1, my-(float64(tRect.Dy()-1)), float64(tRect.Dx()+4), float64(tRect.Dy()+3), glob.ColorToolTipBG)
		text.Draw(screen, toolTip, glob.ToolTipFont, int(mx), int(my), glob.ColorAqua)
	}

	/* Limit frame rate */
	terrain.RenderTerrain()
	if g.ui != nil {
		g.ui.Draw(screen)
	}

	sleepFor := cMAX_RENDER_NS - time.Since(drawStart)
	time.Sleep(sleepFor)

}

func drawMaterials(m *glob.MatData, obj *glob.WObject, op *ebiten.DrawImageOptions, screen *ebiten.Image) {

	if m.Amount > 0 {
		img := m.TypeP.Image
		if img != nil {
			screen.DrawImage(img, op)
		}
	}
}

func drawObject(screen *ebiten.Image, objPos glob.XY, obj *glob.WObject) {

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

		//fmt.Printf("%v,%v (%v)\n", x, y, (float64(obj.TypeP.Size.X)*glob.ZoomScale)/float64(iSize.Max.X))
		screen.DrawImage(obj.TypeP.Image, op)

	}

}
