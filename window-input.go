package main

import (
	"Facility38/util"
	"Facility38/world"
)

func handleOptions(input world.XYs, window *WindowData) bool {
	defer util.ReportPanic("handleOptions")
	WindowsLock.Lock()
	defer WindowsLock.Unlock()

	originX := window.Position.X
	originY := window.Position.Y

	if !gMouseHeld {
		return false
	}

	for i, item := range settingItems {
		b := buttons[i]
		if util.PosWithinRect(
			world.XY{X: uint16(input.X - originX),
				Y: uint16(input.Y - originY)}, b, 1) {
			if (world.WASMMode && !item.WASMExclude) || !world.WASMMode {
				item.Action(i)
				saveOptions()
				window.Dirty = true
				gMouseHeld = false

				return true
			}
		}
	}

	return false
}
