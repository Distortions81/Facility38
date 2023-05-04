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
		gameLock.Lock()
		defer gameLock.Unlock()

		saveDate := time.Now().UTC().Unix()

		savesDir := "saves"
		saveTempName := fmt.Sprintf("save-%v.zip.tmp", saveDate)
		saveName := fmt.Sprintf("save-%v.zip", saveDate)

		os.Mkdir(savesDir, os.ModePerm)

		start := time.Now()
		chat("Saving...")

		/* Pause the whole world ... */
		superChunkListLock.RLock()
		tickListLock.Lock()
		tockListLock.Lock()

		lastSave = time.Now().UTC()

		tempList := gameSave{
			Version:    2,
			Date:       lastSave.UTC().Unix(),
			MapSeed:    MapSeed,
			GameTicks:  GameTick,
			CameraPos:  XYs{X: int32(cameraX), Y: int32(cameraY)},
			CameraZoom: int16(zoomScale)}
		for _, sChunk := range superChunkList {
			for _, chunk := range sChunk.chunkList {
				for _, mObj := range chunk.objList {
					tmpObj := &saveMObj{
						Pos:   CenterXY(mObj.Pos),
						TypeI: mObj.Unique.typeP.typeI,
						Dir:   mObj.Dir,
					}

					tempList.Objects = append(tempList.Objects, tmpObj)
				}
			}
		}
		doLog(true, "WALK COMPLETE: %v", time.Since(start).String())

		b, _ := json.Marshal(tempList)

		superChunkListLock.RUnlock()
		tickListLock.Unlock()
		tockListLock.Unlock()
		doLog(true, "ENCODE DONE (WORLD UNLOCKED): %v", time.Since(start).String())

		if !wasmMode {
			_, err := os.Create(savesDir + "/" + saveTempName)

			if err != nil {
				doLog(true, "SaveGame: os.Create error: %v\n", err)
				chatDetailed("Unable to write to saves directory (check file permissions)", ColorOrange, time.Second*5)
				return
			}
		}

		zip := CompressZip(b)

		if wasmMode {
			// Call the SendBytes function with the data and filename
			go sendBytes(saveName, zip)
		} else {
			err := os.WriteFile(savesDir+"/"+saveTempName, zip, 0644)

			if err != nil {
				doLog(true, "SaveGame: os.WriteFile error: %v\n", err)
				chatDetailed("Unable to write to saves directory (check file permissions)", ColorOrange, time.Second*5)
			}

			err = os.Rename(savesDir+"/"+saveTempName, savesDir+"/"+saveName)

			if err != nil {
				doLog(true, "SaveGame: couldn't rename save file: %v\n", err)
				chatDetailed("Unable to write to saves directory (check file permissions)", ColorOrange, time.Second*5)
				return
			}

			chatDetailed("Game save complete: "+saveName, ColorOrange, time.Second*5)

			doLog(true, "COMPRESS & WRITE COMPLETE: %v", time.Since(start).String())
		}

		chat("Save complete.")
	}()
}

func findNewestSave() string {

	dir := "saves/"
	files, err := os.ReadDir(dir)
	if err != nil {
		doLog(true, "Saves folder not found.")
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
	if wasmMode {
		chat("To load a save game, click 'Choose File' in the top-left of the screen and select the save game file to load.")
	}
	loadGame(false, nil)
}
func loadGame(external bool, data []byte) {
	defer reportPanic("LoadGame")

	if wasmMode && !external {
		return
	}
	if wasmMode && external && len(data) == 0 {
		chat("No save data found.")
		return
	}
	lastSave = time.Now().UTC()

	go func(external bool, data []byte) {

		if !checkAuth() {
			return
		}
		mapLoadPercent = 0

		defer reportPanic("LoadGame goroutine")
		gameLock.Lock()
		defer gameLock.Unlock()

		saveName := "browser attachment"

		if !external {
			saveName = findNewestSave()
			if saveName == "" {
				chat("No saves found!")
				statusText = ""
				return
			}
		}

		statusText = fmt.Sprintf("Reading file: %v\n", saveName)
		mapGenerated.Store(false)
		defer mapGenerated.Store(true)

		chat("Loading saves/" + saveName)

		var b []byte
		var err error
		if !external {
			b, err = os.ReadFile("saves/" + saveName)
			if err != nil {
				doLog(true, "LoadGame: file not found: %v\n", err)
				statusText = ""
				return
			}
		} else {
			b = data
		}

		unzip := UncompressZip(b)
		deCompBuf := bytes.NewBuffer(unzip)
		dec := json.NewDecoder(deCompBuf)

		statusText = "Clearing memory.\n"
		nukeWorld()
		statusText = "Parsing save file.\n"

		/* Pause the whole world ... */
		superChunkListLock.Lock()
		tickListLock.Lock()
		tockListLock.Lock()
		tempList := gameSave{}
		err = dec.Decode(&tempList)
		if err != nil {
			doLog(true, "LoadGame: JSON decode error: %v\n", err)
			statusText = ""
			return
		}

		if tempList.Version != 2 {
			doLog(true, "LoadGame: Invalid save version.")
			statusText = ""
			return
		}

		superChunkListLock.Unlock()
		MapSeed = tempList.MapSeed
		statusText = "Generating map resources.\n"
		resourceMapInit()

		statusText = "Loading objects.\n"
		/* Needs unsafeCreateObj that can accept a starting data set */
		numObj := len(tempList.Objects)
		for i := range tempList.Objects {
			if i%10000 == 0 {
				mapLoadPercent = float32(float32(i)/float32(numObj)) * 100.0
				runEventQueue()
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

		lastSave = time.Unix(tempList.Date, 0).UTC()
		GameTick = tempList.GameTicks
		if tempList.CameraPos.X != 0 && tempList.CameraPos.Y != 0 {
			cameraX = float32(tempList.CameraPos.X)
			cameraY = float32(tempList.CameraPos.Y)
		}
		if tempList.CameraZoom != 0 {
			zoomScale = float32(tempList.CameraZoom)
		}

		visDataDirty.Store(true)

		resetChat()
		go chatDetailed("Load complete!", ColorOrange, time.Second*15)
		mapLoadPercent = 100

		time.Sleep(time.Second)
		mapGenerated.Store(true)
		statusText = ""
	}(external, data)
}

func nukeWorld() {
	defer reportPanic("NukeWorld")

	eventQueueLock.Lock()
	eventQueue = []*eventQueueData{}
	eventQueueLock.Unlock()

	objQueueLock.Lock()
	objQueue = []*objectQueueData{}
	objQueueLock.Unlock()

	GameTick = 0

	superChunkListLock.Lock()

	/* Erase current map */
	for sc, supChunk := range superChunkList {
		for c, chunk := range supChunk.chunkList {

			superChunkList[sc].chunkList[c].parent = nil
			if chunk.terrainImage != nil && chunk.terrainImage != TempChunkImage && !chunk.usingTemporary {
				superChunkList[sc].chunkList[c].terrainImage.Dispose()
			}

			for o, obj := range chunk.objList {
				superChunkList[sc].chunkList[c].objList[o].chunk = nil
				for p := range obj.Ports {
					superChunkList[sc].chunkList[c].objList[o].Ports[p].obj = nil
				}
				superChunkList[sc].chunkList[c].objList[o] = nil
			}

			superChunkList[sc].chunkList[c].objList = nil
			superChunkList[sc].chunkList[c].buildingMap = nil
		}
		superChunkList[sc].chunkList = nil
		superChunkList[sc].chunkMap = nil
		if superChunkList[sc].pixelMap != nil {
			superChunkList[sc].pixelMap.Dispose()
			superChunkList[sc].pixelMap = nil
		}
		superChunkList[sc].resourceMap = nil
	}
	superChunkList = []*mapSuperChunkData{}
	superChunkMap = make(map[XY]*mapSuperChunkData)

	visDataDirty.Store(true)
	zoomScale = defaultZoom

	tickIntervals = []tickInterval{}

	runtime.GC()
	lastSave = time.Now().UTC()
	superChunkListLock.Unlock()
}
