package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

var groundTiles = []*objTypeData{
	{
		base: "tile",
	},
}

/* Toolbar actions and images */
var uiObjs = []*objTypeData{
	//Ui Only
	{
		base: "settings",
		name: "Options", toolbarAction: settingsToggle,
		symbol: "SET", description: "Show game options", qKey: ebiten.KeyF1,
	},
	{
		base: "help",
		name: "Help", toolbarAction: toggleHelp,
		symbol: "?", description: "See game controls and help.", qKey: ebiten.KeyF2,
	},
	{
		base: "changes",
		name: "Changes", toolbarAction: toggleChanges,
		symbol: "CH", description: "Show recent changes to the game.", qKey: ebiten.KeyF3,
	},
	{
		base: "overlay",
		name: "Overlay", toolbarAction: toggleOverlay,
		symbol: "OVRLY", description: "Turn info overlay on/off", qKey: ebiten.KeyF4,
	},
	{
		base: "layer",
		name: "Layer", toolbarAction: switchGameLayer,
		symbol: "LAYER", description: "Toggle between the build and resource layer", qKey: ebiten.KeyF5,
	},
	{
		base: "save", excludeWASM: false,
		name: "Save Game", toolbarAction: saveGame,
		symbol: "SAV", description: "Save game", qKey: ebiten.KeyF6,
	},
	{
		base: "load", excludeWASM: false,
		name: "Load Game", toolbarAction: triggerLoad,
		symbol: "LDG", description: "Load last game", qKey: ebiten.KeyF7,
	},
}

/* Terrain types and images */
var terrainTypes = []*objTypeData{
	{
		base:   "dirt",
		name:   "dirt",
		size:   XYs{X: 1, Y: 1},
		symbol: ".",
	},
	{

		base:   "grass",
		name:   "grass",
		size:   XYs{X: 1, Y: 1},
		symbol: ".",
	},
}

/* Overlay images */
var worldOverlays = []*objTypeData{
	{
		base: "arrow-north",
		name: "Arrow North", symbol: "^"},
	{
		base: "arrow-east",
		name: "Arrow East", symbol: ">"},
	{
		base: "arrow-south",
		name: "Arrow South", symbol: "v"},
	{
		base: "arrow-west",
		name: "Arrow West", symbol: "<"},
	{
		base: "blocked",
		name: "Blocked", symbol: "*"},
	{
		base: "nofuel",
		name: "NO FUEL", symbol: "&"},
	{
		base: "check-on",
		name: "Check-On", symbol: "1"},
	{
		base: "check-off",
		name: "Check-Off", symbol: "0"},
	{
		base: "close",
		name: "Close", symbol: "X"},
	{
		base: "obj-sel",
		name: "obj-sel", symbol: "_"},
}

type subTypeData struct {
	folder string
	list   []*objTypeData
}

/* Toolbar item types, array of array of ObjType */
var subTypes = []subTypeData{
	{
		folder: "ui",
		list:   uiObjs,
	},
	{
		folder: "world-obj",
		list:   worldObjs,
	},
	{
		folder: "overlays",
		list:   worldOverlays,
	},
	{
		folder: "terrain",
		list:   terrainTypes,
	},
	{
		folder: "ground",
		list:   groundTiles,
	},
}
