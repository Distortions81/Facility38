package main

/* Figure out what option item user clicked */
func handleOptions(input XYs, window *windowData) bool {
	defer reportPanic("handleOptions")
	windowsLock.Lock()
	defer windowsLock.Unlock()

	originX := window.position.X
	originY := window.position.Y

	if !gMouseHeld {
		return false
	}

	for i, item := range settingItems {
		b := buttons[i]
		if PosWithinRect(
			XY{X: uint16(input.X - originX),
				Y: uint16(input.Y - originY)}, b, 1) {
			if (WASMMode && !item.WASMExclude) || !WASMMode {
				item.action(i)
				saveOptions()
				window.dirty = true
				gMouseHeld = false

				return true
			}
		}
	}

	return false
}
