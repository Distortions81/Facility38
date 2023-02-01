package objects

import (
	"GameTest/consts"
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/save"
	"bytes"
	"encoding/json"
	"os"
)

var (
	UIObjsTypes = []*glob.ObjType{
		//Ui Only
		{Name: "Save", ImagePath: "ui/save.png", ToolbarAction: save.SaveGame,
			Symbol: "SAVE", ItemColor: &glob.ColorRed, SymbolColor: &glob.ColorWhite},
		{Name: "Load", ImagePath: "ui/load.png", ToolbarAction: save.LoadGame,
			Symbol: "LOAD", ItemColor: &glob.ColorBlue, SymbolColor: &glob.ColorWhite},
	}

	GameObjTypes = []*glob.ObjType{
		//Game Objects
		{ImagePath: "world-obj/basic-miner.png",
			Name:        "Basic miner",
			TypeI:       consts.ObjTypeBasicMiner,
			Size:        glob.XY{X: 1, Y: 1},
			UpdateObj:   minerUpdate,
			MinerKGTock: 1,
			CapacityKG:  500,
			Symbol:      "MINE", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			HasMatOutput: true,
		},

		{ImagePath: "world-obj/basic-belt.png",
			Name:       "Basic belt",
			TypeI:      consts.ObjTypeBasicBelt,
			Size:       glob.XY{X: 1, Y: 1},
			CapacityKG: 20,
			Rotatable:  true,
			UpdateObj:  beltUpdate,
			Symbol:     "BELT", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			HasMatInput:  1,
			HasMatOutput: true},

		{ImagePath: "world-obj/basic-splitter.png",
			Name:      "Basic Splitter",
			TypeI:     consts.ObjTypeBasicSplit,
			Size:      glob.XY{X: 1, Y: 1},
			Rotatable: true,
			UpdateObj: splitterUpdate,
			Symbol:    "SPLIT", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			HasMatInput:  2,
			HasMatOutput: true},

		{ImagePath: "world-obj/basic-box.png",
			Name:       "Basic box",
			TypeI:      consts.ObjTypeBasicBox,
			Size:       glob.XY{X: 1, Y: 1},
			CapacityKG: 5000,
			Symbol:     "BOX", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			UpdateObj:   boxUpdate,
			HasMatInput: consts.DIR_MAX},

		{ImagePath: "world-obj/basic-smelter-1.png",
			Name:       "Basic smelter",
			TypeI:      consts.ObjTypeBasicSmelter,
			Size:       glob.XY{X: 1, Y: 1},
			CapacityKG: 50,
			Symbol:     "SMELT", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			UpdateObj:    smelterUpdate,
			HasMatInput:  2,
			HasMatOutput: true},

		{ImagePath: "world-obj/iron-rod-caster.png",
			Name:   "Iron rod caster",
			TypeI:  consts.ObjTypeBasicIronCaster,
			Size:   glob.XY{X: 1, Y: 1},
			Symbol: "CAST", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			UpdateObj:    ironCasterUpdate,
			HasMatInput:  2,
			HasMatOutput: true},

		{ImagePath: "world-obj/basic-boiler.png",
			Name:       "Basic boiler",
			TypeI:      consts.ObjTypeBasicBoiler,
			Size:       glob.XY{X: 1, Y: 1},
			CapacityKG: 500,
			Symbol:     "BOIL", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			UpdateObj:    steamEngineUpdate,
			HasMatInput:  1,
			HasMatOutput: true},

		{ImagePath: "world-obj/steam-engine.png",
			Name:       "Steam engine",
			TypeI:      consts.ObjTypeSteamEngine,
			Size:       glob.XY{X: 1, Y: 1},
			CapacityKG: 500,
			Symbol:     "STEAM", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			UpdateObj:   steamEngineUpdate,
			HasMatInput: 1},
	}

	TerrainTypes = []*glob.ObjType{
		//Overlays
		{ImagePath: "terrain/grass-1.png", Name: "grass",
			Size:   glob.XY{X: 1, Y: 1},
			Symbol: ".", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorGreen},
		{ImagePath: "terrain/dirt-1.png", Name: "dirt",
			Size:   glob.XY{X: 1, Y: 1},
			Symbol: ".", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorBrown},
	}

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

	MatTypes = []*glob.ObjType{
		//Materials
		{Symbol: "?", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorRed, Name: "Error"},
		{Symbol: "WOOD", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorBrown, Name: "Wood"},
		{Symbol: "COAL", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorDarkGray, ImagePath: "belt-obj/coal.png", Name: "Coal Ore"},
		{Symbol: "COPPER", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorAqua, Name: "Copper Ore"},
	}

	SubTypes = [][]*glob.ObjType{
		UIObjsTypes,
		GameObjTypes,
		MatTypes,
		ObjOverlayTypes,
		TerrainTypes,
	}
)

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
