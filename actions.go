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

/* Toggle help window */
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
	if overlayMode {
		overlayMode = false
		chatDetailed("Info overlay is now off.", ColorOrange, time.Second*5)
	} else {
		overlayMode = true
		chatDetailed("Info overlay is now on.", ColorOrange, time.Second*5)
	}
}

/* Switch between normal and resource layers */
func switchGameLayer() {
	defer reportPanic("switchGameLayer")
	showResourceLayerLock.Lock()

	if showResourceLayer {
		showResourceLayer = false
		chatDetailed("Switched from resource layer to game.", ColorOrange, time.Second*10)
	} else {
		showResourceLayer = true
		chatDetailed("Switched from game to resource layer.", ColorOrange, time.Second*10)
	}
	for _, sChunk := range superChunkList {
		sChunk.pixmapDirty = true
	}
	showResourceLayerLock.Unlock()
}
