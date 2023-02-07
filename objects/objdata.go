package objects

import (
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/gv"
	"GameTest/save"
	"bytes"
	"encoding/json"
	"os"
)

var (

	/* Toolbar actions and images */
	UIObjsTypes = []*glob.ObjType{
		//Ui Only
		{Name: "Save", ImagePath: "ui/save.png", ToolbarAction: save.SaveGame,
			Symbol: "SAVE", ItemColor: &glob.ColorRed, SymbolColor: &glob.ColorWhite},
		{Name: "Load", ImagePath: "ui/load.png", ToolbarAction: save.LoadGame,
			Symbol: "LOAD", ItemColor: &glob.ColorBlue, SymbolColor: &glob.ColorWhite},
	}

	/* World objects and images */
	GameObjTypes = []*glob.ObjType{
		//Game Objects
		{ImagePath: "world-obj/basic-miner.png",
			Name:        "Basic miner",
			TypeI:       gv.ObjTypeBasicMiner,
			Size:        glob.XY{X: 1, Y: 1},
			UpdateObj:   minerUpdate,
			MinerKGTock: 1,
			CapacityKG:  500,
			ShowArrow:   true,
			ShowBlocked: true,
			Symbol:      "MINE", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			Ports: [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},

		{ImagePath: "world-obj/basic-belt.png",
			Name:       "Basic belt",
			TypeI:      gv.ObjTypeBasicBelt,
			Size:       glob.XY{X: 1, Y: 1},
			CapacityKG: 20,
			Rotatable:  true,
			UpdateObj:  beltUpdate,
			Symbol:     "BELT", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			Ports: [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},

		{ImagePath: "world-obj/basic-splitter.png",
			Name:        "Basic Splitter",
			TypeI:       gv.ObjTypeBasicSplit,
			Size:        glob.XY{X: 1, Y: 1},
			Rotatable:   true,
			ShowArrow:   true,
			ShowBlocked: true,
			UpdateObj:   splitterUpdate,
			Symbol:      "SPLT", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			Ports: [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},

		{ImagePath: "world-obj/basic-box.png",
			Name:       "Basic box",
			TypeI:      gv.ObjTypeBasicBox,
			Size:       glob.XY{X: 1, Y: 1},
			CapacityKG: 5000,
			Symbol:     "BOX", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			UpdateObj:   boxUpdate,
			CanContain:  true,
			ShowBlocked: true,
			Ports:       [gv.DIR_MAX]uint8{gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},

		{ImagePath: "world-obj/basic-smelter-1.png",
			Name:        "Basic smelter",
			TypeI:       gv.ObjTypeBasicSmelter,
			Size:        glob.XY{X: 1, Y: 1},
			CapacityKG:  50,
			ShowArrow:   true,
			ShowBlocked: true,
			Symbol:      "SMLT", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			UpdateObj: smelterUpdate,
			Ports:     [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},

		{ImagePath: "world-obj/iron-rod-caster.png",
			Name:   "Iron rod caster",
			TypeI:  gv.ObjTypeBasicIronCaster,
			Size:   glob.XY{X: 1, Y: 1},
			Symbol: "CAST", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			UpdateObj:   ironCasterUpdate,
			ShowArrow:   true,
			ShowBlocked: true,
			Ports:       [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},

		{ImagePath: "world-obj/basic-boiler.png",
			Name:        "Basic boiler",
			TypeI:       gv.ObjTypeBasicBoiler,
			Size:        glob.XY{X: 1, Y: 1},
			CapacityKG:  500,
			ShowArrow:   true,
			ShowBlocked: true,
			Symbol:      "BOIL", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			UpdateObj: steamEngineUpdate,
			Ports:     [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},

		{ImagePath: "world-obj/steam-engine.png",
			Name:        "Steam engine",
			TypeI:       gv.ObjTypeSteamEngine,
			Size:        glob.XY{X: 1, Y: 1},
			CapacityKG:  500,
			ShowArrow:   true,
			ShowBlocked: true,
			Symbol:      "STEM", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			UpdateObj: steamEngineUpdate,
			Ports:     [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},
	}

	/* Terrain types and images */
	TerrainTypes = []*glob.ObjType{
		//Overlays
		{ImagePath: "terrain/grass-1.png", Name: "grass",
			Size:   glob.XY{X: 1, Y: 1},
			Symbol: ".", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorGreen},
		{ImagePath: "terrain/dirt-1.png", Name: "dirt",
			Size:   glob.XY{X: 1, Y: 1},
			Symbol: ".", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorBrown},
	}

	/* Overlay images */
	ObjOverlayTypes = []*glob.ObjType{
		//Overlays
		{ImagePath: "overlays/arrow-north.png", Name: "Arrow North",
			Symbol: "^", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorOrange},
		{ImagePath: "overlays/arrow-east.png", Name: "Arrow East",
			Symbol: ">", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorOrange},
		{ImagePath: "overlays/arrow-south.png", Name: "Arrow South",
			Symbol: "v", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorOrange},
		{ImagePath: "overlays/arrow-west.png", Name: "Arrow West",
			Symbol: "<", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorOrange},
		{ImagePath: "overlays/blocked.png", Name: "Blocked",
			Symbol: "*", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorOrange},
	}

	/* Materials and images */
	MatTypes = []*glob.ObjType{
		//Materials
		{Symbol: "ERR", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorRed,
			Name: "Error"},
		{Symbol: "WOOD", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorBrown,
			Name: "Wood"},
		{Symbol: "COAL", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorDarkGray,
			ImagePath: "belt-obj/coal.png"},
		{Symbol: "ERR", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorRed,
			Name: "Error"},
	}

	/* Toolbar item types, array of array of ObjType */
	SubTypes = [][]*glob.ObjType{
		UIObjsTypes,
		GameObjTypes,
		MatTypes,
		ObjOverlayTypes,
		TerrainTypes,
	}
)

func init() {
	for i := range MatTypes {
		MatTypes[i].TypeI = uint8(i)
	}
	for i := range ObjOverlayTypes {
		ObjOverlayTypes[i].TypeI = uint8(i)
	}
	for i := range UIObjsTypes {
		UIObjsTypes[i].TypeI = uint8(i)
	}
	for i := range MatTypes {
		MatTypes[i].TypeI = uint8(i)
	}
}

/* Debug quick dump GameObjTypes */
func DumpItems() bool {

	outbuf := new(bytes.Buffer)
	enc := json.NewEncoder(outbuf)
	enc.SetIndent("", "\t")

	if err := enc.Encode(GameObjTypes); err != nil {
		cwlog.DoLog("DumpItems: %v", err)
		return false
	}

	_, err := os.Create("items.json")

	if err != nil {
		cwlog.DoLog("DumpItems: %v", err)
		return false
	}

	err = os.WriteFile("items.json", outbuf.Bytes(), 0644)

	if err != nil {
		cwlog.DoLog("DumpItems: %v", err)
		return false
	}

	return true
}
