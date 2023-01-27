package main

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/objects"
	"GameTest/util"
	"fmt"
	"math"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

var toolBG *ebiten.Image

func (g *Game) Draw(screen *ebiten.Image) {

	drawStart := time.Now()

	if !glob.DrewMap {
		screen.DrawImage(glob.BootImage, nil)
		return
	}

	/* Draw start */

	/* Adjust cam position for zoom */
	camXPos := float64(-glob.CameraX) + (float64(glob.ScreenWidth/2) / glob.ZoomScale)
	camYPos := float64(-glob.CameraY) + (float64(glob.ScreenHeight/2) / glob.ZoomScale)

	/* Get camera bounds */
	camStartX := int((1/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale)))
	camStarty := int((1/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale)))
	camEndX := int((float64(glob.ScreenWidth)/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale)))
	camEndY := int((float64(glob.ScreenHeight)/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale)))

	/* Pre-calc camera chunk position */
	chunkStartX := camStartX / consts.ChunkSize
	chunkStartY := camStarty / consts.ChunkSize
	chunkEndX := camEndX / consts.ChunkSize
	chunkEndY := camEndY / consts.ChunkSize

	glob.WorldMapLock.Lock()

	/* When needed, make a list of chunks to draw */
	if glob.CameraDirty {
		glob.ListTop = 0
		for chunkPos, chunk := range glob.WorldMap {

			/* Is this chunk on the screen? */
			if chunkPos.X < chunkStartX || chunkPos.X > chunkEndX || chunkPos.Y < chunkStartY || chunkPos.Y > chunkEndY {
				chunk.Visible = false
				continue
			}
			chunk.Visible = true
			chunk.LastSaw = time.Now()

			if glob.ListTop < consts.MAX_DRAW_CHUNKS {
				glob.CameraList[glob.ListTop] = chunk
				glob.XYList[glob.ListTop] = chunkPos
				glob.ListTop++
			} else {
				break
			}
			glob.CameraDirty = false
		}
	}

	/* Draw world */
	if glob.ZoomScale < consts.SpriteScale {
		//Pixel Mode
		screen.Fill(glob.ColorCharcol)

		for i := 0; i < glob.ListTop; i++ {
			chunk := glob.CameraList[i]

			/* Draw objects in chunk */
			for objPos, obj := range chunk.WObject {

				/* Is this object on the screen? */
				if objPos.X < camStartX || objPos.X > camEndX || objPos.Y < camStarty || objPos.Y > camEndY {
					continue
				}

				/* camera + object */
				objOffX := camXPos + (float64(objPos.X))
				objOffY := camYPos + (float64(objPos.Y))

				/* camera zoom */
				objCamPosX := objOffX * glob.ZoomScale
				objCamPosY := objOffY * glob.ZoomScale

				/* Time to draw it */
				DrawObject(screen, objCamPosX, objCamPosY, obj, true)
			}
		}

	} else {
		screen.Fill(glob.ColorBlack)
		for i := 0; i < glob.ListTop; i++ {
			chunkPos := glob.XYList[i]
			chunk := glob.CameraList[i]

			/* Draw ground */
			var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

			op.GeoM.Reset()
			chunk.GroundLock.Lock()
			if chunk.GroundImg != nil {
				iSize := chunk.GroundImg.Bounds().Size()
				op.GeoM.Scale((consts.ChunkSize*glob.ZoomScale)/float64(iSize.X), (consts.ChunkSize*glob.ZoomScale)/float64(iSize.Y))
				op.GeoM.Translate((camXPos+float64(chunkPos.X*consts.ChunkSize))*glob.ZoomScale, (camYPos+float64(chunkPos.Y*consts.ChunkSize))*glob.ZoomScale)
				screen.DrawImage(chunk.GroundImg, op)
			}
			chunk.GroundLock.Unlock()

			/* Draw objects in chunk */
			for objPos, obj := range chunk.WObject {

				/* Is this object on the screen? */
				if objPos.X < camStartX || objPos.X > camEndX || objPos.Y < camStarty || objPos.Y > camEndY {
					continue
				}

				/* camera + object */
				objOffX := camXPos + (float64(objPos.X))
				objOffY := camYPos + (float64(objPos.Y))

				/* camera zoom */
				objCamPosX := objOffX * glob.ZoomScale
				objCamPosY := objOffY * glob.ZoomScale

				/* Time to draw it */
				DrawObject(screen, objCamPosX, objCamPosY, obj, false)
			}
		}

		/* Draw overlays */
		for i := 0; i < glob.ListTop; i++ {

			for objPos, obj := range glob.CameraList[i].WObject {

				/* Is this object on the screen? */
				if objPos.X < camStartX || objPos.X > camEndX || objPos.Y < camStarty || objPos.Y > camEndY {
					continue
				}

				/* camera + object */
				objOffX := camXPos + (float64(objPos.X))
				objOffY := camYPos + (float64(objPos.Y))

				/* camera zoom */
				objCamPosX := objOffX * glob.ZoomScale
				objCamPosY := objOffY * glob.ZoomScale

				/* Overlays */
				/* Draw belt overlays */
				if obj.TypeP.TypeI == consts.ObjTypeBasicBelt {

					var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
					op.GeoM.Reset()
					iSize := obj.TypeP.Image.Bounds()
					op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X), ((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
					op.GeoM.Translate(objCamPosX, objCamPosY)

					/* Draw Input Materials */
					if obj.OutputBuffer.Amount > 0 {
						matTween(obj.OutputBuffer, obj, op, screen)
					} else {
						for _, m := range obj.InputBuffer {
							if m != nil && m.Amount > 0 {
								matTween(m, obj, op, screen)
								break
							}
						}
					}

				} else if glob.ShowInfoLayer {
					/* Info Overlays, such as arrows and blocked indicator */
					var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
					op.GeoM.Reset()
					iSize := obj.TypeP.Image.Bounds()
					op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X), ((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
					op.GeoM.Translate(objCamPosX, objCamPosY)

					if obj.TypeP.HasMatOutput {
						op.GeoM.Reset()
						iSize := obj.TypeP.Image.Bounds()
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
							op.GeoM.Reset()

							iSize := obj.TypeP.Image.Bounds()
							op.GeoM.Translate(float64(iSize.Max.X)-float64(objects.ObjOverlayTypes[consts.ObjOverlayBlocked].Image.Bounds().Max.X)-consts.BlockedIndicatorOffset, consts.BlockedIndicatorOffset)
							op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X), ((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
							op.GeoM.Translate(objCamPosX, objCamPosY)
							screen.DrawImage(img, op)
						}
					}
				}
			}
		}
	}

	glob.WorldMapLock.Unlock()

	/* Get mouse position on world */
	worldMouseX := (glob.MouseX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
	worldMouseY := (glob.MouseY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))

	/* Draw debug info */
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("FPS: %.2f,UPS: %.2f Work: Workers: %v, Job-size: %v, Active Objects: %v, Chunks-Drawn: %v (v%v-%v-%v)",
			ebiten.ActualFPS(), 1000000000.0/float64(glob.MeasuredObjectUPS_ns),
			objects.NumWorkers, humanize.SIWithDigits(float64(objects.TockWorkSize), 2, ""), humanize.SIWithDigits(float64(objects.TockWorkSize*objects.NumWorkers), 2, ""), glob.ListTop,
			consts.Version, consts.Build, glob.DetectedOS),
		0, glob.ScreenHeight-20)

	/* Draw toolbar */
	for i := 0; i < objects.ToolbarMax; i++ {
		DrawToolItem(screen, i)
	}

	/* Toolbar tool tip */
	uipix := float64(objects.ToolbarMax * int(consts.ToolBarScale))
	if glob.MouseX <= uipix+consts.ToolBarOffsetX && glob.MouseY <= consts.ToolBarScale+consts.ToolBarOffsetY {
		val := int(glob.MouseX / consts.ToolBarScale)
		if val >= 0 && val < objects.ToolbarMax {

			item := objects.ToolbarItems[int(glob.MouseX/consts.ToolBarScale)]

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
							toolTip += fmt.Sprintf(" Contents: %v: %v", t.TypeP.Name, humanize.SIWithDigits(float64(t.Amount), 2, ""))
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
	sleepFor := consts.MAX_RENDER_NS - time.Since(drawStart)
	if sleepFor > time.Millisecond {
		time.Sleep(sleepFor)
	}

}

func matTween(m *glob.MatData, obj *glob.WObject, op *ebiten.DrawImageOptions, screen *ebiten.Image) {

	if m.Amount > 0 {
		img := m.TypeP.Image
		if img != nil {
			screen.DrawImage(img, op)
		}
	}
}

func DrawObject(screen *ebiten.Image, x float64, y float64, obj *glob.WObject, miniMap bool) {

	/* Draw sprite */
	if obj.TypeP.Image == nil {
		return
	} else {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
		if miniMap {
			iSize := obj.TypeP.Image.Bounds()
			op.GeoM.Scale((float64(obj.TypeP.Size.X)*glob.ZoomScale)/float64(iSize.Max.X), (float64(obj.TypeP.Size.Y)*glob.ZoomScale)/float64(iSize.Max.Y))
			op.GeoM.Translate(math.Floor(x), math.Floor(y))
			screen.DrawImage(glob.MiniMapTile, op)
		} else {

			iSize := obj.TypeP.Image.Bounds()

			if obj.TypeP.Rotatable && obj.Direction > 0 {
				x := float64(iSize.Size().X / 2)
				y := float64(iSize.Size().Y / 2)
				op.GeoM.Translate(-x, -y)
				op.GeoM.Rotate(consts.NinetyDeg * float64(obj.Direction))
				op.GeoM.Translate(x, y)
			}

			op.GeoM.Scale((float64(obj.TypeP.Size.X)*glob.ZoomScale)/float64(iSize.Max.X), (float64(obj.TypeP.Size.Y)*glob.ZoomScale)/float64(iSize.Max.Y))

			op.GeoM.Translate(math.Floor(x), math.Floor(y))
			screen.DrawImage(obj.TypeP.Image, op)
		}
	}

}

func DrawToolItem(screen *ebiten.Image, pos int) {
	item := objects.ToolbarItems[pos]

	x := float64(consts.ToolBarScale * int(pos))

	if item.OType.Image == nil {
		return
	} else {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

		op.GeoM.Reset()
		op.GeoM.Translate(x, 0)
		screen.DrawImage(toolBG, op)

		op.GeoM.Reset()
		iSize := item.OType.Image.Bounds()

		if item.OType.Rotatable && item.OType.Direction > 0 {
			x := float64(iSize.Size().X / 2)
			y := float64(iSize.Size().Y / 2)
			op.GeoM.Translate(-x, -y)
			op.GeoM.Rotate(consts.NinetyDeg * float64(item.OType.Direction))
			op.GeoM.Translate(x, y)
		}

		if item.OType.Image.Bounds().Max.X != consts.ToolBarScale {
			op.GeoM.Scale(1.0/(float64(iSize.Max.X)/consts.ToolBarScale), 1.0/(float64(iSize.Max.Y)/consts.ToolBarScale))
		}
		op.GeoM.Translate(x, 0)

		screen.DrawImage(item.OType.Image, op)
	}

	if item.SType == consts.ObjSubGame {
		if item.OType.TypeI == objects.SelectedItemType {
			ebitenutil.DrawRect(screen, consts.ToolBarOffsetX+float64(pos)*consts.ToolBarScale, consts.ToolBarOffsetY, consts.TBThick, consts.ToolBarScale, glob.ColorTBSelected)
			ebitenutil.DrawRect(screen, consts.ToolBarOffsetX+float64(pos)*consts.ToolBarScale, consts.ToolBarOffsetY, consts.ToolBarScale, consts.TBThick, glob.ColorTBSelected)

			ebitenutil.DrawRect(screen, consts.ToolBarOffsetX+float64(pos)*consts.ToolBarScale, consts.ToolBarOffsetY+consts.ToolBarScale-consts.TBThick, consts.ToolBarScale, consts.TBThick, glob.ColorTBSelected)
			ebitenutil.DrawRect(screen, consts.ToolBarOffsetX+(float64(pos)*consts.ToolBarScale)+consts.ToolBarScale-consts.TBThick, consts.ToolBarOffsetY, consts.TBThick, consts.ToolBarScale, glob.ColorTBSelected)
		}
	}
}
