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
	sx = int((1/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale)) / glob.DrawScale)
	sy = int((1/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale)) / glob.DrawScale)
	ex = int((float64(glob.ScreenWidth)/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale)) / glob.DrawScale)
	ey = int((float64(glob.ScreenHeight)/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale)) / glob.DrawScale)

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

			//Item spacing
			if glob.DrawScale >= 1.0 {
				xisize = float64(glob.GameObjTypes[mobj.Type].Size.X) - glob.ItemSpacing
				yisize = float64(glob.GameObjTypes[mobj.Type].Size.Y) - glob.ItemSpacing
			}

			//Item size, scaled
			xs = xisize * glob.DrawScale
			ys = yisize * glob.DrawScale

			//Item size, scaled
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

			DrawObject(screen, scrX, scrY, xs, ys, mobj.Type, glob.ObjSubGame, false)
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
	for i := 1; i <= glob.UITypeMax; i++ {
		DrawObject(screen, glob.ToolBarOffsetX+glob.TBSize*float64(i-1), glob.ToolBarOffsetY, glob.TBSize, glob.TBSize, i, glob.ObjSubUI, true)
	}
	spos := (glob.TBSize * float64(glob.UITypeMax))
	for i := 1; i <= glob.GameTypeMax; i++ {
		DrawObject(screen, spos+glob.ToolBarOffsetX+glob.TBSize*float64(i-1), glob.ToolBarOffsetY, glob.TBSize, glob.TBSize, i, glob.ObjSubGame, true)
		//Draw item selected
		if i == glob.SelectedItemType && glob.GameObjTypes[i].SubType == glob.ObjSubGame {
			ebitenutil.DrawRect(screen, spos+glob.ToolBarOffsetX+float64(i-1)*glob.TBSize, glob.ToolBarOffsetY, glob.TBThick, glob.TBSize, glob.ColorTBSelected)
			ebitenutil.DrawRect(screen, spos+glob.ToolBarOffsetX+float64(i-1)*glob.TBSize, glob.ToolBarOffsetY, glob.TBSize, glob.TBThick, glob.ColorTBSelected)

			ebitenutil.DrawRect(screen, spos+glob.ToolBarOffsetX+float64(i-1)*glob.TBSize, glob.ToolBarOffsetY+glob.TBSize-glob.TBThick, glob.TBSize, glob.TBThick, glob.ColorTBSelected)
			ebitenutil.DrawRect(screen, spos+glob.ToolBarOffsetX+(float64(i-1)*glob.TBSize)+glob.TBSize-glob.TBThick, glob.ToolBarOffsetY, glob.TBThick, glob.TBSize, glob.ColorTBSelected)
		}
	}

	/* Toolbar tool tip */
	if glob.MousePosX <= float64(glob.GameTypeMax)*glob.TBSize && glob.MousePosY <= glob.TBSize {
		toolTip := fmt.Sprintf("%v", glob.GameObjTypes[int(glob.MousePosX/glob.TBSize)+1].Name)
		tRect := text.BoundString(glob.TipFont, toolTip)
		mx := glob.MousePosX + 20
		my := glob.MousePosY + 20
		ebitenutil.DrawRect(screen, mx-1, my-(float64(tRect.Dy()-1)), float64(tRect.Dx()+4), float64(tRect.Dy()+3), glob.ColorToolTipBG)
		text.Draw(screen, toolTip, glob.TipFont, int(mx), int(my), glob.ColorAqua)
	} else {
		/* Draw tool tip */
		pos := util.FloatXYToPosition(gwx, gwy)
		chunk := util.GetChunk(pos)
		if chunk != nil {
			obj := chunk.MObj[pos]
			if obj != nil {
				toolTip := ""
				if obj.Type != 0 {
					toolTip = fmt.Sprintf("%v (%5.0f, %5.0f)", glob.GameObjTypes[obj.Type].Name, gwx, gwy)
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
	}
}

func DrawObject(screen *ebiten.Image, x float64, y float64, xs float64, ys float64, objType int, subType int, isUI bool) {

	var zoom float64 = glob.ZoomScale

	if isUI {
		zoom = 1
	}

	/* Skip if not visible */
	if objType > glob.ObjTypeNone {
		temp := glob.SubTypes[subType]
		typeData := temp[objType]

		/* Draw rect */
		/* Symbols */
		if typeData.Image == nil {
			fmt.Println("DrawObject: nil ebten.*image eencountered.")
			return
		} else {
			var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
			op.GeoM.Reset()
			if !isUI {
				op.GeoM.Scale(zoom/glob.SpriteScale, zoom/glob.SpriteScale)
			}
			op.GeoM.Translate(x, y)
			if !isUI && zoom < 64 {
				op.Filter = ebiten.FilterLinear
			}
			screen.DrawImage(typeData.Image, op)
		}
	}
}
