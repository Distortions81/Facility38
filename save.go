package main

import (
	"Facility38/cwlog"
	"Facility38/def"
	"Facility38/util"
	"Facility38/world"
	"bytes"
	"encoding/json"
	"os"
	"runtime"
	"time"
)

/* Used to munge data into a test save file */
/* TODO: SAVE VERSION AND MAP SEED INTO FILE */
type gameSave struct {
	Version  int
	Date     int64
	MapSeeds MapSeedsData
	Objects  []*saveMObj
}

type MapSeedsData struct {
	Grass  int64
	Oil    int64
	Gas    int64
	Coal   int64
	Iron   int64
	Copper int64
	Stone  int64
}

type saveMObj struct {
	Pos      world.XYs                       `json:"p,omitempty"`
	TypeI    uint8                           `json:"i,omitempty"`
	Dir      uint8                           `json:"d,omitempty"`
	Contents *world.MaterialContentsType     `json:"c,omitempty"`
	KGFuel   float32                         `json:"kf,omitempty"`
	KGHeld   float32                         `json:"k,omitempty"`
	Ports    [def.DIR_MAX]*world.ObjPortData `json:"po,omitempty"`
	Ticks    uint8                           `json:"t,omitempty"`
}

/* WIP */
func SaveGame() {
	defer util.ReportPanic("SaveGame")

	if world.WASMMode {
		return
	}

	go func() {
		defer util.ReportPanic("SaveGame goroutine")
		GameLock.Lock()
		defer GameLock.Unlock()

		tempPath := "saves/save.dat.tmp"
		finalPath := "saves/save.dat"
		os.Mkdir("saves", 0666)

		start := time.Now()
		cwlog.DoLog(true, "Save starting.")

		/* Pause the whole world ... */
		world.SuperChunkListLock.RLock()
		world.TickListLock.Lock()
		world.TockListLock.Lock()

		var seeds MapSeedsData
		for _, nl := range NoiseLayers {
			switch nl.TypeI {
			case def.MAT_NONE:
				seeds.Grass = nl.RandomSeed
			case def.MAT_OIL:
				seeds.Oil = nl.RandomSeed
			case def.MAT_GAS:
				seeds.Gas = nl.RandomSeed
			case def.MAT_COAL:
				seeds.Coal = nl.RandomSeed
			case def.MAT_IRON_ORE:
				seeds.Iron = nl.RandomSeed
			case def.MAT_COPPER_SHOT:
				seeds.Copper = nl.RandomSeed
			case def.MAT_STONE_BLOCK:
				seeds.Stone = nl.RandomSeed
			}
		}

		tempList := gameSave{
			Version:  2,
			Date:     time.Now().Unix(),
			MapSeeds: seeds}
		for _, sChunk := range world.SuperChunkList {
			for _, chunk := range sChunk.ChunkList {
				for _, mObj := range chunk.ObjList {
					tobj := &saveMObj{
						Pos:      util.CenterXY(mObj.Pos),
						TypeI:    mObj.Unique.TypeP.TypeI,
						Dir:      mObj.Dir,
						Contents: mObj.Unique.Contents,
						KGFuel:   mObj.Unique.KGFuel,
						KGHeld:   mObj.KGHeld,
					}

					/* Convert pointer to type int */
					for c := range tobj.Contents.Mats {
						if tobj.Contents.Mats[c] == nil {
							continue
						}
						if tobj.Contents.Mats[c].TypeP == nil {
							continue
						}
					}

					/* Convert pointer to type int */
					for p := range tobj.Ports {
						if tobj.Ports[p] == nil {
							continue
						}
						if tobj.Ports[p].Buf.TypeP == nil {
							continue
						}
						tobj.Ports[p].Obj = nil
					}
					tempList.Objects = append(tempList.Objects, tobj)
				}
			}
		}
		cwlog.DoLog(true, "WALK COMPLETE:", time.Since(start).String())

		b, _ := json.Marshal(tempList)

		world.SuperChunkListLock.RUnlock()
		world.TickListLock.Unlock()
		world.TockListLock.Unlock()
		cwlog.DoLog(true, "ENCODE DONE (WORLD UNLOCKED):", time.Since(start).String())

		_, err := os.Create(tempPath)

		if err != nil {
			cwlog.DoLog(true, "SaveGame: os.Create error: %v\n", err)
			return
		}

		zip := util.CompressZip(b)

		err = os.WriteFile(tempPath, zip, 0644)

		if err != nil {
			cwlog.DoLog(true, "SaveGame: os.WriteFile error: %v\n", err)
		}

		err = os.Rename(tempPath, finalPath)

		if err != nil {
			cwlog.DoLog(true, "SaveGame: couldn't rename save file: %v\n", err)
			return
		}

		util.ChatDetailed("Game save complete: "+finalPath, world.ColorOrange, time.Second*15)

		cwlog.DoLog(true, "COMPRESS & WRITE COMPLETE:", time.Since(start).String())
	}()
}

/* WIP */
func LoadGame() {
	defer util.ReportPanic("LoadGame")
	util.Chat("Load is current disabled (wip).")
	return

	if world.WASMMode {
		return
	}

	go func() {
		defer util.ReportPanic("LoadGame goroutine")
		GameLock.Lock()
		defer GameLock.Unlock()

		//start := time.Now()

		b, err := os.ReadFile("save.dat")
		if err != nil {
			cwlog.DoLog(true, "LoadGame: file not found: %v\n", err)
			return
		}

		unzip := util.UncompressZip(b)
		dbuf := bytes.NewBuffer(unzip)

		dec := json.NewDecoder(dbuf)

		NukeWorld()

		/* Pause the whole world ... */
		world.SuperChunkListLock.RLock()
		world.TickListLock.Lock()
		world.TockListLock.Lock()
		tempList := gameSave{}
		err = dec.Decode(&tempList)
		if err != nil {
			cwlog.DoLog(true, "LoadGame: JSON decode error: %v\n", err)
			return
		}

		if tempList.Version != 2 {
			cwlog.DoLog(true, "LoadGame: Invalid save version.")
		}

		world.SuperChunkListLock.RUnlock()
		for n, nl := range NoiseLayers {
			switch nl.TypeI {
			case def.MAT_NONE:
				NoiseLayers[n].RandomSeed = tempList.MapSeeds.Grass
			case def.MAT_OIL:
				NoiseLayers[n].RandomSeed = tempList.MapSeeds.Oil
			case def.MAT_GAS:
				NoiseLayers[n].RandomSeed = tempList.MapSeeds.Gas
			case def.MAT_COAL:
				NoiseLayers[n].RandomSeed = tempList.MapSeeds.Coal
			case def.MAT_IRON_ORE:
				NoiseLayers[n].RandomSeed = tempList.MapSeeds.Iron
			case def.MAT_COPPER_SHOT:
				NoiseLayers[n].RandomSeed = tempList.MapSeeds.Copper
			case def.MAT_STONE_BLOCK:
				NoiseLayers[n].RandomSeed = tempList.MapSeeds.Stone
			}
		}

		ResourceMapInit()

		/* Needs unsafeCreateObj that can accept a starting data set */
		count := 0
		for i := range tempList.Objects {

			obj := &world.ObjData{
				Pos: util.UnCenterXY(tempList.Objects[i].Pos),
				Unique: &world.UniqueObject{
					TypeP:    WorldObjs[tempList.Objects[i].TypeI],
					Contents: tempList.Objects[i].Contents,
					KGFuel:   tempList.Objects[i].KGFuel,
				},
				Dir:    tempList.Objects[i].Dir,
				KGHeld: tempList.Objects[i].KGHeld,
			}

			for c := range obj.Unique.Contents.Mats {
				if obj.Unique.Contents.Mats[c] == nil {
					continue
				}
			}

			/* Relink */
			MakeChunk(util.UnCenterXY(tempList.Objects[i].Pos))
			chunk := util.GetChunk(util.UnCenterXY(tempList.Objects[i].Pos))
			obj.Chunk = chunk

			obj.Chunk.BuildingMap[util.UnCenterXY(tempList.Objects[i].Pos)].Obj = obj
			obj.Chunk.ObjList = append(obj.Chunk.ObjList, obj)
			chunk.Parent.PixmapDirty = true
			chunk.NumObjs++

			if obj.Unique.TypeP.InitObj != nil {
				obj.Unique.TypeP.InitObj(obj)
			}

			chunk.Parent.PixmapDirty = true

			count++
		}

		world.VisDataDirty.Store(true)

		world.TickListLock.Unlock()
		world.TockListLock.Unlock()

		util.ChatDetailed("Game load complete.", world.ColorOrange, time.Second*15)
	}()
}

func NukeWorld() {
	defer util.ReportPanic("NukeWorld")
	if world.TockCount == 0 && world.TickCount == 0 {
		return
	}

	world.TickListLock.Lock()
	world.TickList = []world.TickEvent{}
	world.TickCount = 0
	world.TickListLock.Unlock()

	world.TockListLock.Lock()
	world.TockList = []world.TickEvent{}
	world.TockCount = 0
	world.TockListLock.Unlock()

	world.EventQueueLock.Lock()
	world.EventQueue = []*world.EventQueueData{}
	world.EventQueueLock.Unlock()

	world.ObjQueueLock.Lock()
	world.ObjQueue = []*world.ObjectQueueData{}
	world.ObjQueueLock.Unlock()

	world.SuperChunkListLock.Lock()

	/* Erase current map */
	for sc, superchunk := range world.SuperChunkList {
		for c, chunk := range superchunk.ChunkList {

			world.SuperChunkList[sc].ChunkList[c].Parent = nil
			if chunk.TerrainImage != nil && chunk.TerrainImage != world.TempChunkImage && !chunk.UsingTemporary {
				world.SuperChunkList[sc].ChunkList[c].TerrainImage.Dispose()
			}

			for o, obj := range chunk.ObjList {
				world.SuperChunkList[sc].ChunkList[c].ObjList[o].Chunk = nil
				for p := range obj.Ports {
					world.SuperChunkList[sc].ChunkList[c].ObjList[o].Ports[p].Obj = nil
				}
				world.SuperChunkList[sc].ChunkList[c].ObjList[o] = nil
			}

			world.SuperChunkList[sc].ChunkList[c].ObjList = nil
			world.SuperChunkList[sc].ChunkList[c].BuildingMap = nil
		}
		world.SuperChunkList[sc].ChunkList = nil
		world.SuperChunkList[sc].ChunkMap = nil
		if world.SuperChunkList[sc].PixelMap != nil {
			world.SuperChunkList[sc].PixelMap.Dispose()
			world.SuperChunkList[sc].PixelMap = nil
		}
		world.SuperChunkList[sc].ResourceMap = nil
	}
	world.SuperChunkList = []*world.MapSuperChunk{}
	world.SuperChunkMap = make(map[world.XY]*world.MapSuperChunk)

	world.VisDataDirty.Store(true)
	world.ZoomScale = def.DefaultZoom

	TickIntervals = []TickInterval{}

	runtime.GC()
	world.SuperChunkListLock.Unlock()
}
