package main

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/objects"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func drawToolItems(screen *ebiten.Image) {

	for pos := 0; pos < objects.ToolbarMax; pos++ {
		item := objects.ToolbarItems[pos]

		x := float64(consts.ToolBarScale * int(pos))

		if item.OType.Image == nil {
			return
		} else {
			var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

			op.GeoM.Reset()
			op.GeoM.Translate(x, 0)
			screen.DrawImage(glob.ToolBG, op)

			op.GeoM.Reset()
			iSize := item.OType.Image.Bounds()

			if item.OType.Rotatable && item.OType.Direction > 0 {
				x := float64(iSize.Size().X / 2)
				y := float64(iSize.Size().Y / 2)
				op.GeoM.Translate(-x, -y)
				op.GeoM.Rotate(cNinetyDeg * float64(item.OType.Direction))
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
}
