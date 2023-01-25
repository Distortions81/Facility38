package main

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/objects"
	"GameTest/util"
	"fmt"
	"math"
	"time"

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
	//screen.Fill(glob.ColorRed)

	/* Adjust cam position for zoom */
	camXPos := float64(-glob.CameraX) + (float64(glob.ScreenWidth/2) / glob.ZoomScale)
	camYPos := float64(-glob.CameraY) + (float64(glob.ScreenHeight/2) / glob.ZoomScale)

	/* Get camera bounds */
	camStartX := int((1/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale)))
	camStarty := int((1/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale)))
	camEndX := int((float64(glob.ScreenWidth)/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale)))
	camEndY := int((float64(glob.ScreenHeight)/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale)))

	/* Draw stats */
	chunkSkip := 0
	objSkip := 0
	chunkCount := 0
	objCount := 0

	/* Pre-calc camera chunk position */
	chunkStartX := camStartX / consts.ChunkSize
	chunkStartY := camStarty / consts.ChunkSize
	chunkEndX := camEndX / consts.ChunkSize
	chunkEndY := camEndY / consts.ChunkSize

	drawTerrain(screen, camXPos, camYPos, camStartX, camStarty, camEndX, camEndY)

	/* Draw world */
	glob.WorldMapLock.Lock()
	for chunkPos, chunk := range glob.WorldMap {

		/* Is this chunk on the screen? */
		if chunkPos.X < chunkStartX || chunkPos.X > chunkEndX || chunkPos.Y < chunkStartY || chunkPos.Y > chunkEndY {
			chunkSkip++
			continue
		}
		chunkCount++

		/* Draw objects in chunk */
		for objPos, obj := range chunk.WObject {

			/* Is this object on the screen? */
			if objPos.X < camStartX || objPos.X > camEndX || objPos.Y < camStarty || objPos.Y > camEndY {
				objSkip++
				continue
			}
			objCount++

			/* camera + object */
			objOffX := camXPos + (float64(objPos.X))
			objOffY := camYPos + (float64(objPos.Y))

			/* camera zoom */
			objCamPosX := objOffX * glob.ZoomScale
			objCamPosY := objOffY * glob.ZoomScale

			/* Time to draw it */
			DrawObject(screen, objCamPosX, objCamPosY, obj)
		}
	}
	glob.WorldMapLock.Unlock()

	/* Draw overlays */
	glob.WorldMapLock.Lock()
	for chunkPos, chunk := range glob.WorldMap {

		/* Is this chunk on the screen? */
		if chunkPos.X < chunkStartX || chunkPos.X > chunkEndX || chunkPos.Y < chunkStartY || chunkPos.Y > chunkEndY {
			continue
		}

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

			/* Overlays */
			/* Draw belt overlays */
			if obj.TypeP.TypeI == consts.ObjTypeBasicBelt ||
				obj.TypeP.TypeI == consts.ObjTypeBasicBeltVert {

				var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
				op.GeoM.Reset()
				iSize := obj.TypeP.Bounds
				op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X), ((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
				op.GeoM.Translate(objCamPosX, objCamPosY)

				/* Draw Input Materials */
				if obj.OutputBuffer.Amount > 0 {
					matTween(obj.OutputBuffer, obj, op, screen)
				} else {
					for _, m := range obj.InputBuffer {
						if m.Amount > 0 {
							matTween(m, obj, op, screen)
							break
						}
					}
				}

			} else if glob.ShowInfoLayer {
				/* Info Overlays, such as arrows and blocked indicator */
				var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
				op.GeoM.Reset()
				iSize := obj.TypeP.Bounds
				op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X), ((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
				op.GeoM.Translate(objCamPosX, objCamPosY)

				if obj.TypeP.HasMatOutput {
					op.GeoM.Reset()
					iSize := obj.TypeP.Bounds
					op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X), ((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
					op.GeoM.Translate(objCamPosX, objCamPosY)
					/* Draw Arrow */
					img := objects.ObjOverlayTypes[obj.OutputDir].Image
					if img != nil {
						screen.DrawImage(img, op)
					}
					/* Show blocked outputs */
					img = objects.ObjOverlayTypes[consts.ObjTypeBlocked].Image
					if obj.OutputObj != nil && obj.OutputBuffer.Amount > 0 &&
						obj.OutputObj.InputBuffer[obj] != nil && obj.OutputObj.InputBuffer[obj].Amount > 0 {

						var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
						op.GeoM.Reset()

						iSize := obj.TypeP.Bounds
						op.GeoM.Translate(float64(iSize.Max.X)-float64(objects.ObjOverlayTypes[consts.ObjTypeBlocked].Bounds.Max.X)-consts.BlockedIndicatorOffset, consts.BlockedIndicatorOffset)
						op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X), ((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
						op.GeoM.Translate(objCamPosX, objCamPosY)
						screen.DrawImage(img, op)
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
	if glob.StatusStr != "" {
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("FPS: %.2f, IPS: %.2f, UPS: %.2f, TockPerSec: %.2fm  (v%v-%v)",
				ebiten.ActualFPS(), ebiten.ActualTPS(), 1000000000.0/float64(glob.MeasuredObjectUPS_ns),
				(float64(objects.TockCount)*(1000000000.0/float64(glob.MeasuredObjectUPS_ns)))/1000000,
				consts.Version, consts.Build),
			0, glob.ScreenHeight-20)
	}

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
				toolTip = fmt.Sprintf("(%5.0f, %5.0f) %v", math.Floor(worldMouseX-consts.XYCenter), math.Floor(worldMouseY-consts.XYCenter), o.TypeP.Name)

				found := false
				for _, t := range o.Contents {
					if t != nil && t.Amount > 0 {
						found = true
						toolTip += fmt.Sprintf(" Contents: %v: %v", t.TypeP.Name, t.Amount)
					}
				}
				if !found {
					for _, t := range o.InputBuffer {
						if t != nil && t.Amount > 0 {
							toolTip += fmt.Sprintf(" Contents: %v: %v", t.TypeP.Name, t.Amount)
						}
					}
				}

			}
		}

		/* No object contents found */
		if !found {
			toolTip = fmt.Sprintf("(%5.0f, %5.0f)", math.Floor(worldMouseX-consts.XYCenter), math.Floor(worldMouseY-consts.XYCenter))
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

/* Prototype, needs optimization */
func drawTerrain(screen *ebiten.Image, camXPos float64, camYPos float64, camStartX int, camStartY int, camEndX int, camEndY int) {

	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM.Reset()
	img := objects.TerrainTypes[0].Image
	iSize := objects.TerrainTypes[0].Bounds
	oSize := objects.TerrainTypes[0].Size

	for j := 0; j < 1000; j += oSize.X {
		for i := 0; i < 1000; i += oSize.Y {
			pos := glob.Position{X: int(float64((consts.XYCenter) - 500.0 + float64(i))),
				Y: int(float64((consts.XYCenter) - 500.0 + float64(j)))}
			if pos.X+oSize.X < camStartX || pos.X-oSize.X*2 > camEndX || pos.Y+oSize.Y*2 < camStartY || pos.Y-oSize.Y*2 > camEndY {
				continue
			}
			op.GeoM.Scale(float64(oSize.X)*glob.ZoomScale/float64(iSize.Max.X-1), float64(oSize.Y)*glob.ZoomScale/float64(iSize.Max.X-1))
			op.GeoM.Translate((float64(pos.X)+camXPos)*glob.ZoomScale, (float64(pos.Y)+camYPos)*glob.ZoomScale)

			screen.DrawImage(img, op)
			op.GeoM.Reset()

		}
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

func DrawObject(screen *ebiten.Image, x float64, y float64, obj *glob.WObject) {

	/* Draw sprite */
	if obj.TypeP.Image == nil {
		return
	} else {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
		op.GeoM.Reset()
		iSize := obj.TypeP.Bounds
		op.GeoM.Scale((float64(obj.TypeP.Size.X)*glob.ZoomScale)/float64(iSize.Max.X), (float64(obj.TypeP.Size.Y)*glob.ZoomScale)/float64(iSize.Max.Y))

		op.GeoM.Translate(math.Floor(x), math.Floor(y))
		screen.DrawImage(obj.TypeP.Image, op)
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
		if item.OType.Bounds.Max.X != consts.ToolBarScale {
			iSize := item.OType.Bounds
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
