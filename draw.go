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

func (g *Game) Draw(screen *ebiten.Image) {

	/* Init */
	if glob.DrewStartup {
		//Load map here eventually
		glob.DrewMap = true
		glob.StatusStr = ""
	} else {
		glob.DrewStartup = true
	}

	if !glob.DrewMap {
		glob.BootImage.Fill(glob.BootColor)
		str := "Loading..."
		tRect := text.BoundString(glob.BootFont, str)
		text.Draw(glob.BootImage, str, glob.BootFont, (glob.ScreenWidth/2)-int(tRect.Max.X/2), (glob.ScreenHeight/2)+int(tRect.Max.Y/2), glob.ColorWhite)
		screen.DrawImage(glob.BootImage, nil)
		glob.DrewStartup = true
		return
	}

	/* Draw start */
	screen.Fill(glob.BGColor)

	/* Adjust camerea position for zoom */
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

	/*Draw world*/
	for chunkPos, chunk := range glob.WorldMap {

		//Is this chunk on the screen?
		if chunkPos.X < chunkStartX || chunkPos.X > chunkEndX || chunkPos.Y < chunkStartY || chunkPos.Y > chunkEndY {
			chunkSkip++
			continue
		}
		chunkCount++

		for objPos, obj := range chunk.MObj {

			//Is this object on the screen?
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

			DrawObject(screen, objCamPosX, objCamPosY, obj)
		}
	}

	/* Draw overlays */
	for chunkPos, chunk := range glob.WorldMap {

		//Is this chunk on the screen?
		if chunkPos.X < chunkStartX || chunkPos.X > chunkEndX || chunkPos.Y < chunkStartY || chunkPos.Y > chunkEndY {
			continue
		}

		for objPos, obj := range chunk.MObj {

			//Is this object on the screen?
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
			if glob.ShowAltView || obj.TypeP.Key == consts.ObjTypeBasicBelt {

				var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
				op.GeoM.Reset()
				iSize := obj.TypeP.Image.Bounds()
				op.GeoM.Scale(((float64(obj.TypeP.Size.X))*glob.ZoomScale)/float64(iSize.Max.X), ((float64(obj.TypeP.Size.Y))*glob.ZoomScale)/float64(iSize.Max.Y))
				op.GeoM.Translate(objCamPosX, objCamPosY)

				if obj.TypeP.Key == consts.ObjTypeBasicBelt {

					/* Draw Output Materials */
					for _, m := range obj.OutputBuffer {
						if m == nil {
							continue
						}
						if m.Amount <= 0 {
							continue
						}
						img := m.TypeP.Image
						if img != nil {
							if m.TweenStamp.IsZero() {
								m.TweenStamp = time.Now()
							}
							move := time.Since(m.TweenStamp).Nanoseconds()
							amount := (float64(move) / float64(glob.RealUPS_ns))

							//Limit item movement, but go off end to smoothly transition between belts
							if obj.OutputObj != nil {
								if amount > 1 {
									amount = 1.5
								}
							} else {
								//If the belt is a dead end, stop before we go off
								if amount > consts.HBeltLimitEnd {
									amount = consts.HBeltLimitEnd
								}
							}

							op.GeoM.Translate(math.Round(amount*glob.ZoomScale), math.Round(consts.HBeltVertOffset*glob.ZoomScale))
							screen.DrawImage(img, op)

							//fmt.Println("Amount:", amount)
						} else {
							fmt.Println("Mat image not found.", m.TypeP.Name)
						}
					}

				} else {
					/* Draw contents */

					for _, c := range obj.Contains {
						if c == nil {
							continue
						}
						if c.Amount <= 0 {
							continue
						}
						img := c.TypeP.Image
						if img != nil {
							screen.DrawImage(img, op)
						} else {
							fmt.Println("Mat image not found.", c.TypeP.Name)
						}
					}
				}

				if obj.TypeP.HasOutput && obj.TypeP.Key != consts.ObjTypeBasicBelt {
					/* Draw Arrow */
					img := objects.ObjOverlayTypes[obj.OutputDir].Image
					if img != nil {
						screen.DrawImage(img, op)
					} else {
						fmt.Println("Arrow overlay image not found.")
					}
				}
			}
		}
	}

	//Get mouse position on world
	dtx := (glob.MousePosX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
	dty := (glob.MousePosY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))

	/* Draw debug info */
	if glob.StatusStr != "" {
		ebitenutil.DebugPrint(screen, glob.StatusStr)
	} else {
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("FPS: %.2f, IPS: %.2f, UPS: %.2f Zoom: %v Draw: c%v/o%v Skip: c%v/o%v (v%v-%v)",
				ebiten.CurrentFPS(), ebiten.CurrentTPS(), 1000000000.0/float64(glob.RealUPS_ns),
				glob.ZoomScale, chunkCount, objCount, chunkSkip, objSkip, consts.Version, consts.Build),
			0, glob.ScreenHeight-20)
	}

	/* Draw toolbar */
	for i := 0; i < objects.ToolbarMax; i++ {
		DrawToolItem(screen, i)
	}

	/* Toolbar tool tip */
	uipix := float64(objects.ToolbarMax * int(consts.TBSize))
	if glob.MousePosX <= uipix+consts.ToolBarOffsetX && glob.MousePosY <= consts.TBSize+consts.ToolBarOffsetY {
		temp := objects.ToolbarItems[int(glob.MousePosX/consts.TBSize)]
		item := temp.Link[temp.Key]

		toolTip := fmt.Sprintf("%v", item.Name)
		tRect := text.BoundString(glob.TipFont, toolTip)
		mx := glob.MousePosX + 20
		my := glob.MousePosY + 20
		ebitenutil.DrawRect(screen, mx-1, my-(float64(tRect.Dy()-1)), float64(tRect.Dx()+4), float64(tRect.Dy()+3), glob.ColorToolTipBG)
		text.Draw(screen, toolTip, glob.TipFont, int(mx), int(my), glob.ColorAqua)
	} else {
		/* Draw tool tip */
		pos := util.FloatXYToPosition(dtx, dty)
		chunk := util.GetChunk(&pos)

		toolTip := ""
		found := false
		if chunk != nil {
			o := chunk.MObj[pos]
			if o != nil {
				found = true
				toolTip = fmt.Sprintf("(%5.0f, %5.0f) %v", dtx, dty, o.TypeP.Name)
			}
		}

		if !found {
			toolTip = fmt.Sprintf("(%5.0f, %5.0f)", dtx, dty)
		}

		tRect := text.BoundString(glob.TipFont, toolTip)
		mx := glob.MousePosX + 20
		my := glob.MousePosY + 20
		ebitenutil.DrawRect(screen, mx-1, my-(float64(tRect.Dy()-1)), float64(tRect.Dx()+4), float64(tRect.Dy()+3), glob.ColorToolTipBG)
		text.Draw(screen, toolTip, glob.TipFont, int(mx), int(my), glob.ColorAqua)
	}
}

func DrawObject(screen *ebiten.Image, x float64, y float64, obj *glob.MObj) {

	/* Skip if not visible */
	if obj.TypeP.Key > consts.ObjTypeNone {

		/* Draw sprite */
		if obj.TypeP.Image == nil {
			fmt.Println("DrawObject: nil ebiten.*image encountered:", obj.TypeP.Name)
			return
		} else {
			var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
			op.GeoM.Reset()
			iSize := obj.TypeP.Image.Bounds()
			op.GeoM.Scale((float64(obj.TypeP.Size.X)*glob.ZoomScale)/float64(iSize.Max.X), (float64(obj.TypeP.Size.Y)*glob.ZoomScale)/float64(iSize.Max.Y))
			op.GeoM.Translate(x, y)
			if glob.ZoomScale < consts.SpriteScale {
				op.Filter = ebiten.FilterLinear
			}
			screen.DrawImage(obj.TypeP.Image, op)
		}

	} else {
		fmt.Println("DrawObject: empty object encountered.")
	}
}

func DrawToolItem(screen *ebiten.Image, pos int) {
	temp := objects.ToolbarItems[pos]
	item := temp.Link[temp.Key]

	x := float64(consts.TBSize * int(pos))

	if item.Image == nil {
		fmt.Println("DrawToolItem: nil ebiten.*image encountered:", item.Name)
		return
	} else {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
		op.GeoM.Reset()
		if item.Image.Bounds().Max.X != consts.TBSize {
			iSize := item.Image.Bounds()
			op.GeoM.Scale(1.0/(float64(iSize.Max.X)/consts.TBSize), 1.0/(float64(iSize.Max.Y)/consts.TBSize))
		}
		op.GeoM.Translate(x, 0)
		screen.DrawImage(item.Image, op)
	}

	if temp.Type == consts.ObjSubGame {
		if temp.Key == objects.SelectedItemType {
			ebitenutil.DrawRect(screen, consts.ToolBarOffsetX+float64(pos)*consts.TBSize, consts.ToolBarOffsetY, consts.TBThick, consts.TBSize, glob.ColorTBSelected)
			ebitenutil.DrawRect(screen, consts.ToolBarOffsetX+float64(pos)*consts.TBSize, consts.ToolBarOffsetY, consts.TBSize, consts.TBThick, glob.ColorTBSelected)

			ebitenutil.DrawRect(screen, consts.ToolBarOffsetX+float64(pos)*consts.TBSize, consts.ToolBarOffsetY+consts.TBSize-consts.TBThick, consts.TBSize, consts.TBThick, glob.ColorTBSelected)
			ebitenutil.DrawRect(screen, consts.ToolBarOffsetX+(float64(pos)*consts.TBSize)+consts.TBSize-consts.TBThick, consts.ToolBarOffsetY, consts.TBThick, consts.TBSize, glob.ColorTBSelected)
		}
	}
}
