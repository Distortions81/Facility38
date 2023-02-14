package objects

import (
	"GameTest/cwlog"
	"GameTest/gv"
	"GameTest/world"
	"bytes"
	"encoding/json"
	"os"
)

func init() {

	/* Pre-calculate some object values */
	for i := range GameObjTypes {
		if GameObjTypes[i].KgHourMine > 0 {
			GameObjTypes[i].KgMineEach = ((GameObjTypes[i].KgHourMine / 60 / 60 / world.ObjectUPS) * float32(GameObjTypes[i].Interval)) * gv.TIMESCALE_MULTI
		}
		if GameObjTypes[i].HP > 0 {
			KW := GameObjTypes[i].HP * gv.HP_PER_KW
			COALKG := KW / gv.COAL_KWH_PER_KG
			GameObjTypes[i].KgFuelEach = ((COALKG / 60 / 60 / world.ObjectUPS) * float32(GameObjTypes[i].Interval)) * gv.TIMESCALE_MULTI
		} else if GameObjTypes[i].KW > 0 {
			COALKG := GameObjTypes[i].KW / gv.COAL_KWH_PER_KG
			GameObjTypes[i].KgFuelEach = ((COALKG / 60 / 60 / world.ObjectUPS) * float32(GameObjTypes[i].Interval)) * gv.TIMESCALE_MULTI
		}

		if GameObjTypes[i].KgFuelEach > 0 {
			GameObjTypes[i].MaxFuelKG = (GameObjTypes[i].KgFuelEach * 10)
			if GameObjTypes[i].MaxFuelKG < 50 {
				GameObjTypes[i].MaxFuelKG = 50
			}
		}

		if GameObjTypes[i].KgMineEach > 0 {
			GameObjTypes[i].MaxContainKG = (GameObjTypes[i].KgMineEach * 10)
			if GameObjTypes[i].MaxContainKG < 50 {
				GameObjTypes[i].MaxContainKG = 50
			}
		}
	}
}

var (

	/* Toolbar actions and images */
	UIObjsTypes = []*world.ObjType{
		//Ui Only
		{Name: "Save Game", ImagePath: "ui/save.png", ToolbarAction: SaveGame,
			Symbol: "SAVE", ExcludeWASM: true, Info: "Quicksave game to the single save-slot."},
		{Name: "Load Game", ImagePath: "ui/load.png", ToolbarAction: LoadGame,
			Symbol: "LOAD", ExcludeWASM: true, Info: "Erase map, and load quicksave."},
		{ImagePath: "ui/layer.png", Name: "Layer", ToolbarAction: SwitchLayer,
			Symbol: "LAYER", Info: "Toggle between the normal and Resource layer."},
		{ImagePath: "ui/overlay.png", Name: "Overlay", ToolbarAction: toggleOverlay,
			Symbol: "OVRLY", Info: "Toggle info overlays on/off"},
	}

	/* World objects and images */
	GameObjTypes = []*world.ObjType{
		//Game Objects
		{ImagePath: "world-obj/basic-miner.png", ImagePathActive: "world-obj/basic-miner-active.png",
			UIPath:       "ui/miner.png",
			Name:         "Basic miner",
			Info:         "Mines soild resources where placed, requires coal fuel.",
			TypeI:        gv.ObjTypeBasicMiner,
			Size:         world.XY{X: 1, Y: 1},
			UpdateObj:    minerUpdate,
			InitObj:      initMiner,
			KgHourMine:   1000,
			KW:           360,
			Interval:     uint8(world.ObjectUPS) * 2,
			ShowArrow:    true,
			ToolBarArrow: true,
			ShowBlocked:  true,
			Symbol:       "MINE",
			Ports:        [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},

		{ImagePath: "world-obj/basic-belt.png",
			Name:      "Basic belt",
			Info:      "Moves items from rear or sides in direction of arrow.",
			TypeI:     gv.ObjTypeBasicBelt,
			Size:      world.XY{X: 1, Y: 1},
			Rotatable: true,
			UpdateObj: beltUpdate,
			Symbol:    "BELT",
			Ports:     [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},
		{ImagePath: "world-obj/basic-belt-inter-right.png",
			Name:      "Basic Intersection-Right",
			Info:      "A belt that has an under-pass going to the right when facing north.",
			TypeI:     gv.ObjTypeBasicBeltInterRight,
			Size:      world.XY{X: 1, Y: 1},
			Rotatable: true,
			UpdateObj: beltUpdateInter,
			Symbol:    "iBLT",
			Ports:     [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},
		{ImagePath: "world-obj/basic-belt-inter-left.png",
			Name:      "Basic Intersection-Left",
			Info:      "A belt that has an under-pass going to the left when facing north.",
			TypeI:     gv.ObjTypeBasicBeltInterLeft,
			Size:      world.XY{X: 1, Y: 1},
			Rotatable: true,
			UpdateObj: beltUpdateInter,
			Symbol:    "iBLT",
			Ports:     [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_OUTPUT},
		},

		{ImagePath: "world-obj/basic-splitter.png", ImagePathActive: "world-obj/basic-splitter-active.png",
			Name:        "Basic Splitter",
			Info:        "Input from back, ouputs equally to up to 3 outputs.",
			TypeI:       gv.ObjTypeBasicSplit,
			Size:        world.XY{X: 1, Y: 1},
			Rotatable:   true,
			ShowArrow:   false,
			ShowBlocked: true,
			Interval:    1,
			KW:          100,
			UpdateObj:   splitterUpdate,
			Symbol:      "SPLT",
			Ports:       [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_OUTPUT},
		},

		{ImagePath: "world-obj/basic-box.png", ImagePathActive: "world-obj/basic-box-active.png",
			Info:         "Currently only stores objects (no unloader yet).",
			UIPath:       "ui/box.png",
			Name:         "Basic box",
			TypeI:        gv.ObjTypeBasicBox,
			Size:         world.XY{X: 1, Y: 1},
			MaxContainKG: 1000,
			Symbol:       "BOX",
			UpdateObj:    boxUpdate,
			CanContain:   true,
			ShowBlocked:  false,
			ToolBarArrow: false,
			Ports:        [gv.DIR_MAX]uint8{gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},

		{ImagePath: "world-obj/basic-smelter.png", ImagePathActive: "world-obj/basic-smelter-active.png",
			UIPath:       "ui/smelter.png",
			Name:         "Basic smelter",
			Info:         "Bakes solid ores into metal or stone bricks, requires coal fuel.",
			TypeI:        gv.ObjTypeBasicSmelter,
			Size:         world.XY{X: 1, Y: 1},
			KW:           320,
			KgHourMine:   40,
			Interval:     uint8(world.ObjectUPS * 60),
			ShowArrow:    true,
			ShowBlocked:  true,
			ToolBarArrow: true,
			Symbol:       "SMLT",
			UpdateObj:    smelterUpdate,
			Ports:        [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},

		{ImagePath: "world-obj/basic-fuel-hopper.png",
			Name:        "Basic Fuel Hopper",
			Info:        "Not Yet Implemented, loads fuel into buildings.",
			TypeI:       gv.ObjTypeBasicFuelHopper,
			Size:        world.XY{X: 1, Y: 1},
			Rotatable:   true,
			ShowArrow:   false,
			ShowBlocked: true,
			UpdateObj:   fuelHopperUpdate,
			Symbol:      "FHOP",
			Ports:       [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
		},

		/*		{ImagePath: "world-obj/iron-rod-caster.png",
					Name:   "Iron rod caster",
					TypeI:  gv.ObjTypeBasicIronCaster,
					Size:   world.XY{X: 1, Y: 1},
					Symbol: "CAST", ItemColor: &world.ColorVeryDarkGray, SymbolColor: &world.ColorWhite,
					UpdateObj:   ironCasterUpdate,
					ShowArrow:   true,
					ShowBlocked: true,
					Ports:       [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
				},

				{ImagePath: "world-obj/basic-boiler.png",
					Name:        "Basic boiler",
					TypeI:       gv.ObjTypeBasicBoiler,
					Size:        world.XY{X: 1, Y: 1},
					CapacityKG:  500,
					ShowArrow:   true,
					ShowBlocked: true,
					Symbol:      "BOIL", ItemColor: &world.ColorVeryDarkGray, SymbolColor: &world.ColorWhite,
					UpdateObj: steamEngineUpdate,
					Ports:     [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
				},

				{ImagePath: "world-obj/steam-engine.png",
					Name:        "Steam engine",
					TypeI:       gv.ObjTypeSteamEngine,
					Size:        world.XY{X: 1, Y: 1},
					CapacityKG:  500,
					ShowArrow:   true,
					ShowBlocked: true,
					Symbol:      "STEM", ItemColor: &world.ColorVeryDarkGray, SymbolColor: &world.ColorWhite,
					UpdateObj: steamEngineUpdate,
					Ports:     [gv.DIR_MAX]uint8{gv.PORT_OUTPUT, gv.PORT_INPUT, gv.PORT_INPUT, gv.PORT_INPUT},
				}, */
	}

	/* Terrain types and images */
	TerrainTypes = []*world.ObjType{
		//Overlays
		{ImagePath: "terrain/grass-1.png", Name: "grass",
			Size:   world.XY{X: 1, Y: 1},
			Symbol: "."},
		{ImagePath: "terrain/dirt-1.png", Name: "dirt",
			Size:   world.XY{X: 1, Y: 1},
			Symbol: "."},
	}

	/* Overlay images */
	ObjOverlayTypes = []*world.ObjType{
		//Overlays
		{ImagePath: "overlays/arrow-north.png", Name: "Arrow North", Symbol: "^"},
		{ImagePath: "overlays/arrow-east.png", Name: "Arrow East", Symbol: ">"},
		{ImagePath: "overlays/arrow-south.png", Name: "Arrow South", Symbol: "v"},
		{ImagePath: "overlays/arrow-west.png", Name: "Arrow West", Symbol: "<"},
		{ImagePath: "overlays/blocked.png", Name: "Blocked", Symbol: "*"},
		{ImagePath: "overlays/nofuel.png", Name: "NO FUEL", Symbol: "&"},
	}

	/* Materials and images */
	MatTypes = []*world.MaterialType{
		//Materials
		{Symbol: "NIL", Name: "NONE", TypeI: gv.MAT_NONE},

		{Symbol: "C", Name: "Coal", UnitName: " kg", ImagePath: "belt-obj/coal-ore.png",
			IsSolid: true, TypeI: gv.MAT_COAL},

		{Symbol: "Oil", Name: "Oil", UnitName: " L", ImagePath: "belt-obj/coal-ore.png",
			IsFluid: true, TypeI: gv.MAT_OIL},

		{Symbol: "Gas", Name: "Natural Gas", UnitName: " cm", ImagePath: "belt-obj/coal-ore.png",
			IsGas: true, TypeI: gv.MAT_GAS},

		/* Ore */
		{Symbol: "FEo", Name: "Iron Ore", UnitName: " kg", ImagePath: "belt-obj/iron-ore.png",
			IsSolid: true, Result: gv.MAT_IRON, TypeI: gv.MAT_IRON_ORE},

		{Symbol: "Cuo", Name: "Copper Ore", UnitName: " kg", ImagePath: "belt-obj/copper-ore.png",
			IsSolid: true, Result: gv.MAT_COPPER, TypeI: gv.MAT_COPPER_ORE},

		{Symbol: "STOo", Name: "Stone Ore", UnitName: " kg", ImagePath: "belt-obj/stone-ore.png",
			IsSolid: true, Result: gv.MAT_STONE, TypeI: gv.MAT_STONE_ORE},

		/* Metal */
		{Symbol: "FE", Name: "Iron Bar", UnitName: " kg", ImagePath: "belt-obj/iron.png",
			IsSolid: true, TypeI: gv.MAT_IRON},

		{Symbol: "Cu", Name: "Copper Bar", UnitName: " kg", ImagePath: "belt-obj/copper.png",
			IsSolid: true, TypeI: gv.MAT_COPPER},

		{Symbol: "STO", Name: "Stone Block", UnitName: " kg", ImagePath: "belt-obj/stone.png",
			IsSolid: true, TypeI: gv.MAT_STONE},

		{Symbol: "SLG", Name: "Slag", UnitName: " kg", ImagePath: "belt-obj/stone.png",
			IsSolid: true, TypeI: gv.MAT_SLAG},
	}

	/* Toolbar item types, array of array of ObjType */
	SubTypes = [][]*world.ObjType{
		UIObjsTypes,
		GameObjTypes,
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
