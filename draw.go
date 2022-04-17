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
	var sx, sy, ex, ey int

	/* Get the camera position */
	mainx := float64(-glob.CameraX) + (float64(glob.ScreenWidth/2) / glob.ZoomScale)
	mainy := float64(-glob.CameraY) + (float64(glob.ScreenHeight/2) / glob.ZoomScale)

	/* Calculate screen on world */
	sx = int((1/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale)))
	sy = int((1/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale)))
	ex = int((float64(glob.ScreenWidth)/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale)))
	ey = int((float64(glob.ScreenHeight)/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale)))

	/* Draw world */
	cSkip := 0
	oSkip := 0

	cCount := 0
	oCount := 0

	csx := sx / consts.ChunkSize
	csy := sy / consts.ChunkSize
	cex := ex / consts.ChunkSize
	cey := ey / consts.ChunkSize

	for ckey, chunk := range glob.WorldMap {

		//Is this chunk on the screen?
		if ckey.X < csx || ckey.X > cex || ckey.Y < csy || ckey.Y > cey {
			cSkip++
			continue
		}
		cCount++

		for mkey, mobj := range chunk.MObj {

			//Is this object on the screen?
			if mkey.X < sx || mkey.X > ex || mkey.Y < sy || mkey.Y > ey {
				oSkip++
				continue
			}

			oCount++

			/* camera + object */
			newx := mainx + (float64(mkey.X))
			newy := mainy + (float64(mkey.Y))

			/* camera zoom */
			scrX := newx * glob.ZoomScale
			scrY := newy * glob.ZoomScale

			DrawObject(screen, scrX, scrY, float64(mobj.TypeP.Size.X), float64(mobj.TypeP.Size.Y), mobj)

		}
	}

	//Draw overlays, we should queue these in a list (if needed)
	//This is a quick test fix
	for ckey, chunk := range glob.WorldMap {

		//Is this chunk on the screen?
		if ckey.X < csx || ckey.X > cex || ckey.Y < csy || ckey.Y > cey {
			cSkip++
			continue
		}
		cCount++

		for mkey, o := range chunk.MObj {

			//Is this object on the screen?
			if mkey.X < sx || mkey.X > ex || mkey.Y < sy || mkey.Y > ey {
				oSkip++
				continue
			}

			if glob.ShowAltView || o.TypeP.Key == consts.ObjTypeBasicBelt {

				xs := float64(o.TypeP.Size.X)
				ys := float64(o.TypeP.Size.Y)

				/* camera + object */
				newx := mainx + (float64(mkey.X))
				newy := mainy + (float64(mkey.Y))

				/* camera zoom */
				scrX := newx * glob.ZoomScale
				scrY := newy * glob.ZoomScale

				var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
				op.GeoM.Reset()
				iSize := o.TypeP.Image.Bounds()
				op.GeoM.Scale(((xs)*glob.ZoomScale)/float64(iSize.Max.X), ((ys)*glob.ZoomScale)/float64(iSize.Max.Y))
				op.GeoM.Translate(scrX, scrY)

				if o.TypeP.Key == consts.ObjTypeBasicBelt {
					/* Draw Ext */

					for _, m := range o.OutputBuffer {
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
							if o.OutputObj != nil {
								if amount > 1 {
									amount = 1
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

					for _, c := range o.Contains {
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

				if o.TypeP.HasOutput && o.TypeP.Key != consts.ObjTypeBasicBelt {
					/* Draw Arrow */
					img := objects.ObjOverlayTypes[o.OutputDir].Image
					if img != nil {
						screen.DrawImage(img, op)
					} else {
						fmt.Println("Arrow overlay image not found.")
					}
				}
			}

		}
	}

	//fmt.Println("Chunks skipped:", cSkip, "Objects skipped:", oSkip)
	//fmt.Println("Chunks drawn:", cCount, "Objects drawn:", oCount)

	//Get mouse position on world
	dtx := (glob.MousePosX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
	dty := (glob.MousePosY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))

	/* Draw debug info */
	if glob.StatusStr != "" {
		ebitenutil.DebugPrint(screen, glob.StatusStr)
	} else {
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("FPS: %.2f, IPS: %.2f, UPS: %.2f Zoom: %v (v%v-%v)",
				ebiten.CurrentFPS(), ebiten.CurrentTPS(), 1000000000.0/float64(glob.RealUPS_ns), glob.ZoomScale, consts.Version, consts.Build),
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

				for _, c := range o.Contains {
					if c == nil {
						continue
					}
					if c.Amount == 0 {
						continue
					}
					if c.Amount > 0 {
						toolTip += fmt.Sprintf(" (%vkg %v cont) %v", c.Amount, c.TypeP.Name, c.TypeP.Name)
					}
				}

				for _, m := range o.OutputBuffer {

					if m == nil {
						continue
					}
					if m.Amount == 0 {
						continue
					}
					if m.Amount > 0 {
						toolTip += fmt.Sprintf(" (%vkg %v out) %v", m.Amount, m.TypeP.Name, m.TypeP.Name)
					}

				}

				for _, c := range o.InputBuffer {
					for _, m := range c {
						if m == nil {
							continue
						}
						if m.Amount == 0 {
							continue
						}
						if m.Amount > 0 {
							toolTip += fmt.Sprintf(" (%vkg %v in) %v", m.Amount, m.TypeP.Name, m.TypeP.Name)
						}
					}
				}

				if o.OutputObj != nil {
					toolTip += fmt.Sprintf(" Output To: %v (%v)", o.OutputObj.TypeP.Name, util.DirToName(o.OutputDir))
				}

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

func DrawObject(screen *ebiten.Image, x float64, y float64, xs float64, ys float64, o *glob.MObj) {

	/* Skip if not visible */
	if o.TypeP.Key > consts.ObjTypeNone {

		/* Draw sprite */
		if o.TypeP.Image == nil {
			fmt.Println("DrawObject: nil ebiten.*image encountered:", o.TypeP.Name)
			return
		} else {
			var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
			op.GeoM.Reset()
			iSize := o.TypeP.Image.Bounds()
			op.GeoM.Scale(((xs)*glob.ZoomScale)/float64(iSize.Max.X), ((ys)*glob.ZoomScale)/float64(iSize.Max.Y))
			op.GeoM.Translate(x, y)
			if glob.ZoomScale < consts.SpriteScale {
				op.Filter = ebiten.FilterLinear
			}
			screen.DrawImage(o.TypeP.Image, op)
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
