package main

import "time"

/* Figure out what option item user clicked */
func handleOptions(input XYs, window *windowData) bool {
	defer reportPanic("handleOptions")
	windowsLock.Lock()
	defer windowsLock.Unlock()

	if !gMouseHeld {
		return false
	}

	originX := window.position.X
	originY := window.position.Y

	for i, item := range settingItems {
		b := optionWindowButtons[i]
		if PosWithinRect(
			XY{X: uint16(input.X - originX),
				Y: uint16(input.Y - originY)}, b, 1) {
			if (wasmMode && !item.WASMExclude) || !wasmMode {
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

func handleHelpWindow(input XYs, window *windowData) bool {
	defer reportPanic("handleHelpWindow")
	windowsLock.Lock()
	defer windowsLock.Unlock()

	if !gMouseHeld {
		return false
	}

	originX := window.position.X
	originY := window.position.Y

	for i := range updateWindowButtons {
		b := updateWindowButtons[i]
		if PosWithinRect(
			XY{X: uint16(input.X - originX),
				Y: uint16(input.Y - originY)}, b, 1) {

			if !updatingGame.Load() {
				gMouseHeld = false
				updatingGame.Store(true)
				window.dirty = true

				go func() {
					time.Sleep(time.Second)
					downloadBuild()
				}()
			}
			return true
		}
	}

	return false
}
