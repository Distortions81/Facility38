package main

import (
	"time"
)

/* This file is for toolbar actions */

/* Toggle settings window */
func settingsToggle() {
	defer reportPanic("settingsToggle")
	if windows[0].active {
		closeWindow(windows[0])
	} else {
		openWindow(windows[0])
	}
}

/* Toggle help windw */
func toggleHelp() {
	defer reportPanic("toggleHelp")

	if windows[1].active {
		closeWindow(windows[1])
	} else {
		openWindow(windows[1])
	}
}

/* Toggle info overlay */
func toggleOverlay() {
	defer reportPanic("toggleOverlay")
	if OverlayMode {
		OverlayMode = false
		ChatDetailed("Info overlay is now off.", ColorOrange, time.Second*5)
	} else {
		OverlayMode = true
		ChatDetailed("Info overlay is now on.", ColorOrange, time.Second*5)
	}
}

/* Switch between normal and resource layers */
func switchGameLayer() {
	defer reportPanic("switchGameLayer")
	ShowResourceLayerLock.Lock()

	if ShowResourceLayer {
		ShowResourceLayer = false
		ChatDetailed("Switched from resource layer to game.", ColorOrange, time.Second*10)
	} else {
		ShowResourceLayer = true
		ChatDetailed("Switched from game to resource layer.", ColorOrange, time.Second*10)
	}
	for _, sChunk := range SuperChunkList {
		sChunk.pixmapDirty = true
	}
	ShowResourceLayerLock.Unlock()
}
