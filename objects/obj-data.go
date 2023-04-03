package objects

import (
	"GameTest/cwlog"
	"GameTest/world"
	"bytes"
	"encoding/json"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

var GroundTiles = []*world.ObjType{
	{
		Images: world.ObjectImages{
			ImagePath: "gtile/paver.png"},
	},
}

/* Toolbar actions and images */
var UIObjsTypes = []*world.ObjType{
	//Ui Only
	{
		Images: world.ObjectImages{
			ImagePath: "ui/settings.png",
		},
		Name: "Options", ToolbarAction: settingsToggle,
		Symbol: "SET", Description: "Show game options", QKey: ebiten.KeyF1,
	},
	{
		Images: world.ObjectImages{
			ImagePath: "ui/overlay.png",
		},
		Name: "Overlay", ToolbarAction: toggleOverlay,
		Symbol: "OVRLY", Description: "Turn info overlay on/off", QKey: ebiten.KeyF2,
	},
	{
		Images: world.ObjectImages{
			ImagePath: "ui/layer.png",
		},
		Name: "Layer", ToolbarAction: SwitchLayer,
		Symbol: "LAYER", Description: "Toggle between the build and resource layer", QKey: ebiten.KeyF3,
	},
	{
		Name: "Save Game",
		Images: world.ObjectImages{
			ImagePath: "ui/save.png",
		},
		ToolbarAction: SaveGame,
		Symbol:        "SAVE", ExcludeWASM: true, Description: "Quicksave game", QKey: ebiten.KeyF5,
	},
	{
		Name: "Load Game",
		Images: world.ObjectImages{
			ImagePath: "ui/load.png",
		},
		ToolbarAction: LoadGame,
		Symbol:        "LOAD", ExcludeWASM: true, Description: "Load quicksave", QKey: ebiten.KeyF6,
	},
}

/* Terrain types and images */
var TerrainTypes = []*world.ObjType{
	{

		Images: world.ObjectImages{
			ImagePath: "terrain/grass-1.png",
		},
		Name:   "grass",
		Size:   world.XYs{X: 1, Y: 1},
		Symbol: "."},
	{
		Images: world.ObjectImages{
			ImagePath: "terrain/dirt-1.png",
		},
		Name:   "dirt",
		Size:   world.XYs{X: 1, Y: 1},
		Symbol: "."},
}

/* Overlay images */
var ObjOverlayTypes = []*world.ObjType{
	{
		Images: world.ObjectImages{
			ImagePath: "overlays/arrow-north.png",
		},
		Name: "Arrow North", Symbol: "^"},
	{
		Images: world.ObjectImages{
			ImagePath: "overlays/arrow-east.png",
		},
		Name: "Arrow East", Symbol: ">"},
	{
		Images: world.ObjectImages{
			ImagePath: "overlays/arrow-south.png",
		},
		Name: "Arrow South", Symbol: "v"},
	{
		Images: world.ObjectImages{
			ImagePath: "overlays/arrow-west.png",
		},
		Name: "Arrow West", Symbol: "<"},
	{
		Images: world.ObjectImages{
			ImagePath: "overlays/blocked.png",
		},
		Name: "Blocked", Symbol: "*"},
	{
		Images: world.ObjectImages{
			ImagePath: "overlays/nofuel.png",
		},
		Name: "NO FUEL", Symbol: "&"},
	{
		Images: world.ObjectImages{
			ImagePath: "ui/check-on.png",
		},
		Name: "Check-On", Symbol: "✓"},
	{
		Images: world.ObjectImages{
			ImagePath: "ui/check-off.png",
		},
		Name: "Check-Off", Symbol: "X"},
	{
		Images: world.ObjectImages{
			ImagePath: "ui/close.png",
		},
		Name: "Close", Symbol: "X"},
}

/* Toolbar item types, array of array of ObjType */
var SubTypes = [][]*world.ObjType{
	UIObjsTypes,
	GameObjTypes,
	ObjOverlayTypes,
	TerrainTypes,
	GroundTiles,
}

/* Debug quick dump GameObjTypes */
func DumpItems() bool {

	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	if err := enc.Encode(GameObjTypes); err != nil {
		cwlog.DoLog(true, "DumpItems: %v", err)
		return false
	}

	_, err := os.Create("items.json")

	if err != nil {
		cwlog.DoLog(true, "DumpItems: %v", err)
		return false
	}

	err = os.WriteFile("items.json", outbuf.Bytes(), 0644)

	if err != nil {
		cwlog.DoLog(true, "DumpItems: %v", err)
		return false
	}

	return true
}
