package objects

import (
	"GameTest/gv"
	"GameTest/world"
)

/* World objects and images */
var WorldObjs = []*world.ObjType{
	//Game Objects
	{
		Base:        "basic-miner",
		Name:        "Basic Miner",
		Description: "Mines solid resources where placed, requires coal fuel.",
		TypeI:       gv.ObjTypeBasicMiner,
		Category:    gv.ObjCatGeneric,
		Size:        world.XYs{X: 2, Y: 2},
		UpdateObj:   minerUpdate,
		InitObj:     initMiner,
		DeInitObj:   deinitMiner,
		LinkObj:     linkMiner,
		MachineSettings: world.MachineData{
			KgHourMine: 1000,
			KW:         360,
		},
		TockInterval: uint8(world.ObjectUPS/2) * 5,
		ShowArrow:    true,
		ToolBarArrow: true,
		Symbol:       "MIN",
		Ports: []world.ObjPortData{
			{Dir: gv.DIR_NORTH, Type: gv.PORT_OUT},

			/* Fuel inputs */
			{Dir: gv.DIR_NORTH, Type: gv.PORT_FIN},
			{Dir: gv.DIR_EAST, Type: gv.PORT_FIN},
			{Dir: gv.DIR_SOUTH, Type: gv.PORT_FIN},
			{Dir: gv.DIR_WEST, Type: gv.PORT_FIN},
		},
		SubObjs: []world.XYs{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}},
	},

	{
		Base:        "basic-belt",
		Name:        "Basic Belt",
		Description: "Moves items from rear and sides in direction of arrow.",
		TypeI:       gv.ObjTypeBasicBelt,
		Category:    gv.ObjCatBelt,
		Size:        world.XYs{X: 1, Y: 1},
		Rotatable:   true,
		UpdateObj:   beltUpdate,
		LinkObj:     linkBelt,
		Symbol:      "BLT",
		Ports: []world.ObjPortData{
			{Dir: gv.DIR_NORTH, Type: gv.PORT_OUT},
			{Dir: gv.DIR_EAST, Type: gv.PORT_IN},
			{Dir: gv.DIR_SOUTH, Type: gv.PORT_IN},
			{Dir: gv.DIR_WEST, Type: gv.PORT_IN},
		},
	},
	{
		Base:        "basic-belt-over",
		Name:        "Basic Belt Overpass",
		Description: "A belt that has an underpass.",
		TypeI:       gv.ObjTypeBasicBeltOver,
		Category:    gv.ObjCatBelt,
		Size:        world.XYs{X: 1, Y: 3},
		Rotatable:   true,
		UpdateObj:   beltUpdateOver,
		InitObj:     initBeltOver,
		LinkObj:     linkBeltOver,
		Symbol:      "BOP",
		Ports: []world.ObjPortData{
			/* Overpass is one direction */
			{Dir: gv.DIR_NORTH, Type: gv.PORT_OUT},
			{Dir: gv.DIR_SOUTH, Type: gv.PORT_IN},

			/* Underpass is bidirectional */
			{Dir: gv.DIR_WEST, Type: gv.PORT_OUT},
			{Dir: gv.DIR_EAST, Type: gv.PORT_OUT},
			{Dir: gv.DIR_WEST, Type: gv.PORT_IN},
			{Dir: gv.DIR_EAST, Type: gv.PORT_IN},
		},
		SubObjs: []world.XYs{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 0, Y: 2}},
	},

	{
		Base:         "basic-splitter",
		Name:         "Basic Splitter",
		Description:  "Input from back, outputs equally to up to 3 outputs.",
		TypeI:        gv.ObjTypeBasicSplit,
		Category:     gv.ObjCatBelt,
		Size:         world.XYs{X: 1, Y: 1},
		ShowArrow:    true,
		ToolBarArrow: true,
		MachineSettings: world.MachineData{
			KW: 100,
		},
		UpdateObj: splitterUpdate,
		LinkObj:   linkSplitter,
		Symbol:    "SPT",
		Ports: []world.ObjPortData{
			{Dir: gv.DIR_NORTH, Type: gv.PORT_OUT},
			{Dir: gv.DIR_EAST, Type: gv.PORT_OUT},
			{Dir: gv.DIR_WEST, Type: gv.PORT_OUT},
			{Dir: gv.DIR_SOUTH, Type: gv.PORT_IN},
		},
	},

	{
		Base:        "basic-box",
		Name:        "Basic Box",
		Description: "Currently only stores objects (no unloader yet).",
		TypeI:       gv.ObjTypeBasicBox,
		Category:    gv.ObjCatGeneric,
		Size:        world.XYs{X: 2, Y: 2},
		MachineSettings: world.MachineData{
			MaxContainKG: 1000,
		},
		Symbol:       "BOX",
		UpdateObj:    boxUpdate,
		LinkObj:      linkBox,
		TockInterval: uint8(world.ObjectUPS/2) * 1,
		CanContain:   true,
		ToolBarArrow: false,
		Ports: []world.ObjPortData{
			{Dir: gv.DIR_NORTH, Type: gv.PORT_IN},
			{Dir: gv.DIR_EAST, Type: gv.PORT_IN},
			{Dir: gv.DIR_SOUTH, Type: gv.PORT_IN},
			{Dir: gv.DIR_WEST, Type: gv.PORT_IN},

			{Dir: gv.DIR_NORTH, Type: gv.PORT_OUT},
			{Dir: gv.DIR_EAST, Type: gv.PORT_OUT},
			{Dir: gv.DIR_SOUTH, Type: gv.PORT_OUT},
			{Dir: gv.DIR_WEST, Type: gv.PORT_OUT},
		},
		SubObjs: []world.XYs{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}},
	},

	{
		Base:        "basic-smelter",
		Name:        "Basic Smelter",
		Description: "Bakes solid ores into metal or stone bricks, requires coal fuel.",
		TypeI:       gv.ObjTypeBasicSmelter,
		Category:    gv.ObjCatGeneric,
		Size:        world.XYs{X: 2, Y: 2},
		MachineSettings: world.MachineData{
			KW:         320,
			KgHourMine: 40,
		},
		TockInterval: uint8(world.ObjectUPS/2) * 30,
		ShowArrow:    true,
		ToolBarArrow: true,
		Symbol:       "SMT",
		UpdateObj:    smelterUpdate,
		InitObj:      initSmelter,
		LinkObj:      linkMachine,
		Ports: []world.ObjPortData{
			{Dir: gv.DIR_NORTH, Type: gv.PORT_OUT},
			{Dir: gv.DIR_SOUTH, Type: gv.PORT_IN},

			{Dir: gv.DIR_EAST, Type: gv.PORT_FIN},
			{Dir: gv.DIR_WEST, Type: gv.PORT_FIN},
		},
		SubObjs: []world.XYs{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}},
	},

	{
		Base:        "basic-caster",
		Name:        "Basic Caster",
		Description: "Casts metal shot into bars.",
		TypeI:       gv.ObjTypeBasicCaster,
		Category:    gv.ObjCatGeneric,
		Size:        world.XYs{X: 2, Y: 2},
		MachineSettings: world.MachineData{
			KW:         320,
			KgHourMine: 40,
		},
		TockInterval: uint8(world.ObjectUPS/2) * 15,
		ShowArrow:    true,
		ToolBarArrow: true,
		Symbol:       "CST",
		UpdateObj:    casterUpdate,
		InitObj:      initSmelter,
		LinkObj:      linkMachine,
		Ports: []world.ObjPortData{
			{Dir: gv.DIR_NORTH, Type: gv.PORT_OUT},
			{Dir: gv.DIR_SOUTH, Type: gv.PORT_IN},

			{Dir: gv.DIR_EAST, Type: gv.PORT_FIN},
			{Dir: gv.DIR_WEST, Type: gv.PORT_FIN},
		},
		SubObjs: []world.XYs{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}},
	},

	{
		Base:        "basic-rod-caster",
		Name:        "Basic Rod Caster",
		Description: "Casts metal bars into rods.",
		TypeI:       gv.ObjTypeBasicRodCaster,
		Category:    gv.ObjCatGeneric,
		Size:        world.XYs{X: 2, Y: 2},
		MachineSettings: world.MachineData{
			KW:         320,
			KgHourMine: 40,
		},
		TockInterval: uint8(world.ObjectUPS/2) * 15,
		ShowArrow:    true,
		ToolBarArrow: true,
		Symbol:       "ROD",
		UpdateObj:    rodCasterUpdate,
		InitObj:      initSmelter,
		LinkObj:      linkMachine,
		Ports: []world.ObjPortData{
			{Dir: gv.DIR_NORTH, Type: gv.PORT_OUT},
			{Dir: gv.DIR_SOUTH, Type: gv.PORT_IN},

			{Dir: gv.DIR_EAST, Type: gv.PORT_FIN},
			{Dir: gv.DIR_WEST, Type: gv.PORT_FIN},
		},
		SubObjs: []world.XYs{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}},
	},

	{
		Base:        "basic-fuel-hopper",
		Name:        "Basic Fuel Hopper",
		Description: "Loads soild fuel into machines",
		TypeI:       gv.ObjTypeBasicFuelHopper,
		Category:    gv.ObjCatLoader,
		Size:        world.XYs{X: 1, Y: 1},
		Rotatable:   true,
		ShowArrow:   false,
		UpdateObj:   fuelHopperUpdate,
		LinkObj:     linkFuelHopper,
		MachineSettings: world.MachineData{
			KW:           10,
			KgHopperMove: 1,
		},
		TockInterval: uint8(world.ObjectUPS/2) * 2,
		Symbol:       "FHP",
		Ports: []world.ObjPortData{
			{Dir: gv.DIR_NORTH, Type: gv.PORT_IN},
			{Dir: gv.DIR_EAST, Type: gv.PORT_IN},
			{Dir: gv.DIR_WEST, Type: gv.PORT_IN},
			{Dir: gv.DIR_SOUTH, Type: gv.PORT_FOUT},
		},
	},
	{
		Base:        "basic-unloader",
		Name:        "Basic Unloader",
		Description: "Unloads Material from objects.",
		TypeI:       gv.ObjTypeBasicUnloader,
		Category:    gv.ObjCatLoader,
		Size:        world.XYs{X: 1, Y: 1},
		Rotatable:   true,
		UpdateObj:   loaderUpdate,
		LinkObj:     linkUnloader,
		MachineSettings: world.MachineData{
			KW:           10,
			KgHopperMove: 1,
		},
		TockInterval: 0,
		Symbol:       "ULD",
		Ports: []world.ObjPortData{
			{Dir: gv.DIR_NORTH, Type: gv.PORT_OUT},
			{Dir: gv.DIR_SOUTH, Type: gv.PORT_IN},
		},
	},
	{
		Base:        "basic-loader",
		Name:        "Basic Loader",
		Description: "Loads Material into objects.",
		TypeI:       gv.ObjTypeBasicLoader,
		Category:    gv.ObjCatLoader,
		Size:        world.XYs{X: 1, Y: 1},
		Rotatable:   true,
		ShowArrow:   false,
		UpdateObj:   loaderUpdate,
		LinkObj:     linkUnloader,
		MachineSettings: world.MachineData{
			KW:           10,
			KgHopperMove: 1,
		},
		TockInterval: 0,
		Symbol:       "LD",
		Ports: []world.ObjPortData{
			{Dir: gv.DIR_NORTH, Type: gv.PORT_OUT},
			{Dir: gv.DIR_SOUTH, Type: gv.PORT_IN},
		},
	},
	{
		Base:        "basic-slip-roller",
		Name:        "Basic Slip Roller",
		Description: "Press metal bar into sheets",
		TypeI:       gv.ObjTypeBasicSlipRoller,
		Category:    gv.ObjCatGeneric,
		Size:        world.XYs{X: 2, Y: 2},
		UpdateObj:   slipRollerUpdate,
		LinkObj:     linkMachine,
		MachineSettings: world.MachineData{
			KW: 10,
		},
		TockInterval: uint8(world.ObjectUPS/2) * 15,
		Symbol:       "SR",
		Ports: []world.ObjPortData{
			{Dir: gv.DIR_NORTH, Type: gv.PORT_OUT},
			{Dir: gv.DIR_SOUTH, Type: gv.PORT_IN},
		},
	},
}
