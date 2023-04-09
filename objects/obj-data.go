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
		Base: "tile",
	},
}

/* Toolbar actions and images */
var UIObjs = []*world.ObjType{
	//Ui Only
	{
		Base: "settings",
		Name: "Options", ToolbarAction: settingsToggle,
		Symbol: "SET", Description: "Show game options", QKey: ebiten.KeyF1,
	},
	{
		Base: "overlay",
		Name: "Overlay", ToolbarAction: toggleOverlay,
		Symbol: "OVRLY", Description: "Turn info overlay on/off", QKey: ebiten.KeyF2,
	},
	{
		Base: "layer",
		Name: "Layer", ToolbarAction: SwitchLayer,
		Symbol: "LAYER", Description: "Toggle between the build and resource layer", QKey: ebiten.KeyF3,
	},
	{
		Base: "save",
		Name: "Save Game", ToolbarAction: SaveGame,
		Symbol: "SAVE", ExcludeWASM: true, Description: "Quicksave game", QKey: ebiten.KeyF5,
	},
	{
		Base: "load",
		Name: "Load Game", ToolbarAction: LoadGame,
		Symbol: "LOAD", ExcludeWASM: true, Description: "Load quicksave", QKey: ebiten.KeyF6,
	},
}

/* Terrain types and images */
var TerrainTypes = []*world.ObjType{
	{
		Base:   "dirt",
		Name:   "dirt",
		Size:   world.XYs{X: 1, Y: 1},
		Symbol: ".",
	},
	{

		Base:   "grass",
		Name:   "grass",
		Size:   world.XYs{X: 1, Y: 1},
		Symbol: ".",
	},
}

/* Overlay images */
var WorldOverlays = []*world.ObjType{
	{
		Base: "arrow-north",
		Name: "Arrow North", Symbol: "^"},
	{
		Base: "arrow-east",
		Name: "Arrow East", Symbol: ">"},
	{
		Base: "arrow-south",
		Name: "Arrow South", Symbol: "v"},
	{
		Base: "arrow-west",
		Name: "Arrow West", Symbol: "<"},
	{
		Base: "blocked",
		Name: "Blocked", Symbol: "*"},
	{
		Base: "nofuel",
		Name: "NO FUEL", Symbol: "&"},
	{
		Base: "check-on",
		Name: "Check-On", Symbol: "✓"},
	{
		Base: "check-off",
		Name: "Check-Off", Symbol: "X"},
	{
		Base: "close",
		Name: "Close", Symbol: "X"},
}

type SubTypeData struct {
	Folder string
	List   []*world.ObjType
}

/* Toolbar item types, array of array of ObjType */
var SubTypes = []SubTypeData{
	{
		Folder: "ui",
		List:   UIObjs,
	},
	{
		Folder: "world-obj",
		List:   WorldObjs,
	},
	{
		Folder: "overlays",
		List:   WorldOverlays,
	},
	{
		Folder: "terrain",
		List:   TerrainTypes,
	},
	{
		Folder: "ground",
		List:   GroundTiles,
	},
}

/* Debug quick dump GameObjTypes
 */
func DumpItems() bool {

	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	if err := enc.Encode(WorldObjs); err != nil {
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
