package obj

import (
	"GameTest/consts"
	"GameTest/glob"
	"fmt"
	"time"
)

var (
	Tick             uint64 = 0
	CurrentWorldStep int
	TickList         map[uint64][]glob.TickEvent
	TockList         map[uint64][]glob.TickEvent
	AddToWorld       []*glob.MObj
	DelFromWorld     []*glob.MObj
)

func GLogic() {

	lastUpdate := time.Now()

	for {

		if time.Since(lastUpdate) > consts.GameLogicRate {
			Tick++

			RunTicks()
			//RunMods()
			RunTocks()
		}

		//Reduce busy waiting
		time.Sleep(consts.GameLogicSleep)
	}
}

func MinerUpdate(key glob.Position, o *glob.MObj) {
	//oData := glob.GameObjTypes[Obj.Type]
}

func SmelterUpdate(key glob.Position, obj *glob.MObj) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func IronCasterUpdate(key glob.Position, obj *glob.MObj) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func LoaderUpdate(key glob.Position, obj *glob.MObj) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func RunTicks() {
	//wg := sizedwaitgroup.New(runtime.NumCPU())

	for _, event := range TickList[Tick] {
		for dir, o := range event.Target.SendTo {
			if o.Contents[dir].Amount > 0 && len(o.SendTo) > 0 {
				//Send to object
				if o.SendTo[dir].External[dir].Amount == 0 {
					o.SendTo[dir].External[dir] = o.Contents[dir]

					fmt.Println("Sent ", o.External[dir].Amount, " to ", o.SendTo[dir].External[dir].Type)
					o.Contents[dir] = nil
				}
			}
		}
	}
}

func RunTocks() {

	for _, event := range TockList[Tick] {
		for dir, o := range event.Target.External {

			if o.Amount > 0 {
				event.Target.Contents[dir] = o

				fmt.Println("Sent ", o.Amount, " to ", o.Type)
				o = nil
			}
		}
	}
}

func RevDir(dir int) int {
	if dir == consts.DIR_NORTH {
		return consts.DIR_SOUTH
	} else if dir == consts.DIR_SOUTH {
		return consts.DIR_NORTH
	} else if dir == consts.DIR_EAST {
		return consts.DIR_WEST
	} else if dir == consts.DIR_WEST {
		return consts.DIR_EAST
	} else {
		return -1
	}
}
