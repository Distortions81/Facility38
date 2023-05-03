package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

/* Used to munge data into a test save file */
type gameSave struct {
	Version    int
	Date       int64
	MapSeed    int64
	GameTicks  uint64
	CameraPos  XYs
	CameraZoom int16
	Objects    []*saveMObj
}

type saveMObj struct {
	Pos   XYs   `json:"p,omitempty"`
	TypeI uint8 `json:"i,omitempty"`
	Dir   uint8 `json:"d,omitempty"`
}

/* WIP */
func saveGame() {
	defer reportPanic("saveGame")

	go func() {
		if !checkAuth() {
			return
		}

		defer reportPanic("saveGame goroutine")
		GameLock.Lock()
		defer GameLock.Unlock()

		savenum := time.Now().UTC().Unix()

		savesDir := "saves"
		saveTempName := fmt.Sprintf("save-%v.zip.tmp", savenum)
		saveName := fmt.Sprintf("save-%v.zip", savenum)

		os.Mkdir(savesDir, os.ModePerm)

		start := time.Now()
		Chat("Saving...")

		/* Pause the whole world ... */
		SuperChunkListLock.RLock()
		TickListLock.Lock()
		TockListLock.Lock()

		LastSave = time.Now().UTC()

		tempList := gameSave{
			Version:    2,
			Date:       LastSave.UTC().Unix(),
			MapSeed:    MapSeed,
			GameTicks:  GameTick,
			CameraPos:  XYs{X: int32(CameraX), Y: int32(CameraY)},
			CameraZoom: int16(ZoomScale)}
		for _, sChunk := range SuperChunkList {
			for _, chunk := range sChunk.chunkList {
				for _, mObj := range chunk.objList {
					tobj := &saveMObj{
						Pos:   CenterXY(mObj.Pos),
						TypeI: mObj.Unique.typeP.typeI,
						Dir:   mObj.Dir,
					}

					tempList.Objects = append(tempList.Objects, tobj)
				}
			}
		}
		DoLog(true, "WALK COMPLETE: %v", time.Since(start).String())

		b, _ := json.Marshal(tempList)

		SuperChunkListLock.RUnlock()
		TickListLock.Unlock()
		TockListLock.Unlock()
		DoLog(true, "ENCODE DONE (WORLD UNLOCKED): %v", time.Since(start).String())

		if !WASMMode {
			_, err := os.Create(savesDir + "/" + saveTempName)

			if err != nil {
				DoLog(true, "SaveGame: os.Create error: %v\n", err)
				ChatDetailed("Unable to write to saves directory (check file permissions)", ColorOrange, time.Second*5)
				return
			}
		}

		zip := CompressZip(b)

		if WASMMode {
			// Call the SendBytes function with the data and filename
			go SendBytes(saveName, zip)

			// Wait for incoming messages from the JavaScript side
			//<-make(chan struct{})
		} else {
			err := os.WriteFile(savesDir+"/"+saveTempName, zip, 0644)

			if err != nil {
				DoLog(true, "SaveGame: os.WriteFile error: %v\n", err)
				ChatDetailed("Unable to write to saves directory (check file permissions)", ColorOrange, time.Second*5)
			}

			err = os.Rename(savesDir+"/"+saveTempName, savesDir+"/"+saveName)

			if err != nil {
				DoLog(true, "SaveGame: couldn't rename save file: %v\n", err)
				ChatDetailed("Unable to write to saves directory (check file permissions)", ColorOrange, time.Second*5)
				return
			}

			ChatDetailed("Game save complete: "+saveName, ColorOrange, time.Second*5)

			DoLog(true, "COMPRESS & WRITE COMPLETE: %v", time.Since(start).String())
		}

		Chat("Save complete.")
	}()
}

func findNewstSave() string {

	dir := "saves/"
	files, err := os.ReadDir(dir)
	if err != nil {
		DoLog(true, "Saves folder not found.")
		return ""
	}

	var newestFile fs.FileInfo
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".zip" {
			fi, err := file.Info()
			if err != nil {
				continue
			}
			if newestFile == nil || fi.ModTime().After(newestFile.ModTime()) {
				newestFile = fi
			}
		}
	}

	if newestFile != nil {
		return newestFile.Name()
	} else {
		return ""
	}
}

func triggerLoad() {
	if WASMMode {
		Chat("To load a save game, click 'Choose File' in the top-left of the screen and select the save game file to load.")
	}
	loadGame(false, nil)
}
func loadGame(external bool, data []byte) {
	defer reportPanic("LoadGame")

	if WASMMode && !external {
		return
	}
	if WASMMode && external && len(data) == 0 {
		Chat("No save data found.")
		return
	}
	LastSave = time.Now().UTC()

	go func(external bool, data []byte) {

		if !checkAuth() {
			return
		}
		MapLoadPercent = 0

		defer reportPanic("LoadGame goroutine")
		GameLock.Lock()
		defer GameLock.Unlock()

		saveName := "browser attachment"

		if !external {
			saveName = findNewstSave()
			if saveName == "" {
				Chat("No saves found!")
				statusText = ""
				return
			}
		}

		statusText = fmt.Sprintf("Reading file: %v\n", saveName)
		MapGenerated.Store(false)
		defer MapGenerated.Store(true)

		Chat("Loading saves/" + saveName)

		var b []byte
		var err error
		if !external {
			b, err = os.ReadFile("saves/" + saveName)
			if err != nil {
				DoLog(true, "LoadGame: file not found: %v\n", err)
				statusText = ""
				return
			}
		} else {
			b = data
		}

		unzip := UncompressZip(b)
		dbuf := bytes.NewBuffer(unzip)
		dec := json.NewDecoder(dbuf)

		statusText = "Clearing memory.\n"
		nukeWorld()
		statusText = "Parsing save file.\n"

		/* Pause the whole world ... */
		SuperChunkListLock.Lock()
		TickListLock.Lock()
		TockListLock.Lock()
		tempList := gameSave{}
		err = dec.Decode(&tempList)
		if err != nil {
			DoLog(true, "LoadGame: JSON decode error: %v\n", err)
			statusText = ""
			return
		}

		if tempList.Version != 2 {
			DoLog(true, "LoadGame: Invalid save version.")
			statusText = ""
			return
		}

		SuperChunkListLock.Unlock()
		MapSeed = tempList.MapSeed
		statusText = "Generating map resources.\n"
		resourceMapInit()

		statusText = "Loading objects.\n"
		/* Needs unsafeCreateObj that can accept a starting data set */
		numObj := len(tempList.Objects)
		for i := range tempList.Objects {
			if i%10000 == 0 {
				MapLoadPercent = float32(float32(i)/float32(numObj)) * 100.0
				RunEventQueue()
			}

			obj := &ObjData{
				Pos: UnCenterXY(tempList.Objects[i].Pos),
				Unique: &UniqueObject{
					typeP: worldObjs[tempList.Objects[i].TypeI],
				},
				Dir: tempList.Objects[i].Dir,
			}

			newObj := placeObj(obj.Pos, obj.Unique.typeP.typeI, nil, obj.Dir, true)
			newObj.KGHeld = obj.KGHeld
		}
		statusText = "Complete!\n"

		LastSave = time.Unix(tempList.Date, 0).UTC()
		GameTick = tempList.GameTicks
		if tempList.CameraPos.X != 0 && tempList.CameraPos.Y != 0 {
			CameraX = float32(tempList.CameraPos.X)
			CameraY = float32(tempList.CameraPos.Y)
		}
		if tempList.CameraZoom != 0 {
			ZoomScale = float32(tempList.CameraZoom)
		}

		VisDataDirty.Store(true)

		resetChat()
		go ChatDetailed("Load complete!", ColorOrange, time.Second*15)
		MapLoadPercent = 100

		time.Sleep(time.Second)
		MapGenerated.Store(true)
		statusText = ""
	}(external, data)
}

func nukeWorld() {
	defer reportPanic("NukeWorld")

	EventQueueLock.Lock()
	EventQueue = []*eventQueueData{}
	EventQueueLock.Unlock()

	ObjQueueLock.Lock()
	ObjQueue = []*objectQueueData{}
	ObjQueueLock.Unlock()

	GameTick = 0

	SuperChunkListLock.Lock()

	/* Erase current map */
	for sc, superchunk := range SuperChunkList {
		for c, chunk := range superchunk.chunkList {

			SuperChunkList[sc].chunkList[c].parent = nil
			if chunk.terrainImage != nil && chunk.terrainImage != TempChunkImage && !chunk.usingTemporary {
				SuperChunkList[sc].chunkList[c].terrainImage.Dispose()
			}

			for o, obj := range chunk.objList {
				SuperChunkList[sc].chunkList[c].objList[o].chunk = nil
				for p := range obj.Ports {
					SuperChunkList[sc].chunkList[c].objList[o].Ports[p].obj = nil
				}
				SuperChunkList[sc].chunkList[c].objList[o] = nil
			}

			SuperChunkList[sc].chunkList[c].objList = nil
			SuperChunkList[sc].chunkList[c].buildingMap = nil
		}
		SuperChunkList[sc].chunkList = nil
		SuperChunkList[sc].chunkMap = nil
		if SuperChunkList[sc].pixelMap != nil {
			SuperChunkList[sc].pixelMap.Dispose()
			SuperChunkList[sc].pixelMap = nil
		}
		SuperChunkList[sc].resourceMap = nil
	}
	SuperChunkList = []*mapSuperChunkData{}
	SuperChunkMap = make(map[XY]*mapSuperChunkData)

	VisDataDirty.Store(true)
	ZoomScale = DefaultZoom

	tickIntervals = []tickInterval{}

	runtime.GC()
	LastSave = time.Now().UTC()
	SuperChunkListLock.Unlock()
}
