package main

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/util"
	"fmt"

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
	startx := (1/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale)) / glob.DrawScale
	starty := (1/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale)) / glob.DrawScale
	endx := (float64(glob.ScreenWidth)/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale)) / glob.DrawScale
	endy := (float64(glob.ScreenHeight)/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale)) / glob.DrawScale

	sx = int(startx)
	sy = int(starty)
	ex = int(endx)
	ey = int(endy)

	/* Draw world */
	for ckey, chunk := range glob.WorldMap {
		//Is this chunk on the screen?
		if ckey.X < sx/glob.ChunkSize || ckey.X > ex/glob.ChunkSize || ckey.Y < sy/glob.ChunkSize || ckey.Y > ey/glob.ChunkSize {
			continue
		}
		for mkey, mobj := range chunk.MObj {
			//Is this obj on the screen?
			/*if mkey.X < sx || mkey.X > ex || mkey.Y < sy || mkey.Y > ey {
				continue
			}*/

			/* Item size, scaled */
			if glob.DrawScale >= 1.0 {
				xisize = mobj.Size - glob.ItemSpacing
				yisize = mobj.Size - glob.ItemSpacing
			}

			/* Draw scale */
			xs = xisize * glob.DrawScale
			ys = yisize * glob.DrawScale

			/* Item x/y, scaled */
			x = (float64(mkey.X) * glob.DrawScale)
			y = (float64(mkey.Y) * glob.DrawScale)

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

			DrawObject(screen, scrX, scrY, xss, yss, mobj, false)
		}
	}

	//Get mouse position on world
	dtx := (glob.MousePosX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
	dty := (glob.MousePosY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))
	//Adjust for draw scale
	gwx := (dtx / glob.DrawScale)
	gwy := (dty / glob.DrawScale)

	/* Draw debug info */
	if glob.StatusStr != "" {
		ebitenutil.DebugPrint(screen, glob.StatusStr)
	} else {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("v%v-%v, %vfps, z: %v", consts.Version, consts.Build, int(ebiten.CurrentFPS()), glob.ZoomScale), 0, glob.ScreenHeight-20)
	}

	/* Draw toolbar */
	for i := 0; i < glob.ObjTypeMax; i++ {
		DrawObject(screen, glob.TBSize*float64(i), 0, glob.TBSize, glob.TBSize, glob.MObj{Type: i + 1}, true)
		//Draw item selected
		if i == glob.SelectedItemType-1 {
			ebitenutil.DrawRect(screen, float64(i)*glob.TBSize, 0, glob.TBThick, glob.TBSize, glob.ColorWhite)
			ebitenutil.DrawRect(screen, float64(i)*glob.TBSize, 0, glob.TBSize, glob.TBThick, glob.ColorWhite)

			ebitenutil.DrawRect(screen, float64(i)*glob.TBSize, glob.TBSize-glob.TBThick, glob.TBSize, glob.TBThick, glob.ColorWhite)
			ebitenutil.DrawRect(screen, (float64(i)*glob.TBSize)+glob.TBSize-glob.TBThick, 0, glob.TBThick, glob.TBSize, glob.ColorWhite)
		}
	}

	/* Toolbar tool tip */
	if glob.MousePosX <= float64(glob.ObjTypeMax)*glob.TBSize && glob.MousePosY <= glob.TBSize {
		toolTip := fmt.Sprintf("%v", glob.ObjTypes[int(glob.MousePosX/glob.TBSize)+1].Name)
		tRect := text.BoundString(glob.TipFont, toolTip)
		mx := glob.MousePosX + 20
		my := glob.MousePosY + 20
		ebitenutil.DrawRect(screen, mx-1, my-(float64(tRect.Dy()-1)), float64(tRect.Dx()+4), float64(tRect.Dy()+3), glob.ColorToolTipBG)
		text.Draw(screen, toolTip, glob.TipFont, int(mx), int(my), glob.ColorAqua)
	} else {
		/* Draw tool tip */
		pos := util.FloatXYToPosition(gwx, gwy)
		chunk := util.GetChunk(pos)
		obj := chunk.MObj[pos]

		toolTip := ""
		if obj.Type != 0 {
			toolTip = fmt.Sprintf("%v (%5.0f, %5.0f)", glob.ObjTypes[obj.Type].Name, gwx, gwy)
		} else {
			toolTip = fmt.Sprintf("(%5.0f, %5.0f)", gwx, gwy)
		}
		tRect := text.BoundString(glob.TipFont, toolTip)
		mx := glob.MousePosX + 20
		my := glob.MousePosY + 20
		ebitenutil.DrawRect(screen, mx-1, my-(float64(tRect.Dy()-1)), float64(tRect.Dx()+4), float64(tRect.Dy()+3), glob.ColorToolTipBG)
		text.Draw(screen, toolTip, glob.TipFont, int(mx), int(my), glob.ColorAqua)
	}
}

func DrawObject(screen *ebiten.Image, x float64, y float64, xs float64, ys float64, mobj glob.MObj, isUI bool) {

	var zoom float64 = glob.ZoomScale

	if isUI {
		zoom = ((xs + ys) / 2.0) / 2.6
	}

	/* Skip if not visible */
	if mobj.Type != glob.ObjTypeNone {
		typeData := glob.ObjTypes[mobj.Type]

		/* Draw rect */
		ebitenutil.DrawRect(screen, x, y, xs, ys, typeData.ItemColor)

		/* Symbols */
		if zoom > 3 {
			tRect := text.BoundString(glob.ItemFont, typeData.Symbol)
			opt := &ebiten.DrawImageOptions{}
			opt.GeoM.Scale(zoom/glob.FontScale, zoom/glob.FontScale)
			opt.GeoM.Translate(x, y+(((float64(tRect.Dy())*1.1)/glob.FontScale)*zoom))
			text.DrawWithOptions(screen, typeData.Symbol, glob.ItemFont, opt)
		}
	}
}
