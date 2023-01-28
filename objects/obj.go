package objects

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/noise"
	"GameTest/util"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/remeh/sizedwaitgroup"
)

var (
	gWorldTick uint64 = 0

	ListLock sync.Mutex
	TickList []glob.TickEvent
	TockList []glob.TickEvent

	ObjectHitlist []*glob.ObjectHitlistData
	EventHitlist  []*glob.EventHitlistData

	gTickCount     int
	gTockCount     int
	gTickWorkSize  int
	gNumChunkImage int

	TockWorkSize int
	NumWorkers   int

	wg sizedwaitgroup.SizedWaitGroup
)

const (
	cChunkGroundCacheTime = time.Second * 15
	cCacheMax             = 100
)

func TickTockDaemon() {
	var start time.Time
	wg = sizedwaitgroup.New(NumWorkers)

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

		gWorldTick++
		gTickWorkSize = gTickCount / NumWorkers
		TockWorkSize = gTockCount / NumWorkers

		ListLock.Lock()
		runTocks()         //Process objects
		runTicks()         //Move external
		runEventHitlist()  //Queue to add/remove events
		runObjectHitlist() //Queue to add/remove objects
		ListLock.Unlock()

		if !consts.UPSBench {
			sleepFor := glob.ObjectUPS_ns - time.Since(start)
			time.Sleep(sleepFor)
		}
		glob.MeasuredObjectUPS_ns = time.Since(start)
	}

	//pprof.StopCPUProfile()

}

func killTerrainCache(chunk *glob.MapChunk) {
	if chunk.UsingTemporary || chunk.GroundImg == nil {
		return
	}
	chunk.GroundLock.Lock()
	chunk.GroundImg.Dispose()
	chunk.GroundImg = nil
	gNumChunkImage--
	chunk.GroundLock.Unlock()
}

func TerrainCacheDaemon() {
	time.Sleep(time.Second)
	wg := sizedwaitgroup.New(NumWorkers)

	for {
		glob.WorldMapLock.Lock()
		tmpWorld := glob.WorldMap
		glob.WorldMapLock.Unlock()

		/* If we zoom out, decallocate everything */
		if glob.ZoomScale < consts.MiniMapLevel {
			for _, chunk := range tmpWorld {
				killTerrainCache(chunk)
			}
			continue
		}

		for cpos, chunk := range tmpWorld {
			if chunk.GroundImg == nil {
				continue
			}
			wg.Add()
			go func(chunk *glob.MapChunk, cpos glob.XY) {
				glob.WorldMapLock.Lock()

				if chunk.Visible || chunk.GroundImg == nil {
					if chunk.UsingTemporary {
						RenderChunkGround(chunk, true, cpos)
					}
				} else {
					if gNumChunkImage > cCacheMax &&
						time.Since(chunk.LastSaw) > cChunkGroundCacheTime {
						killTerrainCache(chunk)
					}
				}

				glob.WorldMapLock.Unlock()
				wg.Done()
			}(chunk, cpos)
		}
		wg.Wait()
		time.Sleep(time.Millisecond * 100)
	}
}

func tickObj(o *glob.WObject) {

	if o.OutputObj != nil {
		revDir := util.ReverseDirection(o.Direction)
		if o.OutputObj.InputBuffer[revDir] != nil && o.OutputObj.InputBuffer[revDir].Amount == 0 &&
			o.OutputBuffer.Amount > 0 {

			o.OutputObj.InputBuffer[revDir].Amount = o.OutputBuffer.Amount
			o.OutputObj.InputBuffer[revDir].TypeP = o.OutputBuffer.TypeP

			o.OutputBuffer.Amount = 0
		}
	}
}

// Move materials from one object to another
func runTicks() {

	l := gTickCount - 1
	if l < 1 {
		return
	}

	numWorkers := l / gTickWorkSize
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
				tickObj(TickList[i].Target)
			}
			wg.Done()
		}(p, p+each)
		p += each

	}
	wg.Wait()
}

// Process objects
func runTocks() {

	l := gTockCount - 1
	if l < 1 {
		return
	}

	numWorkers := l / TockWorkSize
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
				TockList[i].Target.TypeP.UpdateObj(TockList[i].Target)
			}
			wg.Done()
		}(p, p+each, tickNow)
		p += each

	}
	wg.Wait()
}

func ticklistAdd(target *glob.WObject) {
	TickList = append(TickList, glob.TickEvent{Target: target})
	gTickCount++
}

func tockListAdd(target *glob.WObject) {
	TockList = append(TockList, glob.TickEvent{Target: target})
	gTockCount++
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
	gTickCount--
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
	gTockCount--
}

func linkOut(pos glob.XY, obj *glob.WObject, dir int) {
	destObj := util.GetNeighborObj(obj, pos, dir)

	if destObj != nil {
		obj.OutputObj = destObj
		destObj.InputBuffer[util.ReverseDirection(dir)] = &glob.MatData{}
	}
}

func LinkObj(pos glob.XY, obj *glob.WObject) {

	//Link inputs
	var i int

	for i = consts.DIR_NORTH; i <= consts.DIR_NONE; i++ {
		neigh := util.GetNeighborObj(obj, pos, i)

		if neigh != nil {
			if neigh.TypeP.HasMatOutput && util.ReverseDirection(neigh.Direction) == i {
				neigh.OutputObj = obj
				obj.InputBuffer[i] = &glob.MatData{}
			}
		}
	}

	//Link output
	if obj.TypeP.HasMatOutput {
		linkOut(pos, obj, obj.Direction)

		//Link up additonal outputs for splitters
		if obj.TypeI == consts.ObjTypeBasicSplit {
			dir := util.RotCW(obj.Direction)
			linkOut(pos, obj, dir)

			dir = util.RotCW(dir)
			linkOut(pos, obj, dir)

			dir = util.RotCW(dir)
			linkOut(pos, obj, dir)

		}
	}

}

func MakeChunk(pos glob.XY) {
	//Make chunk if needed

	newPos := pos
	chunk := util.GetChunk(&newPos)
	if chunk == nil {
		cpos := util.PosToChunkPos(&newPos)
		//fmt.Println("Made chunk:", cpos)

		glob.WorldMapLock.Lock()

		chunk = &glob.MapChunk{}
		glob.WorldMap[cpos] = chunk
		chunk.WObject = make(map[glob.XY]*glob.WObject)
		glob.CameraDirty = true

		glob.WorldMapLock.Unlock()
	}
}

func RenderChunkGround(chunk *glob.MapChunk, doDetail bool, cpos glob.XY) {
	/* Make optimized background */
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

	chunkPix := (consts.SpriteScale * consts.ChunkSize)

	bg := TerrainTypes[0].Image
	sx := int(float64(bg.Bounds().Size().X))
	sy := int(float64(bg.Bounds().Size().Y))
	var tImg *ebiten.Image

	if sx > 0 && sy > 0 {

		chunk.GroundLock.Lock()
		tImg = ebiten.NewImage(chunkPix, chunkPix)
		gNumChunkImage++
		chunk.UsingTemporary = false
		chunk.GroundLock.Unlock()

		for i := 0; i < consts.ChunkSize; i++ {
			for j := 0; j < consts.ChunkSize; j++ {
				op.GeoM.Reset()
				op.GeoM.Translate(float64(i*sx), float64(j*sy))

				if doDetail {
					x := (float64(cpos.X*consts.ChunkSize) + float64(i))
					y := (float64(cpos.Y*consts.ChunkSize) + float64(j))
					h := noise.NoiseMap(x, y)

					//fmt.Printf("%.2f,%.2f: %.2f\n", x, y, h)
					op.ColorM.Reset()
					op.ColorM.Scale(h*2, 1, 1, 1)
				}

				tImg.DrawImage(bg, op)
			}
		}

	} else {
		panic("No valid bg texture.")
	}
	chunk.GroundLock.Lock()
	chunk.GroundImg = tImg
	chunk.GroundLock.Unlock()
}

func ExploreMap(input int) {
	/* Explore some map */

	area := input * consts.ChunkSize
	offs := int(consts.XYCenter)
	for x := -area; x < area; x += consts.ChunkSize {
		for y := -area; y < area; y += consts.ChunkSize {
			pos := glob.XY{X: offs - x, Y: offs - y}

			MakeChunk(pos)
		}
	}
}

func CreateObj(pos glob.XY, mtype int, dir int) *glob.WObject {

	//Make chunk if needed
	MakeChunk(pos)
	chunk := util.GetChunk(&pos)

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
	obj.OutputBuffer = &glob.MatData{}
	obj.Direction = dir

	if obj.TypeP.HasMatOutput {
		EventHitlistAdd(obj, consts.QUEUE_TYPE_TICK, false)
	}
	/* Only add to list if the object calls an update function */
	if obj.TypeP.UpdateObj != nil {
		EventHitlistAdd(obj, consts.QUEUE_TYPE_TOCK, false)
	}

	//Put in chunk map
	glob.WorldMap[util.PosToChunkPos(&pos)].WObject[pos] = obj
	//fmt.Println("Made obj:", pos, obj.TypeP.Name)
	LinkObj(pos, obj)

	return obj
}

func ObjectHitlistAdd(obj *glob.WObject, otype int, pos *glob.XY, delete bool, dir int) {
	ObjectHitlist = append(ObjectHitlist, &glob.ObjectHitlistData{Obj: obj, OType: otype, Pos: pos, Delete: delete, Dir: dir})
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
			EventHitlistAdd(item.Obj, consts.QUEUE_TYPE_TICK, true)
			EventHitlistAdd(item.Obj, consts.QUEUE_TYPE_TOCK, true)

			glob.WorldMapLock.Lock()
			delete(glob.WorldMap[util.PosToChunkPos(item.Pos)].WObject, *item.Pos)
			glob.WorldMapLock.Unlock()

		} else {
			//Add
			CreateObj(*item.Pos, item.OType, item.Dir)
		}
	}
	ObjectHitlist = []*glob.ObjectHitlistData{}
}
