package action

import (
	"GameTest/gv"
)

func SwitchLayer() {
	switch gv.CurrentLayer {
	case gv.LayerNormal:
		gv.CurrentLayer = gv.LayerMineral
	default:
		gv.CurrentLayer = gv.LayerNormal
	}
}
