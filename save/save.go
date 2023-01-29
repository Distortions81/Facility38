package save

import (
	"GameTest/glob"
	"GameTest/util"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func SaveGame() {

	tempPath := "save.dat.tmp"
	finalPath := "save.dat"

	tempList := []*glob.SaveMObj{}
	glob.SuperChunkMapLock.Lock()
	for _, sChunk := range glob.SuperChunkMap {
		for _, chunk := range sChunk.Chunks {
			for pos, mObj := range chunk.WObject {
				tempList = append(tempList, &glob.SaveMObj{O: mObj, P: pos})
			}
		}
	}
	glob.SuperChunkMapLock.Unlock()

	b, err := json.MarshalIndent(tempList, "", "\t")

	if err != nil {
		fmt.Println("SaveGame: enc.Encode failure")
		fmt.Println(err)
		return
	}

	_, err = os.Create(tempPath)

	if err != nil {
		fmt.Println("SaveGame: os.Create failure")
		return
	}

	zip := util.CompressZip(b)

	err = os.WriteFile(tempPath, zip, 0644)

	if err != nil {
		fmt.Println("SaveGame: WriteFile failure")
	}

	err = os.Rename(tempPath, finalPath)

	if err != nil {
		fmt.Println("Couldn't rename SaveGame file.")
		return
	}
}

func LoadGame() {
	b, _ := os.ReadFile("save.dat")
	data := util.UncompressZip(b)
	dbuf := bytes.NewBuffer(data)

	dec := json.NewDecoder(dbuf)
	err := dec.Decode(&glob.SuperChunkMap)
	if err != nil {
		//fmt.Println("LoadGame: dec.Decode failure")
		return
	}
}
