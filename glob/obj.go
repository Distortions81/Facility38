package glob

var (

	//Subtypes
	ObjSubUI   = 0
	ObjSubGame = 1
	ObjSubMat  = 2

	ObjTypeNone = 0

	//UI Only
	ObjTypeSave = 1

	//Buildings
	ObjTypeMiner     = 2
	ObjTypeSmelter   = 3
	ObjTypeAssembler = 4
	ObjTypeTower     = 5

	//Materials
	ObjTypeDefault = 1
	ObjTypeWood    = 2
	ObjTypeCoal    = 3
	ObjTypeIron    = 4

	ObjTypeMax       = 0 //Automatically set
	SelectedItemType = 2

	ObjTypes = map[int]ObjType{
		ObjTypeNone: {ItemColor: &ColorTransparent},

		//Ui Only
		ObjTypeSave: {ItemColor: &ColorGray, Name: "Save", ImagePath: "save.png", Action: SaveGame, SubType: ObjSubUI},

		//Game Objects
		ObjTypeMiner:     {ItemColor: &ColorWhite, Symbol: "M", SymbolColor: &ColorGray, Name: "Miner", Size: Position{X: 1, Y: 1}, SubType: ObjSubGame},
		ObjTypeSmelter:   {ItemColor: &ColorOrange, Symbol: "S", SymbolColor: &ColorWhite, Name: "Smelter", Size: Position{X: 1, Y: 1}, SubType: ObjSubGame},
		ObjTypeAssembler: {ItemColor: &ColorGray, Symbol: "A", SymbolColor: &ColorBlack, Name: "Assembler", Size: Position{X: 1, Y: 1}, SubType: ObjSubGame},
		ObjTypeTower:     {ItemColor: &ColorRed, Symbol: "T", SymbolColor: &ColorWhite, Name: "Tower", Size: Position{X: 1, Y: 1}, SubType: ObjSubGame},
	}

	Mats = map[int]ObjType{
		//Materials
		ObjTypeDefault: {ItemColor: &ColorWhite, Symbol: "?", SymbolColor: &ColorBlack, Name: "Default", Size: Position{X: 1, Y: 1}, SubType: ObjSubMat},
		ObjTypeWood:    {ItemColor: &ColorBrown, Symbol: "w", SymbolColor: &ColorYellow, Name: "Wood", Size: Position{X: 1, Y: 1}, SubType: ObjSubMat},
		ObjTypeCoal:    {ItemColor: &ColorBlack, Symbol: "c", SymbolColor: &ColorWhite, Name: "Coal", Size: Position{X: 1, Y: 1}, SubType: ObjSubMat},
		ObjTypeIron:    {ItemColor: &ColorRust, Symbol: "s", SymbolColor: &ColorBlack, Name: "Iron", Size: Position{X: 1, Y: 1}, SubType: ObjSubMat},
	}
)
