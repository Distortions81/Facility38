package objects

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/util"
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
	start := time.Time{}

	/*var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	} */

	for {
		start = time.Now()

		WorldTick++

		runTocks() //Process objects
		runTicks() //Move external
		runEventHitlist()
		runObjectHitlist()

		//sleepFor := glob.ObjectUPS_ns - time.Since(start)
		//time.Sleep(sleepFor)
		glob.MeasuredObjectUPS_ns = time.Since(start)
	}

	//pprof.StopCPUProfile()

}

func TickObj(o *glob.WObject) {

	if o.OutputObj != nil && o.OutputObj.Valid {
		if o.OutputObj.InputBuffer[o].Amount == 0 &&
			o.OutputBuffer.Amount > 0 {

			o.OutputObj.InputBuffer[o].Amount = o.OutputBuffer.Amount
			o.OutputObj.InputBuffer[o].TypeI = o.OutputBuffer.TypeI
			o.OutputObj.InputBuffer[o].TypeP = o.OutputBuffer.TypeP
			o.OutputObj.InputBuffer[o].TweenStamp = o.OutputBuffer.TweenStamp

			o.OutputBuffer.Amount = 0
		}
	}
}

// Move materials from one object to another
func runTicks() {
	numWorkers := glob.NumWorkers
	wg := sizedwaitgroup.New(numWorkers)

	l := len(TickList) - 1
	if l < 1 {
		return
	}

	numWorkers = l / consts.WorkSize
	if numWorkers < 1 {
		numWorkers = 1
	}
	each := (l / numWorkers)
	p := 0

	if each < 1 {
		each = l + 1
		numWorkers = 1
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

// Process objects
func runTocks() {
	numWorkers := glob.NumWorkers
	wg := sizedwaitgroup.New(numWorkers)

	l := len(TockList) - 1
	if l < 1 {
		return
	}

	numWorkers = l / consts.WorkSize
	if numWorkers < 1 {
		numWorkers = 1
	}
	each := (l / numWorkers)
	p := 0

	if each < 1 {
		each = l + 1
		numWorkers = 1
	}
	tickNow := time.Now()
	for n := 0; n < numWorkers; n++ {
		//Handle remainder on last worker
		if n == numWorkers-1 {
			each = l + 1 - p
		}

		wg.Add()
		go func(start int, end int, tickNow time.Time) {
			for i := start; i < end; i++ {
				TockList[i].Target.TypeP.UpdateObj(TockList[i].Target, tickNow)
			}
			wg.Done()
		}(p, p+each, tickNow)
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

func EventHitlistAdd(obj *glob.WObject, qtype int, delete bool) {
	EventHitlist = append(EventHitlist, &glob.EventHitlistData{Obj: obj, QType: qtype, Delete: delete})
}

func ticklistRemove(obj *glob.WObject) {

	for i, e := range TickList {
		if e.Target == obj {

			if len(TickList) > 1 {
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
		//fmt.Println("pos", pos, "output dir: ", util.DirToName(obj.OutputDir))
		destObj := util.GetNeighborObj(obj, pos, obj.OutputDir)

		if destObj != nil {
			obj.OutputObj = destObj
			destObj.InputBuffer[obj] = &glob.MatData{}
			//fmt.Println("Linked object output: ", obj.TypeP.Name, " to: ", destObj.TypeP.Name)
		}
	}

	//Link inputs
	var i int

	obj.BeltStart = true
	for i = consts.DIR_NORTH; i <= consts.DIR_WEST; i++ {
		neigh := util.GetNeighborObj(obj, pos, i)

		if neigh != nil {
			if neigh.TypeP.HasMatOutput && util.ReverseDirection(neigh.OutputDir) == i {
				neigh.OutputObj = obj
				obj.InputBuffer[neigh] = &glob.MatData{}
				//fmt.Println("Linked object output: ", neigh.TypeP.Name, " to: ", obj.TypeP.Name)
				if neigh.TypeI == consts.ObjTypeBasicBelt || neigh.TypeI == consts.ObjTypeBasicBeltVert {
					obj.BeltStart = false
				}
			}
		}
	}

}

func CreateObj(pos glob.Position, mtype int) *glob.WObject {

	//Make chunk if needed
	chunk := util.GetChunk(&pos)
	if chunk == nil {
		cpos := util.PosToChunkPos(&pos)
		//fmt.Println("Made chunk:", cpos)

		chunk = &glob.MapChunk{}
		glob.WorldMap[cpos] = chunk
		chunk.WObject = make(map[glob.Position]*glob.WObject)
	}

	obj := chunk.WObject[pos]

	if obj != nil {
		//fmt.Println("Object already exists at:", pos)
		return nil
	}

	obj = &glob.WObject{}

	obj.TypeP = GameObjTypes[mtype]
	obj.TypeI = mtype

	obj.OutputObj = nil

	obj.Contents = [consts.MAT_MAX]*glob.MatData{}
	obj.InputBuffer = make(map[*glob.WObject]*glob.MatData)
	obj.OutputBuffer = &glob.MatData{}

	obj.OutputDir = consts.DIR_EAST
	obj.Valid = true

	EventHitlistAdd(obj, consts.QUEUE_TYPE_TICK, false)
	EventHitlistAdd(obj, consts.QUEUE_TYPE_TOCK, false)

	//Put in chunk map
	glob.WorldMap[util.PosToChunkPos(&pos)].WObject[pos] = obj
	//fmt.Println("Made obj:", pos, obj.TypeP.Name)
	LinkObj(pos, obj)

	return obj
}

func ObjectHitlistAdd(obj *glob.WObject, otype int, pos *glob.Position, delete bool) {
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
				item.Obj.Valid = false
			}
			EventHitlistAdd(item.Obj, consts.QUEUE_TYPE_TICK, true)
			EventHitlistAdd(item.Obj, consts.QUEUE_TYPE_TOCK, true)

			delete(glob.WorldMap[util.PosToChunkPos(item.Pos)].WObject, *item.Pos)

		} else {
			//Add
			CreateObj(*item.Pos, item.OType)
		}
	}
	ObjectHitlist = []*glob.ObjectHitlistData{}
}
