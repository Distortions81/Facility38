package obj

import (
	"GameTest/consts"
	"GameTest/glob"
	"fmt"
	"time"
)

func GLogic() {

	lastUpdate := time.Now()
	var ticks uint64 = 0

	for {
		if time.Since(lastUpdate) > consts.GameLogicRate {
			ticks++
			fmt.Println("Tick:", ticks)

			for _, item := range glob.WorldMap {
				for okey, o := range item.MObj {
					oData := GameObjTypes[o.Type]
					if oData.ObjUpdate != nil {
						//fmt.Println(okey, oData.Name)
						oData.ObjUpdate(okey, o)
					}
				}
			}
		}

		//Reduce busy waiting
		time.Sleep(consts.GameLogicSleep)
	}
}

func MinerUpdate(key glob.Position, o *glob.MObj) {
	matType := consts.ObjTypeCoal
	o.MContents[matType]++
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
