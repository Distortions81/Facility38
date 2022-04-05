package glob

var (
	ObjTypeNone = 0

	//Subtypes
	ObjSubUI   = 1
	ObjSubGame = 2
	ObjSubMat  = 3

	//UI Only
	ObjTypeSave = 1

	//Buildings
	ObjTypeBasicMiner      = 1
	ObjTypeBasicSmelter    = 2
	ObjTypeBasicIronCaster = 3
	ObjTypeBasicLoader     = 4

	//Materials
	ObjTypeDefault = 1
	ObjTypeWood    = 2
	ObjTypeCoal    = 3
	ObjTypeIron    = 4

	//Automatically set
	GameTypeMax = 0
	UITypeMax   = 0
	MatTypeMax  = 0

	SelectedItemType = 2

	UIObjsTypes = map[int]ObjType{
		//Ui Only
		ObjTypeSave: {ItemColor: &ColorGray, Name: "Save", ImagePath: "ui/save.png", Action: SaveGame, SubType: ObjSubUI},
	}

	GameObjTypes = map[int]ObjType{
		//Game Objects
		ObjTypeBasicMiner:      {ImagePath: "basic-miner.png", Name: "Basic miner", Size: Position{X: 1, Y: 1}, SubType: ObjSubGame},
		ObjTypeBasicSmelter:    {ImagePath: "basic-smelter.png", Name: "Basic smelter", Size: Position{X: 1, Y: 1}, SubType: ObjSubGame},
		ObjTypeBasicIronCaster: {ImagePath: "iron-rod-caster.png", Name: "Iron rod caster", Size: Position{X: 1, Y: 1}, SubType: ObjSubGame},
		ObjTypeBasicLoader:     {ImagePath: "basic-loader.png", Name: "Basic loader", Size: Position{X: 1, Y: 1}, SubType: ObjSubGame},
	}

	MatTypes = map[int]ObjType{
		//Materials
		ObjTypeDefault: {ItemColor: &ColorWhite, Symbol: "?", SymbolColor: &ColorBlack, Name: "Default", Size: Position{X: 1, Y: 1}, SubType: ObjSubMat},
		ObjTypeWood:    {ItemColor: &ColorBrown, Symbol: "w", SymbolColor: &ColorYellow, Name: "Wood", Size: Position{X: 1, Y: 1}, SubType: ObjSubMat},
		ObjTypeCoal:    {ItemColor: &ColorBlack, Symbol: "c", SymbolColor: &ColorWhite, Name: "Coal", Size: Position{X: 1, Y: 1}, SubType: ObjSubMat},
		ObjTypeIron:    {ItemColor: &ColorRust, Symbol: "s", SymbolColor: &ColorBlack, Name: "Iron", Size: Position{X: 1, Y: 1}, SubType: ObjSubMat},
	}

	SubTypes = map[int]map[int]ObjType{
		ObjSubGame: GameObjTypes,
		ObjSubUI:   UIObjsTypes,
		ObjSubMat:  MatTypes,
	}
)
