package obj

import (
	"GameTest/consts"
	"GameTest/glob"
	"fmt"
	"time"
)

var (
	WorldEpoch       time.Time
	WorldTick        uint64 = 0
	CurrentWorldStep int
	TickList         []glob.TickEvent
	TockList         []glob.TickEvent
	ProcList         map[uint64][]glob.TickEvent
	AddToWorld       []*glob.MObj
	DelFromWorld     []*glob.MObj
)

func GLogic() {

	lastUpdate := time.Now()
	WorldEpoch = time.Now()

	for {

		if time.Since(lastUpdate) > glob.GameLogicRate {
			glob.WorldMapUpdateLock.Lock()

			start := time.Now()

			WorldTick++
			RunTicks()
			//RunMods()
			RunTocks()
			RunProcs()

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

func LoaderUpdate(obj *glob.MObj) {
	//oData := glob.GameObjTypes[Obj.Type]

}

//Send to external
func RunTicks() {
	//wg := sizedwaitgroup.New(runtime.NumCPU())

	for _, event := range TickList {
		for dir, o := range event.Target.SendTo {

			if o.Contents[dir].Amount > 0 && len(o.SendTo) > 0 {
				//Send to object
				if o.SendTo[dir].External[dir].Amount == 0 {
					o.SendTo[dir].External[dir] = o.Contents[dir]

					fmt.Println("Sent ", o.External[dir].Amount, " to ", o.SendTo[dir].External[dir].TypeP.Name)
					o.Contents[dir].Amount = 0
				}
			}
		}
	}
}

//Move internal and process
func RunTocks() {

	for _, event := range TockList {

		if len(event.Target.Contents) == 0 {
			continue
		}
		//Move internal
		for dir, o := range event.Target.Contents {

			if o.Amount > 0 {
				event.Target.Contents[dir] = o

				fmt.Println("Got ", o.Amount, " to ", o.TypeP.Name)
				o.Amount = 0
			}
		}
	}
}

func RunProcs() {
	found := false
	count := 0

	//Processes these every tick
	for _, event := range ProcList[0] {
		count++
		if event.Target.Valid {
			event.Target.TypeP.ObjUpdate(event.Target)
		}
	}

	//Process these at specific intervals
	for _, event := range ProcList[WorldTick] {
		count++
		//Process
		if event.Target.Valid {
			event.Target.TypeP.ObjUpdate(event.Target)

			AddProcQ(event.Target, WorldTick+uint64(event.Target.TypeP.ProcSeconds*float64(glob.LogicUPS)))
			found = true
		}
	}
	if found {
		//fmt.Println("Deleted procs for ", WorldTick)
		delete(ProcList, WorldTick)
	}

	//fmt.Println("Count: ", count)
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

func AddTickQ(target *glob.MObj) {
	TickList = append(TickList, glob.TickEvent{Target: target})
}

func AddTockQ(target *glob.MObj) {
	TockList = append(TockList, glob.TickEvent{Target: target})
}

func AddProcQ(target *glob.MObj, tick uint64) {
	ProcList[tick] = append(ProcList[tick], glob.TickEvent{Target: target})
}
