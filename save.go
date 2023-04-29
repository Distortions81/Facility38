package main

import (
	"Facility38/cwlog"
	"Facility38/def"
	"Facility38/util"
	"Facility38/world"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

/* Used to munge data into a test save file */
/* TODO: SAVE VERSION AND MAP SEED INTO FILE */
type gameSave struct {
	Version   int
	Date      int64
	MapSeed   int64
	GameTicks uint64
	Objects   []*saveMObj
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
	Pos      world.XYs                   `json:"p,omitempty"`
	TypeI    uint8                       `json:"i,omitempty"`
	Dir      uint8                       `json:"d,omitempty"`
	Contents *world.MaterialContentsType `json:"c,omitempty"`
	KGFuel   float32                     `json:"kf,omitempty"`
	KGHeld   float32                     `json:"k,omitempty"`
}

/* WIP */
func SaveGame() {
	defer util.ReportPanic("SaveGame")

	if world.WASMMode {
		return
	}

	if world.TickCount == 0 {
		return
	}

	go func() {
		defer util.ReportPanic("SaveGame goroutine")
		GameLock.Lock()
		defer GameLock.Unlock()

		savenum := time.Now().UTC().Unix()

		tempPath := fmt.Sprintf("saves/save-%v.json.tmp", savenum)
		finalPath := fmt.Sprintf("saves/save-%v.json", savenum)
		os.Mkdir("saves", os.ModePerm)

		start := time.Now()
		util.Chat("Saving...")

		/* Pause the whole world ... */
		world.SuperChunkListLock.RLock()
		world.TickListLock.Lock()
		world.TockListLock.Lock()

		world.LastSave = time.Now().UTC()

		tempList := gameSave{
			Version:   2,
			Date:      world.LastSave.UTC().Unix(),
			MapSeed:   world.MapSeed,
			GameTicks: GameTick}
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
					if tobj.Contents != nil {
						for c := range tobj.Contents.Mats {
							if tobj.Contents.Mats[c] == nil {
								continue
							}
							if tobj.Contents.Mats[c].TypeP == nil {
								continue
							}
						}
					}

					tempList.Objects = append(tempList.Objects, tobj)
				}
			}
		}
		cwlog.DoLog(true, "WALK COMPLETE: %v", time.Since(start).String())

		b, _ := json.Marshal(tempList)

		world.SuperChunkListLock.RUnlock()
		world.TickListLock.Unlock()
		world.TockListLock.Unlock()
		cwlog.DoLog(true, "ENCODE DONE (WORLD UNLOCKED): %v", time.Since(start).String())

		_, err := os.Create(tempPath)

		if err != nil {
			cwlog.DoLog(true, "SaveGame: os.Create error: %v\n", err)
			util.ChatDetailed("Unable to write to saves directory (check file permissions)", world.ColorOrange, time.Second*5)
			return
		}

		//zip := util.CompressZip(b)

		err = os.WriteFile(tempPath, b, 0644)

		if err != nil {
			cwlog.DoLog(true, "SaveGame: os.WriteFile error: %v\n", err)
			util.ChatDetailed("Unable to write to saves directory (check file permissions)", world.ColorOrange, time.Second*5)
		}

		err = os.Rename(tempPath, finalPath)

		if err != nil {
			cwlog.DoLog(true, "SaveGame: couldn't rename save file: %v\n", err)
			util.ChatDetailed("Unable to write to saves directory (check file permissions)", world.ColorOrange, time.Second*5)
			return
		}

		util.ChatDetailed("Game save complete: "+finalPath, world.ColorOrange, time.Second*5)

		cwlog.DoLog(true, "COMPRESS & WRITE COMPLETE: %v", time.Since(start).String())

		if time.Since(start) > time.Second*2 {
			util.Chat("Save complete.")
		}
	}()
}

func FindNewstSave() string {

	dir := "saves/"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var newestFile os.FileInfo
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			if newestFile == nil || file.ModTime().After(newestFile.ModTime()) {
				newestFile = file
			}
		}
	}

	if newestFile != nil {
		return newestFile.Name()
	} else {
		return ""
	}
}

/* WIP */
func LoadGame() {
	defer util.ReportPanic("LoadGame")

	if world.WASMMode {
		return
	}

	go func() {
		defer util.ReportPanic("LoadGame goroutine")
		GameLock.Lock()
		defer GameLock.Unlock()

		saveName := FindNewstSave()
		if saveName == "" {
			util.Chat("No saves found!")
			return
		}
		util.Chat("Loading saves/" + saveName)
		b, err := os.ReadFile("saves/" + saveName)
		if err != nil {
			cwlog.DoLog(true, "LoadGame: file not found: %v\n", err)
			return
		}

		world.MapGenerated.Store(false)

		//unzip := util.UncompressZip(b)
		dbuf := bytes.NewBuffer(b)

		dec := json.NewDecoder(dbuf)

		NukeWorld()

		/* Pause the whole world ... */
		world.SuperChunkListLock.Lock()
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

		world.SuperChunkListLock.Unlock()
		world.MapSeed = tempList.MapSeed
		ResourceMapInit()

		/* Needs unsafeCreateObj that can accept a starting data set */
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

			if obj.Unique != nil && obj.Unique.Contents != nil {
				for c := range obj.Unique.Contents.Mats {
					if obj.Unique.Contents.Mats[c] == nil {
						continue
					}
				}
			}

			newObj := PlaceObj(obj.Pos, obj.Unique.TypeP.TypeI, nil, obj.Dir, false)
			newObj.Unique = obj.Unique
			newObj.KGHeld = obj.KGHeld
		}

		world.LastSave = time.Unix(tempList.Date, 0).UTC()
		GameTick = tempList.GameTicks

		world.VisDataDirty.Store(true)

		world.TickListLock.Unlock()
		world.TockListLock.Unlock()

		util.ChatDetailed("Load complete!", world.ColorOrange, time.Second*15)
		time.Sleep(time.Second)
		world.MapGenerated.Store(true)
	}()
}

func NukeWorld() {
	defer util.ReportPanic("NukeWorld")

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

	GameTick = 0

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
	world.LastSave = time.Now().UTC()
	world.SuperChunkListLock.Unlock()
}
