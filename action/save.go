package action

import (
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/util"
	"bytes"
	"encoding/json"
	"os"
)

/* WIP */
func SaveGame() {

	tempPath := "save.dat.tmp"
	finalPath := "save.dat"

	glob.SuperChunkListLock.RLock()
	var SuperChuckListTmp []*glob.MapSuperChunk
	copy(SuperChuckListTmp, glob.SuperChunkList)
	glob.SuperChunkListLock.RUnlock()

	tempList := []*glob.SaveMObj{}
	for _, sChunk := range SuperChuckListTmp {
		sChunk.Lock.RLock()
		for _, chunk := range sChunk.ChunkList {
			chunk.Lock.RLock()
			for _, mObj := range chunk.ObjList {
				tempList = append(tempList, &glob.SaveMObj{O: mObj})
			}
			chunk.Lock.RUnlock()
		}
		sChunk.Lock.RUnlock()
	}

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

/* WIP */
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
