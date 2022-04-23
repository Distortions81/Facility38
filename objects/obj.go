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
	TockList []glob.TickEvent

	ObjectHitlist []*glob.ObjectHitlistData
	EventHitlist  []*glob.EventHitlistData
)

func TickTockLoop() {
	lastUpdate := time.Time{}
	start := time.Now()

	for {
		lastUpdate = start
		start = time.Now()

		/* Calculate real frame time and adjust */
		glob.MeasuredObjectUPS_ns = start.Sub(lastUpdate) //Used for animation tweening

		WorldTick++

		runTocks() //Process objects
		runTicks() //Tick objects
		runEventHitlist()
		runObjectHitlist()

		//If there is time left, sleep
		frameTook := time.Since(start)
		sleepFor := glob.ObjectUPS_ns - frameTook
		time.Sleep(sleepFor)
	}
}

func TickObj(o *glob.WObject) {

	if o.OutputObj != nil && o.OutputObj.Valid && o.OutputObj.InputBuffer[o].Amount == 0 {
		o.OutputObj.InputBuffer[o] = o.OutputBuffer
		o.OutputBuffer = &glob.MatData{}

		fmt.Println(o.TypeP.Name, o.OutputObj.InputBuffer[o].Amount, "Output: ", o.OutputObj.TypeP.Name)
	} else {
		o.OutputObj = nil
	}
}

func MinerUpdate(o *glob.WObject) {

	if o.OutputObj != nil && o.OutputObj.Valid {
		input := uint64((o.TypeP.MinerKGSec * consts.TIMESCALE) / float64(o.TypeP.ProcessInterval))

		o.OutputBuffer.Amount += input
		o.OutputBuffer.TypeI = consts.MAT_COAL
		o.OutputBuffer.TypeP = *MatTypes[consts.MAT_COAL]
		o.OutputBuffer.TweenStamp = time.Now()

		fmt.Println("Miner: ", o.TypeP.Name, " output: ", input)
	} else {
		o.OutputObj = nil
	}
}

func SmelterUpdate(obj *glob.WObject) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func IronCasterUpdate(obj *glob.WObject) {
	//oData := glob.GameObjTypes[Obj.Type]

}

func BeltUpdate(obj *glob.WObject) {
	if obj.OutputBuffer.Amount == 0 {
		for src, mat := range obj.InputBuffer {
			if mat.Amount > 0 {
				mat.TweenStamp = time.Now()
				obj.OutputBuffer = mat
				obj.InputBuffer[src] = &glob.MatData{}
				fmt.Println(obj.TypeP.Name, " moved: ", mat.Amount)
			}
		}
	}

}

func SteamEngineUpdate(obj *glob.WObject) {
}

func BoxUpdate(obj *glob.WObject) {
	for src, mat := range obj.InputBuffer {
		if mat.Amount > 0 {
			src.Contents[mat.TypeI] = mat
			obj.InputBuffer[src] = &glob.MatData{}
			fmt.Println(MatTypes[mat.TypeI].Name, " input: ", mat.Amount)
		}
	}
}

//Move materials from one object to another
func runTicks() {
	numWorkers := runtime.NumCPU()
	wg := sizedwaitgroup.New(numWorkers)

	l := len(TickList) - 1
	if l < 1 {
		return
	}
	each := (l / numWorkers)
	p := 0

	if each < 1 {
		each = l + 1
		numWorkers = 1
	} else {
		fmt.Println("runTicks: ", l, " objects", each, " each")
	}
	for n := 0; n < numWorkers; n++ {
		//Handle remainder on last worker
		if n == numWorkers-1 {
			each = l + 1 - p
		}

		wg.Add()
		go func(start int, end int) {
			for i := start; i < end; i++ {
				TickObj(TickList[i].Target)
			}
			wg.Done()
		}(p, p+each)
		p += each

	}
	wg.Wait()
}

//Process objects
func runTocks() {
	numWorkers := runtime.NumCPU()
	wg := sizedwaitgroup.New(numWorkers)

	l := len(TockList) - 1
	if l < 1 {
		return
	}
	each := (l / numWorkers)
	p := 0

	if each < 1 {
		each = l + 1
		numWorkers = 1
	} else {
		fmt.Println("runTocks: ", l, " objects", each, " each")
	}
	for n := 0; n < numWorkers; n++ {
		//Handle remainder on last worker
		if n == numWorkers-1 {
			each = l + 1 - p
		}

		wg.Add()
		go func(start int, end int) {
			for i := start; i < end; i++ {
				TockList[i].Target.TypeP.UpdateObj(TockList[i].Target)
			}
			wg.Done()
		}(p, p+each)
		p += each

	}
	wg.Wait()
}

func ticklistAdd(target *glob.WObject) {
	TickList = append(TickList, glob.TickEvent{Target: target})
}

func tockListAdd(target *glob.WObject) {
	TockList = append(TockList, glob.TickEvent{Target: target})
}

func eventHitlistAdd(obj *glob.WObject, qtype int, delete bool) {
	for _, e := range EventHitlist {
		if e.QType == qtype && e.Obj == obj {
			fmt.Println("eventHitlistAdd:", obj.TypeP.Name, "already in list")
			return
		}
	}
	EventHitlist = append(EventHitlist, &glob.EventHitlistData{Obj: obj, QType: qtype, Delete: delete})
}

func ticklistRemove(obj *glob.WObject) {

	for i, e := range TickList {
		if e.Target == obj {

			if len(TockList) > 1 {
				TickList = append(TickList[:i], TickList[i+1:]...)
			} else {
				TickList = []glob.TickEvent{}
			}
			break
		}
	}
}

func tocklistRemove(obj *glob.WObject) {

	for i, e := range TockList {
		if e.Target == obj {

			if len(TockList) > 1 {
				TockList = append(TockList[:i], TockList[i+1:]...)
			} else {
				TockList = []glob.TickEvent{}
			}
			break
		}
	}
}

func LinkObj(pos glob.Position, obj *glob.WObject) {

	//Link output
	if obj.TypeP.HasMatOutput {
		fmt.Println("pos", pos, "output dir: ", util.DirToName(obj.OutputDir))
		destObj := util.GetNeighborObj(obj, pos, obj.OutputDir)

		if destObj != nil {
			obj.OutputObj = destObj
			destObj.InputBuffer[obj] = &glob.MatData{}
			eventHitlistAdd(obj, consts.QUEUE_TYPE_TICK, false)
			eventHitlistAdd(obj, consts.QUEUE_TYPE_TOCK, false)
			eventHitlistAdd(destObj, consts.QUEUE_TYPE_TICK, false)
			eventHitlistAdd(destObj, consts.QUEUE_TYPE_TOCK, false)
			fmt.Println("Linked object output: ", obj.TypeP.Name, " to: ", destObj.TypeP.Name)
		}
	}

	//Link inputs
	var i int
	found := false
	for i = consts.DIR_NORTH; i <= consts.DIR_WEST; i++ {
		if obj.TypeP.HasMatOutput && i == obj.OutputDir {
			continue
		}
		neigh := util.GetNeighborObj(obj, pos, i)
		if neigh != nil {
			if !found {
				neigh.OutputObj = obj
				obj.InputBuffer[neigh] = &glob.MatData{}
				eventHitlistAdd(obj, consts.QUEUE_TYPE_TICK, false)
				eventHitlistAdd(obj, consts.QUEUE_TYPE_TOCK, false)
				eventHitlistAdd(neigh, consts.QUEUE_TYPE_TICK, false)
				eventHitlistAdd(neigh, consts.QUEUE_TYPE_TOCK, false)
				fmt.Println("Linked object output: ", neigh.TypeP.Name, " to: ", obj.TypeP.Name)
				break
			}
		}
	}

}

func CreateObj(pos glob.Position, mtype int) *glob.WObject {

	//Make chunk if needed
	chunk := util.GetChunk(&pos)
	if chunk == nil {
		cpos := util.PosToChunkPos(&pos)
		fmt.Println("Made chunk:", cpos)

		chunk = &glob.MapChunk{}
		glob.WorldMap[cpos] = chunk
		chunk.WObject = make(map[glob.Position]*glob.WObject)
	}

	obj := chunk.WObject[pos]

	if obj != nil {
		fmt.Println("Object already exists at:", pos)
		return nil
	}

	obj = &glob.WObject{}

	obj.TypeP = GameObjTypes[mtype]

	obj.OutputObj = nil

	obj.Contents = [consts.MAT_MAX]*glob.MatData{}
	obj.InputBuffer = make(map[*glob.WObject]*glob.MatData)
	obj.OutputBuffer = &glob.MatData{}

	obj.OutputDir = consts.DIR_EAST
	obj.Valid = true

	//Put in chunk map
	glob.WorldMap[util.PosToChunkPos(&pos)].WObject[pos] = obj
	fmt.Println("Made obj:", pos, obj.TypeP.Name)
	LinkObj(pos, obj)

	return obj
}

func ObjectHitlistAdd(obj *glob.WObject, otype int, pos *glob.Position, delete bool) {
	if delete {
		eventHitlistAdd(obj, consts.QUEUE_TYPE_TICK, true)
		eventHitlistAdd(obj, consts.QUEUE_TYPE_TOCK, true)
	}
	ObjectHitlist = append(ObjectHitlist, &glob.ObjectHitlistData{Obj: obj, OType: otype, Pos: pos, Delete: delete})
}

func runEventHitlist() {
	for _, e := range EventHitlist {
		if e.Delete {
			switch e.QType {
			case consts.QUEUE_TYPE_TICK:
				ticklistRemove(e.Obj)
			case consts.QUEUE_TYPE_TOCK:
				tocklistRemove(e.Obj)
			}
		} else {
			switch e.QType {
			case consts.QUEUE_TYPE_TICK:
				ticklistAdd(e.Obj)
			case consts.QUEUE_TYPE_TOCK:
				tockListAdd(e.Obj)
			}
		}
	}
	EventHitlist = []*glob.EventHitlistData{}
}

func runObjectHitlist() {

	for _, item := range ObjectHitlist {
		if item.Delete {
			if item.Obj != nil {
				//Delete
				if item.Obj.Valid {
					item.Obj.Valid = false
					fmt.Println("Deleted:", item.Obj.TypeP.Name)
				}
			}
			delete(glob.WorldMap[util.PosToChunkPos(item.Pos)].WObject, *item.Pos)

		} else {
			//Add
			obj := CreateObj(*item.Pos, item.OType)
			if obj != nil {
				fmt.Println("Added:", obj.TypeP.Name)
			}
		}
	}
	ObjectHitlist = []*glob.ObjectHitlistData{}
}
