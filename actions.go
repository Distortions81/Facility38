package main

import (
	"Facility38/util"
	"Facility38/world"
	"time"
)

func settingsToggle() {
	if Windows[0].Active {
		CloseWindow(Windows[0])
	} else {
		OpenWindow(Windows[0])
	}
}

func toggleOverlay() {
	if world.OverlayMode {
		world.OverlayMode = false
		util.ChatDetailed("Info overlay is now off.", world.ColorOrange, time.Second*5)
	} else {
		world.OverlayMode = true
		util.ChatDetailed("Info overlay is now on.", world.ColorOrange, time.Second*5)
	}
}

func SwitchLayer() {
	world.ShowResourceLayerLock.Lock()

	if world.ShowResourceLayer {
		world.ShowResourceLayer = false
		util.ChatDetailed("Switched from resource layer to game.", world.ColorOrange, time.Second*10)
	} else {
		world.ShowResourceLayer = true
		util.ChatDetailed("Switched from game to resource layer.", world.ColorOrange, time.Second*10)
	}
	for _, sChunk := range world.SuperChunkList {
		sChunk.PixmapDirty = true
	}
	world.ShowResourceLayerLock.Unlock()
}
