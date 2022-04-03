package glob

var (
	ObjTypeNone      = 0
	ObjTypeSave      = 1
	ObjTypeMiner     = 2
	ObjTypeSmelter   = 3
	ObjTypeAssembler = 4
	ObjTypeTower     = 5

	ObjTypeDefault = 100
	ObjTypeWood    = 101
	ObjTypeCoal    = 102
	ObjTypeIron    = 103

	ObjTypeMax       = 0 //Automatically set
	SelectedItemType = 2

	ObjTypes = map[int]ObjType{
		ObjTypeNone:      {ItemColor: &ColorTransparent},
		ObjTypeSave:      {ItemColor: &ColorGray, Name: "Save", ImagePath: "save.png", Action: SaveGame},
		ObjTypeMiner:     {ItemColor: &ColorWhite, Symbol: "M", SymbolColor: &ColorGray, Name: "Miner", Size: Position{X: 1, Y: 1}, GameObj: true},
		ObjTypeSmelter:   {ItemColor: &ColorOrange, Symbol: "S", SymbolColor: &ColorWhite, Name: "Smelter", Size: Position{X: 1, Y: 1}, GameObj: true},
		ObjTypeAssembler: {ItemColor: &ColorGray, Symbol: "A", SymbolColor: &ColorBlack, Name: "Assembler", Size: Position{X: 1, Y: 1}, GameObj: true},
		ObjTypeTower:     {ItemColor: &ColorRed, Symbol: "T", SymbolColor: &ColorWhite, Name: "Tower", Size: Position{X: 1, Y: 1}, GameObj: true},
	}
)
