package objects

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/util"
	"fmt"
	"runtime"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

var (
	WorldTick uint64 = 0

	TickList []glob.TickEvent
	ProcList []glob.TickEvent

	AddRemoveObjList []*glob.QueAddRemoveObjData
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

		ProcessAddDelObjQue()
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

	//fmt.Println("Miner", o.TypeP.Name, "retrieved", o.Contains[consts.MAT_COAL].Amount, "coal")
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
	numWorkers := runtime.NumCPU()
	wg := sizedwaitgroup.New(numWorkers)

	l := len(TickList) - 1
	each := (l / numWorkers)
	p := 0

	//fmt.Println("RunObjOutputs: ", len, " objects", each, " each")
	for n := 0; n < numWorkers; n++ {
		wg.Add()
		go func(start int, end int) {

			for i := start; i < end; i++ {
				if TickList[i].Target != nil {
					if TickList[i].Target.OutputObj != nil {
						util.OutputMaterial(TickList[i].Target)
					}
				}

			}
			wg.Done()
		}(p, p+each)
		p += each
	}
	wg.Wait()
}

//Process objects
func RunProcs() {
	numWorkers := runtime.NumCPU()
	wg := sizedwaitgroup.New(numWorkers)

	l := len(ProcList) - 1
	each := (l / numWorkers)
	p := 0

	//fmt.Println("RunProcs: ", len, " objects", each, " each")
	for n := 0; n < numWorkers; n++ {
		wg.Add()
		go func(start int, end int) {
			for i := start; i < end; i++ {
				ProcList[i].Target.TypeP.ObjUpdate(ProcList[i].Target)
			}
			wg.Done()
		}(p, p+each)
		p += each
	}
	wg.Wait()
}

func ToTickQue(target *glob.MObj) {
	TickList = append(TickList, glob.TickEvent{Target: target})
}

func ToProcQue(target *glob.MObj) {
	ProcList = append(ProcList, glob.TickEvent{Target: target})
}

func RemoveTickQue(pos int) {
	if len(TickList) > 1 {
		TickList = append(TickList[:pos], TickList[pos+1:]...)
	} else {
		TickList = nil
	}
}

func RemoveProcQue(tick uint64, pos int) {
	if len(ProcList) > 1 {
		ProcList = append(ProcList[:pos], ProcList[pos+1:]...)
	} else {
		ProcList = []glob.TickEvent{}
	}
}

func LinkObj(pos glob.Position, obj *glob.MObj) {

	//Link output
	if obj.OutputDir > 0 && obj.TypeP.HasOutput {
		//fmt.Println("pos", pos, "output dir: ", obj.OutputDir)
		destObj := util.GetNeighborObj(obj, pos, obj.OutputDir)

		if destObj != nil {
			obj.OutputObj = destObj
			ToTickQue(obj)
			//fmt.Println("Linked object output: ", obj.TypeP.Name, " to: ", destObj.TypeP.Name)
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
				//fmt.Println("Linked object output: ", neigh.TypeP.Name, " to: ", obj.TypeP.Name)
				break
			}
		}
	}

}

func MakeMObj(pos glob.Position, mtype int) *glob.MObj {

	//Make chunk if needed
	chunk := util.GetChunk(&pos)
	if chunk == nil {
		cpos := util.PosToChunkPos(&pos)
		//fmt.Println("Made chunk:", cpos)

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
	//fmt.Println("Made obj:", pos, obj.TypeP.Name)
	LinkObj(pos, obj)

	if obj.TypeP.ObjUpdate != nil {
		ToProcQue(obj)
		//fmt.Println("Added proc event for:", obj.TypeP.Name)
	}

	return obj
}

func QueAddDelMObj(obj *glob.MObj, otype int, pos *glob.Position, delete bool) {
	AddRemoveObjList = append(AddRemoveObjList, &glob.QueAddRemoveObjData{Obj: obj, OType: otype, Pos: pos, Delete: delete})
}

func ProcessAddDelObjQue() {

	for _, item := range AddRemoveObjList {
		if item.Delete {
			if item.Obj != nil {
				//Delete
				if item.Obj.Valid {
					item.Obj.Valid = false
					fmt.Println("Deleted:", item.Obj.TypeP.Name)
				}
			}
			delete(glob.WorldMap[util.PosToChunkPos(item.Pos)].MObj, *item.Pos)

		} else {
			//Add
			obj := MakeMObj(*item.Pos, item.OType)
			if obj != nil {
				fmt.Println("Added:", obj.TypeP.Name)
			}
		}
	}
	AddRemoveObjList = []*glob.QueAddRemoveObjData{}
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
