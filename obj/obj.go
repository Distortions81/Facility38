package obj

import (
	"GameTest/consts"
	"GameTest/glob"
	"runtime"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

func GLogic() {

	lastUpdate := time.Now()
	var ticks uint64 = 0

	for {

		if time.Since(lastUpdate) > consts.GameLogicRate {
			wg := sizedwaitgroup.New(runtime.NumCPU())
			ticks++
			//fmt.Println("Tick:", ticks)

			glob.WorldMapLock.RLock()
			for _, item := range glob.WorldMap {
				for okey, o := range item.MObj {
					wg.Add()
					go func(okey glob.Position, o *glob.MObj) {

						oData := GameObjTypes[o.Type]
						if oData.ObjUpdate != nil {
							if time.Since(o.LastUpdate) > oData.UpdateInterval {
								o.LastUpdate = time.Now()
								oData.ObjUpdate(okey, o)
							}
						}
						wg.Done()
					}(okey, o)
				}
			}
			wg.Wait()
			glob.WorldMapLock.RUnlock()
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
