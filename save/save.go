package save

import (
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/util"
	"bytes"
	"encoding/json"
	"os"
)

func SaveGame() {

	tempPath := "save.dat.tmp"
	finalPath := "save.dat"

	tempList := []*glob.SaveMObj{}
	glob.SuperChunkMapLock.Lock()
	for _, sChunk := range glob.SuperChunkMap {
		for _, chunk := range sChunk.ChunkMap {
			for pos, mObj := range chunk.ObjMap {
				tempList = append(tempList, &glob.SaveMObj{O: mObj, P: pos})
			}
		}
	}
	glob.SuperChunkMapLock.Unlock()

	b, err := json.MarshalIndent(tempList, "", "\t")

	if err != nil {
		cwlog.DoLog("SaveGame: JSON encode error: %v", err)
		return
	}

	_, err = os.Create(tempPath)

	if err != nil {
		cwlog.DoLog("SaveGame: os.Create error: %v", err)
		return
	}

	zip := util.CompressZip(b)

	err = os.WriteFile(tempPath, zip, 0644)

	if err != nil {
		cwlog.DoLog("SaveGame: os.WriteFile error: %v", err)
	}

	err = os.Rename(tempPath, finalPath)

	if err != nil {
		cwlog.DoLog("SaveGame: couldn't rename save file: %v", err)
		return
	}
}

func LoadGame() {
	b, err := os.ReadFile("save.dat")
	if err != nil {
		cwlog.DoLog("LoadGame: file not found: %v", err)
		return
	}
	data := util.UncompressZip(b)
	dbuf := bytes.NewBuffer(data)

	dec := json.NewDecoder(dbuf)
	err = dec.Decode(&glob.SuperChunkMap)
	if err != nil {
		cwlog.DoLog("LoadGame: JSON decode error: %v", err)
		return
	}

}
