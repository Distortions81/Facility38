package main

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/obj"
	"GameTest/util"
	"fmt"
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
	var x, y, xs, ys, xisize, yisize float64
	var sx, sy, ex, ey int

	/* Get the camera position */
	mainx := float64(-glob.CameraX) + (float64(glob.ScreenWidth/2) / glob.ZoomScale)
	mainy := float64(-glob.CameraY) + (float64(glob.ScreenHeight/2) / glob.ZoomScale)

	/* Calculate screen on world */
	sx = int((1/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale)) / consts.DrawScale)
	sy = int((1/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale)) / consts.DrawScale)
	ex = int((float64(glob.ScreenWidth)/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale)) / consts.DrawScale)
	ey = int((float64(glob.ScreenHeight)/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale)) / consts.DrawScale)

	/* Draw world */
	for ckey, chunk := range glob.WorldMap {

		//Is this chunk on the screen?
		if ckey.X < sx/consts.ChunkSize || ckey.X > ex/consts.ChunkSize || ckey.Y < sy/consts.ChunkSize || ckey.Y > ey/consts.ChunkSize {
			continue
		}
		for mkey, mobj := range chunk.MObj {

			//Item spacing
			if consts.DrawScale >= 1.0 {
				xisize = float64(mobj.TypeP.Size.X) - consts.ItemSpacing
				yisize = float64(mobj.TypeP.Size.Y) - consts.ItemSpacing
			}

			//Item size, scaled
			xs = xisize * consts.DrawScale
			ys = yisize * consts.DrawScale

			//Item size, scaled
			x = (float64(mkey.X) * consts.DrawScale)
			y = (float64(mkey.Y) * consts.DrawScale)

			/* camera + object */
			newx := mainx + (x)
			newy := mainy + (y)

			/* camera zoom */
			scrX := newx * glob.ZoomScale
			scrY := newy * glob.ZoomScale

			/* item magnification */
			xss := xs * glob.ZoomScale
			yss := ys * glob.ZoomScale

			/* Helps far zoom */
			if xss < 1 {
				xss = 1
			}
			if yss < 1 {
				yss = 1
			}

			DrawObject(screen, scrX, scrY, xs, ys, mobj)

		}
	}

	//Get mouse position on world
	dtx := (glob.MousePosX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
	dty := (glob.MousePosY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))
	//Adjust for draw scale
	gwx := (dtx / consts.DrawScale)
	gwy := (dty / consts.DrawScale)

	/* Draw debug info */
	if glob.StatusStr != "" {
		ebitenutil.DebugPrint(screen, glob.StatusStr)
	} else {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("v%v-%v, %vfps, z: %v, wt/rt: %v/%v", consts.Version, consts.Build, int(ebiten.CurrentFPS()), glob.ZoomScale, time.Duration(float64(time.Since(obj.WorldEpoch))*consts.TIMESCALE).Round(time.Second).String(), time.Duration(time.Since(obj.WorldEpoch)).Round(time.Second).String()), 0, glob.ScreenHeight-20)
	}

	/* Draw toolbar */
	for i := 0; i < obj.ToolbarMax; i++ {
		DrawToolItem(screen, i)
	}

	/* Toolbar tool tip */
	uipix := float64(obj.ToolbarMax * int(consts.TBSize))
	if glob.MousePosX <= uipix+consts.ToolBarOffsetX && glob.MousePosY <= consts.TBSize+consts.ToolBarOffsetY {
		temp := obj.ToolbarItems[int(glob.MousePosX/consts.TBSize)]
		item := temp.Link[temp.Key]

		toolTip := fmt.Sprintf("%v", item.Name)
		tRect := text.BoundString(glob.TipFont, toolTip)
		mx := glob.MousePosX + 20
		my := glob.MousePosY + 20
		ebitenutil.DrawRect(screen, mx-1, my-(float64(tRect.Dy()-1)), float64(tRect.Dx()+4), float64(tRect.Dy()+3), glob.ColorToolTipBG)
		text.Draw(screen, toolTip, glob.TipFont, int(mx), int(my), glob.ColorAqua)
	} else {
		/* Draw tool tip */
		pos := util.FloatXYToPosition(gwx, gwy)
		glob.WorldMapLock.RLock()
		chunk := util.GetChunk(pos)
		glob.WorldMapLock.RUnlock()

		toolTip := ""
		found := false
		if chunk != nil {
			o := chunk.MObj[pos]
			if o != nil {
				found = true
				toolTip = fmt.Sprintf("(%5.0f, %5.0f) %v", gwx, gwy, o.TypeP.Name)
				for _, c := range o.Contents {
					if c.Amount == 0 {
						continue
					}
					if c.Amount > 0 {
						toolTip += fmt.Sprintf(" (%vkg %v)", c.Amount, c.TypeP.Name)
					}
				}
			}
		}

		if !found {
			toolTip = fmt.Sprintf("(%5.0f, %5.0f)", gwx, gwy)
		}

		tRect := text.BoundString(glob.TipFont, toolTip)
		mx := glob.MousePosX + 20
		my := glob.MousePosY + 20
		ebitenutil.DrawRect(screen, mx-1, my-(float64(tRect.Dy()-1)), float64(tRect.Dx()+4), float64(tRect.Dy()+3), glob.ColorToolTipBG)
		text.Draw(screen, toolTip, glob.TipFont, int(mx), int(my), glob.ColorAqua)
	}
}

func DrawObject(screen *ebiten.Image, x float64, y float64, xs float64, ys float64, o *glob.MObj) {

	var zoom float64 = glob.ZoomScale

	/* Skip if not visible */
	if o.Type > consts.ObjTypeNone {

		/* Draw sprite */
		if o.TypeP.Image == nil {
			fmt.Println("DrawObject: nil ebiten.*image encountered:", o.TypeP.Name)
			return
		} else {
			var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
			op.GeoM.Reset()
			iSize := o.TypeP.Image.Bounds()
			op.GeoM.Scale(((xs)*zoom)/float64(iSize.Max.X), ((ys)*zoom)/float64(iSize.Max.Y))
			op.GeoM.Translate(x, y)
			if zoom < consts.SpriteScale {
				op.Filter = ebiten.FilterLinear
			}
			screen.DrawImage(o.TypeP.Image, op)

			if glob.ShowAltView && o.TypeP.HasOutput {
				img := obj.ObjOverlayTypes[o.OutputDir].Image
				if img != nil {
					screen.DrawImage(img, op)
				} else {
					fmt.Println("Arrow overlay image not found.")
				}
			}
		}
	} else {
		fmt.Println("DrawObject: empty object encountered.")
	}
}

func DrawToolItem(screen *ebiten.Image, pos int) {
	temp := obj.ToolbarItems[pos]
	item := temp.Link[temp.Key]

	x := float64(consts.TBSize * pos)

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
		if temp.Key == obj.SelectedItemType {
			ebitenutil.DrawRect(screen, consts.ToolBarOffsetX+float64(pos)*consts.TBSize, consts.ToolBarOffsetY, consts.TBThick, consts.TBSize, glob.ColorTBSelected)
			ebitenutil.DrawRect(screen, consts.ToolBarOffsetX+float64(pos)*consts.TBSize, consts.ToolBarOffsetY, consts.TBSize, consts.TBThick, glob.ColorTBSelected)

			ebitenutil.DrawRect(screen, consts.ToolBarOffsetX+float64(pos)*consts.TBSize, consts.ToolBarOffsetY+consts.TBSize-consts.TBThick, consts.TBSize, consts.TBThick, glob.ColorTBSelected)
			ebitenutil.DrawRect(screen, consts.ToolBarOffsetX+(float64(pos)*consts.TBSize)+consts.TBSize-consts.TBThick, consts.ToolBarOffsetY, consts.TBThick, consts.TBSize, glob.ColorTBSelected)
		}
	}
}
