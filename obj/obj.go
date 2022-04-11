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
			WorldTick++
			RunTicks()
			//RunMods()
			RunTocks()
			RunProcs()

			lastUpdate = time.Now()
		}

		//Reduce busy waiting
		time.Sleep(glob.GameLogicSleep)
	}
}

func MinerUpdate(key glob.Position, o *glob.MObj) {

	/* Temporary for testing */
	o.Contents[consts.DIR_INTERNAL].Type = consts.MAT_COAL
	o.Contents[consts.DIR_INTERNAL].TypeP = MatTypes[consts.MAT_COAL]
	/* Temporary for testing */

	o.Contents[consts.DIR_INTERNAL].Amount += ((float64(time.Since(o.LastUpdate).Milliseconds()) / consts.TIMESCALE) / 1000.0) * o.TypeP.MinerProductPerSecond
	o.LastUpdate = time.Now()
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

	//Processes these every tick
	for _, event := range ProcList[0] {
		if event.Target.Valid {
			event.Target.TypeP.ObjUpdate(event.Key, event.Target)
		}
	}

	//Process these at specific intervals
	for _, event := range ProcList[WorldTick] {
		//Process
		if event.Target.Valid {
			event.Target.TypeP.ObjUpdate(event.Key, event.Target)

			AddProcQ(event.Key, event.Target, WorldTick+1+event.Target.TypeP.ProcInterval)
			found = true
		}
	}
	if found {
		//fmt.Println("Deleted procs for ", WorldTick)
		delete(ProcList, WorldTick)
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

func AddTickQ(key glob.Position, target *glob.MObj) {
	fmt.Println("Adding tick for ", key)
	TickList = append(TickList, glob.TickEvent{Target: target, Key: key})
}

func AddTockQ(key glob.Position, target *glob.MObj) {
	fmt.Println("Adding tock for ", key)
	TockList = append(TockList, glob.TickEvent{Target: target, Key: key})
}

func AddProcQ(key glob.Position, target *glob.MObj, tick uint64) {
	fmt.Println("Adding proc for ", key, " at ", tick)
	ProcList[tick] = append(ProcList[tick], glob.TickEvent{Target: target, Key: key})
}
