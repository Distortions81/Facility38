package obj

import (
	"GameTest/consts"
	"GameTest/glob"
	"fmt"
	"time"
)

var (
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

	for {

		if time.Since(lastUpdate) > consts.GameLogicRate {
			RunTicks()
			//RunMods()
			RunTocks()
			RunProcs()
			WorldTick++

			lastUpdate = time.Now()
		}

		//Reduce busy waiting
		time.Sleep(consts.GameLogicSleep)
	}
}

func MinerUpdate(key glob.Position, o *glob.MObj) {
	o.Contents[consts.DIR_INTERNAL].Type = consts.MAT_IRONORE
	o.Contents[consts.DIR_INTERNAL].Amount = o.Contents[consts.DIR_INTERNAL].Amount + 1
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

					fmt.Println("Sent ", o.External[dir].Amount, " to ", o.SendTo[dir].External[dir].Type)
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

				fmt.Println("Sent ", o.Amount, " to ", o.Type)
				o.Amount = 0
			}
		}
	}
}

func RunProcs() {
	found := false
	for _, event := range ProcList[WorldTick] {
		//Process
		if event.Target.Valid {
			GameObjTypes[event.Target.Type].ObjUpdate(event.Key, event.Target)
			AddProcQ(event.Key, event.Target, WorldTick+GameObjTypes[event.Target.Type].ProcInterval)
			found = true
		}
	}
	if found {
		fmt.Println("Delete procs for ", WorldTick)
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
