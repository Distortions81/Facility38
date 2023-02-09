package objects

import (
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/gv"
	"bytes"
	"encoding/json"
	"os"
)

var (

	/* Toolbar actions and images */
	UIObjsTypes = []*glob.ObjType{
		//Ui Only
		{Name: "Save", ImagePath: "ui/save.png", ToolbarAction: SaveGame,
			Symbol: "SAVE", ItemColor: &glob.ColorRed, SymbolColor: &glob.ColorWhite},
		{Name: "Load", ImagePath: "ui/load.png", ToolbarAction: LoadGame,
			Symbol: "LOAD", ItemColor: &glob.ColorBlue, SymbolColor: &glob.ColorWhite},
		{ImagePath: "ui/layer.png", Name: "Layer", ToolbarAction: SwitchLayer,
			Symbol: "LAYER", ItemColor: &glob.ColorGray, SymbolColor: &glob.ColorWhite},
	}

	/* World objects and images */
	GameObjTypes = []*glob.ObjType{
		//Game Objects
		{ImagePath: "world-obj/basic-miner.png", UIPath: "ui/miner.png",
			Name:        "Basic miner",
			TypeI:       gv.ObjTypeBasicMiner,
			Size:        glob.XY{X: 1, Y: 1},
			UpdateObj:   minerUpdate,
			MinerKGTock: 1,
			Interval:    4,
			CapacityKG:  500,
			ShowArrow:   true,
			ShowBlocked: true,
			Symbol:      "MINE", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			Ports: [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},

		{ImagePath: "world-obj/basic-belt.png", UIPath: "ui/belt.png",
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

		{ImagePath: "world-obj/basic-box.png", UIPath: "ui/box.png",
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

		{ImagePath: "world-obj/basic-smelter-1.png", UIPath: "ui/smelter.png",
			Name:        "Basic smelter",
			TypeI:       gv.ObjTypeBasicSmelter,
			Size:        glob.XY{X: 1, Y: 1},
			CapacityKG:  50,
			Interval:    20,
			ShowArrow:   true,
			ShowBlocked: true,
			Symbol:      "SMLT", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorWhite,
			UpdateObj: smelterUpdate,
			Ports:     [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},

		/*		{ImagePath: "world-obj/iron-rod-caster.png",
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
				}, */
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
		{Symbol: "NIL", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorRed,
			Name: "NONE"},
		{Symbol: "C", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorDarkGray,
			Name: "Coal", UnitName: " kg", ImagePath: "belt-obj/coal-ore.png"},

		/* Ore */
		{Symbol: "FEo", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorDarkGray,
			Name: "Iron Ore", UnitName: " kg", ImagePath: "belt-obj/iron-ore.png", IsOre: true, Result: gv.MAT_IRON},
		{Symbol: "Cuo", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorDarkGray,
			Name: "Copper Ore", UnitName: " kg", ImagePath: "belt-obj/copper-ore.png", IsOre: true, Result: gv.MAT_COPPER},
		{Symbol: "STOo", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorDarkGray,
			Name: "Stone Ore", UnitName: " kg", ImagePath: "belt-obj/stone-ore.png", IsOre: true, Result: gv.MAT_STONE},

		/* Metal */
		{Symbol: "FE", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorDarkGray,
			Name: "Iron Bar", UnitName: " kg", ImagePath: "belt-obj/iron.png"},
		{Symbol: "Cu", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorDarkGray,
			Name: "Copper Bar", UnitName: " kg", ImagePath: "belt-obj/copper.png"},
		{Symbol: "STO", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorDarkGray,
			Name: "Stone Block", UnitName: " kg", ImagePath: "belt-obj/stone.png"},
		{Symbol: "SLG", ItemColor: &glob.ColorVeryDarkGray, SymbolColor: &glob.ColorDarkGray,
			Name: "Slag", UnitName: " kg", ImagePath: "belt-obj/stone.png"},
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
