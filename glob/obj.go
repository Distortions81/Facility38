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
	ObjTypeMiner     = 1
	ObjTypeSmelter   = 2
	ObjTypeAssembler = 3
	ObjTypeTower     = 4

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
		ObjTypeSave: {ItemColor: &ColorGray, Name: "Save", ImagePath: "save.png", Action: SaveGame, SubType: ObjSubUI},
	}

	GameObjTypes = map[int]ObjType{
		//Game Objects
		ObjTypeMiner:     {ItemColor: &ColorWhite, Symbol: "M", SymbolColor: &ColorGray, ImagePath: "miner.png", Name: "Miner", Size: Position{X: 2, Y: 2}, SubType: ObjSubGame},
		ObjTypeSmelter:   {ItemColor: &ColorOrange, Symbol: "S", SymbolColor: &ColorWhite, ImagePath: "furnace.png", Name: "Smelter", Size: Position{X: 1, Y: 1}, SubType: ObjSubGame},
		ObjTypeAssembler: {ItemColor: &ColorGray, Symbol: "A", SymbolColor: &ColorBlack, Name: "Assembler", Size: Position{X: 1, Y: 1}, SubType: ObjSubGame},
		ObjTypeTower:     {ItemColor: &ColorRed, Symbol: "T", SymbolColor: &ColorWhite, Name: "Tower", Size: Position{X: 1, Y: 1}, SubType: ObjSubGame},
	}

	MatTypes = map[int]ObjType{
		//Materials
		ObjTypeDefault: {ItemColor: &ColorWhite, Symbol: "?", SymbolColor: &ColorBlack, Name: "Default", Size: Position{X: 1, Y: 1}, SubType: ObjSubMat},
		ObjTypeWood:    {ItemColor: &ColorBrown, Symbol: "w", SymbolColor: &ColorYellow, Name: "Wood", Size: Position{X: 1, Y: 1}, SubType: ObjSubMat},
		ObjTypeCoal:    {ItemColor: &ColorBlack, Symbol: "c", SymbolColor: &ColorWhite, Name: "Coal", Size: Position{X: 1, Y: 1}, SubType: ObjSubMat},
		ObjTypeIron:    {ItemColor: &ColorRust, Symbol: "s", SymbolColor: &ColorBlack, Name: "Iron", Size: Position{X: 1, Y: 1}, SubType: ObjSubMat},
	}

	Toolbar = []int{
		ObjTypeSave,
		ObjTypeMiner,
	}

	SubTypes = map[int]map[int]ObjType{
		ObjSubGame: GameObjTypes,
		ObjSubUI:   UIObjsTypes,
		ObjSubMat:  MatTypes,
	}
)
