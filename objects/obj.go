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
			RunTicks() //Send to other objects
			//RunTocks()
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
	/* Temporary for testing */
	if o.Contains[consts.MAT_COAL] == nil {
		o.Contains[consts.MAT_COAL] = &glob.MatData{}
	}
	o.Contains[consts.MAT_COAL].Type = consts.MAT_COAL
	o.Contains[consts.MAT_COAL].TypeP = MatTypes[consts.MAT_COAL]
	/* Temporary for testing */

	o.Contains[consts.MAT_COAL].Amount += input

	util.MoveMaterialOut(o)
}

func SmelterUpdate(obj *glob.MObj) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func IronCasterUpdate(obj *glob.MObj) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func BeltUpdate(obj *glob.MObj) {
}

func SteamEngineUpdate(obj *glob.MObj) {
}

func BoxUpdate(obj *glob.MObj) {
	for _, v := range obj.External {
		for mtype, m := range v {
			if m == nil {
				continue
			}
			if obj.Contains[mtype] != nil {
				obj.Contains[mtype].Type = m.Type
				obj.Contains[mtype].TypeP = m.TypeP
				obj.Contains[mtype].Amount += m.Amount
			} else {
				obj.Contains[mtype] = &glob.MatData{Type: mtype, Amount: m.Amount}
			}
			m.Amount = 0
		}
	}
}

//Send external to other objects
func RunTicks() {
	//wg := sizedwaitgroup.New(runtime.NumCPU())

	for p, event := range TickList {
		if !event.Target.Valid {
			RemoveTickQue(p)
			fmt.Println("Deleted eternal tick event for invalid object")
			continue
		}
		if event.Target != nil {
			for dir, dest := range event.Target.SendTo {
				if dest != nil {
					if !dest.Valid {
						event.Target.SendTo[dir] = nil
						fmt.Println("Deleted SendTo for invalid object.")
						continue
					}
					util.MoveMaterialToObj(event.Target, dest, dir)
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

func RemoveTickQue(pos int) {
	TickList = append(TickList[:pos], TickList[pos+1:]...)
}

func RemoveTockQue(pos int) {
	TockList = append(TockList[:pos], TockList[pos+1:]...)
}

func RemoveProcQue(tick uint64, pos int) {
	ProcList[tick] = append(ProcList[tick][:pos], ProcList[tick][pos+1:]...)
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
		fmt.Println("pos", pos, "output dir: ", obj.OutputDir)
		destObj := util.GetNeighborObj(pos, obj.OutputDir)
		if destObj != nil {
			obj.SendTo[obj.OutputDir] = destObj
			fmt.Println("Linked object: ", obj.Type, " to: ", destObj.Type)
		} else {
			fmt.Println("Unable to find object to link to.")
		}
	}

	for i := consts.DIR_NORTH; i <= consts.DIR_WEST; i++ {
		neigh := util.GetNeighborObj(pos, i)
		if neigh != nil {
			neigh.SendTo[util.ReverseDirection(i)] = obj
			fmt.Println("Linked object REVERSE: ", obj.Type, " to: ", neigh.Type)
		} else {
			fmt.Println("Unable to find object to reverse link to.")
		}
	}

}
