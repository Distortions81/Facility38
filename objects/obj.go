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
	lastUpdate := time.Time{}
	start := time.Now()

	for {
		lastUpdate = start
		start = time.Now()

		/* Calculate real frame time and adjust */
		glob.RealUPS_ns = start.Sub(lastUpdate) //Used for animation tweening

		glob.WorldMapUpdateLock.Lock()

		WorldTick++
		RunObjOutputs() //Send to other objects
		RunProcs()      //Process objects

		glob.WorldMapUpdateLock.Unlock()

		//If there is time left, sleep
		frameTook := time.Since(start)
		sleepFor := glob.GameLogicRate_ns - frameTook
		time.Sleep(sleepFor)
	}
}

func MinerUpdate(o *glob.MObj) {

	input := uint64((o.TypeP.MinerKGSec * consts.TIMESCALE) / float64(o.TypeP.ProcessInterval))

	/* Temporary for testing */

	if o.Contains[consts.MAT_COAL] == nil {
		o.Contains[consts.MAT_COAL] = &glob.MatData{}
	}
	o.Contains[consts.MAT_COAL].TypeP = MatTypes[consts.MAT_COAL]
	/* Temporary for testing */

	if o.Contains[consts.MAT_COAL].Amount < o.TypeP.CapacityKG {
		o.Contains[consts.MAT_COAL].Amount += input
	}

	fmt.Println("Miner", o.TypeP.Name, "retrieved", o.Contains[consts.MAT_COAL].Amount, "coal")
	util.MoveMateriaslOut(o)
}

func SmelterUpdate(obj *glob.MObj) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func IronCasterUpdate(obj *glob.MObj) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func BeltUpdate(obj *glob.MObj) {
	util.MoveMaterialsAlong(obj)
}

func SteamEngineUpdate(obj *glob.MObj) {
}

func BoxUpdate(obj *glob.MObj) {
	util.MoveMaterialsIn(obj)
}

//Send external to other objects
func RunObjOutputs() {
	//wg := sizedwaitgroup.New(runtime.NumCPU())

	for p, event := range TickList {
		if !event.Target.Valid {
			RemoveTickQue(p)
			fmt.Println("Deleted eternal tick event for invalid object")
			continue
		}
		if event.Target != nil {
			if event.Target.OutputObj != nil {
				if !event.Target.OutputObj.Valid {
					fmt.Println("Deleted OutputObj for invalid object.")
					event.Target.OutputObj = nil
					continue
				}
				util.OutputMaterial(event.Target)
			}
		}
	}
}

//Process objects
func RunProcs() {
	found := false
	count := 0

	//Processes these every tick
	for key, event := range ProcList[0] {
		count++
		if event.Target.Valid {
			//fmt.Println("Processed", event.Target.TypeP.Name)
			event.Target.TypeP.ObjUpdate(event.Target)
		} else {
			//Delete eternal events if object was invalidated
			if len(ProcList[0]) > 1 {
				ProcList[0] = append(ProcList[0][:key], ProcList[0][key+1:]...)
			} else {
				delete(ProcList, 0)
			}
			fmt.Println("Deleted eternal proc event for invalid object")
		}
	}

	//Process these at specific intervals
	for _, event := range ProcList[WorldTick] {
		count++
		found = true

		//Process
		if event.Target.Valid {
			event.Target.TypeP.ObjUpdate(event.Target)

			//fmt.Println("Processed", event.Target.TypeP.Name)
			ToProcQue(event.Target, WorldTick+event.Target.TypeP.ProcessInterval)
		}
	}
	if found {
		//fmt.Println("Deleted procs for ", WorldTick)
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
	if len(TickList) > 1 {
		TickList = append(TickList[:pos], TickList[pos+1:]...)
	} else {
		TickList = nil
	}
}

func RemoveTockQue(pos int) {
	if len(TockList) > 1 {
		TockList = append(TockList[:pos], TockList[pos+1:]...)
	} else {
		TockList = nil
	}
}

func RemoveProcQue(tick uint64, pos int) {
	if len(ProcList[tick]) > 1 {
		ProcList[tick] = append(ProcList[tick][:pos], ProcList[tick][pos+1:]...)
	} else {
		delete(ProcList, tick)
	}
}

func LinkObj(pos glob.Position, obj *glob.MObj) {

	//Link output
	if obj.OutputDir > 0 && obj.TypeP.HasOutput {
		fmt.Println("pos", pos, "output dir: ", obj.OutputDir)
		destObj := util.GetNeighborObj(obj, pos, obj.OutputDir)

		if destObj != nil {
			obj.OutputObj = destObj
			ToTickQue(obj)
			fmt.Println("Linked object output: ", obj.TypeP.Name, " to: ", destObj.TypeP.Name)
		} else {
			//fmt.Println("Unable to find object to link to.")
		}
	}

	//Link inputs
	var i int
	found := false
	for i = consts.DIR_NORTH; i <= consts.DIR_WEST; i++ {
		if obj.TypeP.HasOutput && i == obj.OutputDir {
			continue
		}
		neigh := util.GetNeighborObj(obj, pos, i)
		if neigh != nil {
			if !found {
				neigh.OutputObj = obj
				ToTickQue(neigh)
				fmt.Println("Linked object output: ", neigh.TypeP.Name, " to: ", obj.TypeP.Name)
				break
			}
		} else {
			//fmt.Println("Unable to find object to reverse link to.", pos)
		}
	}

}

func MakeMObj(pos glob.Position, mtype int) *glob.MObj {

	//Make chunk if needed
	chunk := util.GetChunk(&pos)
	if chunk == nil {
		cpos := util.PosToChunkPos(&pos)
		fmt.Println("Made chunk:", cpos)

		chunk = &glob.MapChunk{}
		glob.WorldMap[cpos] = chunk
		chunk.MObj = make(map[glob.Position]*glob.MObj)
	}

	obj := chunk.MObj[pos]

	if obj != nil {
		fmt.Println("Object already exists at:", pos)
		return nil
	}

	obj = &glob.MObj{}

	obj.TypeP = GameObjTypes[mtype]

	obj.OutputObj = nil
	obj.OutputBuffer = [consts.MAT_MAX]*glob.MatData{}

	obj.Contains = [consts.MAT_MAX]*glob.MatData{}

	obj.InputBuffer = make(map[*glob.MObj]*[consts.MAT_MAX]*glob.MatData)

	obj.OutputDir = consts.DIR_EAST
	obj.Valid = true

	//Put in chunk map
	glob.WorldMap[util.PosToChunkPos(&pos)].MObj[pos] = obj
	fmt.Println("Made obj:", pos, obj.TypeP.Name)
	LinkObj(pos, obj)

	if obj.TypeP.ObjUpdate != nil {
		if obj.TypeP.ProcessInterval > 0 {
			//Process on a specifc ticks
			ToProcQue(obj, WorldTick+1+obj.TypeP.ProcessInterval)
		} else {
			//Eternal
			ToProcQue(obj, 0)
		}
		fmt.Println("Added proc event for:", obj.TypeP.Name)
	}

	return obj
}

func DeleteMObj(obj *glob.MObj, pos *glob.Position) {
	if obj == nil || !obj.Valid {
		fmt.Println("DeleteMObj: NIL or invalid object supplied.", pos)
		return
	}

	chunk := util.GetChunk(pos)
	if chunk == nil {
		fmt.Println("DeleteMObj: No chunk found for: ", pos)
		return
	}

	//Delete object
	obj.Valid = false
	delete(chunk.MObj, *pos)
	fmt.Println("Object deleted:", pos, obj.TypeP.Name)

	//Delete chunk if empty
	if len(chunk.MObj) <= 0 {
		cpos := util.PosToChunkPos(pos)
		fmt.Println("Chunk deleted:", cpos)
		delete(glob.WorldMap, cpos)
	}
}
