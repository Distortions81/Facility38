package objects

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/util"
	"fmt"
	"time"
)

var (
	WorldTick uint64 = 0

	TickList []glob.TickEvent
	TockList []glob.TickEvent
	ProcList map[uint64][]glob.TickEvent

	AddToWorld   []*glob.MObj
	DelFromWorld []*glob.MObj
)

func GLogic() {

	lastUpdate := time.Now()

	for {

		if time.Since(lastUpdate) > glob.GameLogicRate {
			glob.WorldMapUpdateLock.Lock()

			start := time.Now()

			WorldTick++
			RunTicks() //Move external and send to objects
			//RunMods()
			RunTocks() //Move internal
			RunProcs() //Process objects

			glob.WorldMapUpdateLock.Unlock()
			lastUpdate = time.Now()
			glob.UpdateTook = time.Since(start)
			//fmt.Println("Update budget used: ", (float64(glob.UpdateTook.Microseconds()/1000.0)/250.0)*100.0, "%")
		}

		//Reduce busy waiting
		time.Sleep(glob.GameLogicSleep)
	}
}

func MinerUpdate(o *glob.MObj) {

	input := int(float64(o.TypeP.MinerKGSec*consts.TIMESCALE) / o.TypeP.ProcSeconds)
	//if o.Contents[consts.DIR_INTERNAL].Amount+input < o.TypeP.CapacityKG {
	o.Contents[consts.DIR_INTERNAL].Amount += input
	//}
}

func SmelterUpdate(obj *glob.MObj) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func IronCasterUpdate(obj *glob.MObj) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func BeltUpdate(obj *glob.MObj) {
}

func BoxUpdate(obj *glob.MObj) {
	for dir, v := range obj.External {
		if v != nil {
			if obj.Contents[dir] != nil {
				obj.Contents[dir].Type = v.Type
				obj.Contents[dir].TypeP = v.TypeP
				obj.Contents[dir].Amount += v.Amount
			} else {
				obj.Contents[dir] = &glob.MatData{Type: v.Type, Amount: v.Amount}
			}
			v.Amount = 0
		}
	}
}

//Send external to other objects
func RunTicks() {
	//wg := sizedwaitgroup.New(runtime.NumCPU())

	for _, event := range TickList {
		if event.Target != nil {
			for dir, dest := range event.Target.SendTo {
				if dest != nil {
					util.MoveMaterialToObj(event.Target, dest, dir)
				}
			}
		}
	}
}

//Move internal to external
func RunTocks() {

	for _, event := range TockList {

		if len(event.Target.Contents) == 0 {
			continue
		}
		//Move internal
		for dir, o := range event.Target.Contents {
			if o != nil {
				if o.Amount > 0 {
					util.MoveMaterialOut(event.Target, dir, o)

				}
			}
		}
	}
}

func RunProcs() {
	found := false
	count := 0

	//Processes these every tick
	for key, event := range ProcList[0] {
		count++
		if event.Target.Valid {
			event.Target.TypeP.ObjUpdate(event.Target)
		} else {
			//Delete eternal events if object was invalidated
			ProcList[0] = append(ProcList[0][:key], ProcList[0][key+1:]...)
		}
	}

	//Process these at specific intervals
	for _, event := range ProcList[WorldTick] {
		count++
		//Process
		if event.Target.Valid {
			event.Target.TypeP.ObjUpdate(event.Target)

			ToProcQue(event.Target, WorldTick+uint64(event.Target.TypeP.ProcSeconds*float64(glob.LogicUPS)))
			found = true
		}
	}
	if found {
		fmt.Println("Deleted procs for ", WorldTick)
		delete(ProcList, WorldTick)
	}
}

func ToTickQue(target *glob.MObj) {
	TickList = append(TickList, glob.TickEvent{Target: target})
}

func ToTockQue(target *glob.MObj) {
	TockList = append(TockList, glob.TickEvent{Target: target})
}

func ToProcQue(target *glob.MObj, tick uint64) {
	ProcList[tick] = append(ProcList[tick], glob.TickEvent{Target: target})
}

func LinkAll() {
	for _, chunk := range glob.WorldMap {
		for pos, obj := range chunk.MObj {
			if obj.OutputDir > 0 {
				destObj := util.GetNeighborObj(pos, obj.OutputDir)
				obj.SendTo[obj.OutputDir] = destObj
			}
		}
	}
}

func LinkObj(pos glob.Position, obj *glob.MObj) {
	if obj.OutputDir > 0 {
		destObj := util.GetNeighborObj(pos, obj.OutputDir)
		if destObj != nil {
			obj.SendTo[obj.OutputDir] = destObj
			fmt.Println("Linked object: ", obj.Type, " to: ", destObj.Type)
		}
	} else {
		for i := consts.DIR_NORTH; i <= consts.DIR_WEST; i++ {
			neigh := util.GetNeighborObj(pos, i)
			if neigh != nil {
				neigh.SendTo[i] = obj
				fmt.Println("Linked object REVERSE: ", obj.Type, " to: ", neigh.Type)
			}
		}
	}
}
