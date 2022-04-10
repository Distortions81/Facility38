package main

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
				for okey, obj := range item.MObj {
					UpdateObject(okey, obj)
				}
			}
		}

		//Reduce busy waiting
		time.Sleep(consts.GameLogicSleep)
	}
}

func UpdateObject(Key glob.Position, Obj *glob.MObj) {
	oData := glob.GameObjTypes[Obj.Type]

	if oData.ObjUpdate != nil {
		oData.ObjUpdate(Key, Obj)
	}
}

func MinerUpdate(Key glob.Position, Obj *glob.MObj) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func SmelterUpdate(Key glob.Position, Obj *glob.MObj) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func IronMineUpdate(Key glob.Position, Obj *glob.MObj) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func LoaderUpdate(Key glob.Position, Obj *glob.MObj) {
	//oData := glob.GameObjTypes[Obj.Type]

}
