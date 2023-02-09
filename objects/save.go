package objects

import (
	"GameTest/glob"
	"GameTest/util"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

/* WIP */
func SaveGame() {

	go func() {
		tempPath := "save.dat.tmp"
		finalPath := "save.dat"

		start := time.Now()
		fmt.Println("Save starting.")

		/* Pause the whole world ... */
		glob.SuperChunkListLock.RLock()
		TickListLock.Lock()
		TockListLock.Lock()

		tempList := []*glob.SaveMObj{}
		for _, sChunk := range glob.SuperChunkList {
			fmt.Println("sc:", sChunk.Pos)
			for _, chunk := range sChunk.ChunkList {
				for _, mObj := range chunk.ObjList {
					tempList = append(tempList, &glob.SaveMObj{O: mObj})
				}
			}
		}
		fmt.Println("WALK COMPLETE:", time.Since(start).String())

		b, err := json.MarshalIndent(tempList, "", "")

		glob.SuperChunkListLock.RUnlock()
		TickListLock.Unlock()
		TockListLock.Unlock()
		fmt.Println("ENCODE DONE (WORLD UNLOCKED):", time.Since(start).String())

		if err != nil {
			fmt.Println("SaveGame: encode error: %v", err)
			return
		}

		_, err = os.Create(tempPath)

		if err != nil {
			fmt.Println("SaveGame: os.Create error: %v", err)
			return
		}

		zip := util.CompressZip(b)

		err = os.WriteFile(tempPath, zip, 0644)

		if err != nil {
			fmt.Println("SaveGame: os.WriteFile error: %v", err)
		}

		err = os.Rename(tempPath, finalPath)

		if err != nil {
			fmt.Println("SaveGame: couldn't rename save file: %v", err)
			return
		}

		fmt.Println("COMPRESS COMPLETE:", time.Since(start).String())
	}()
}

/* WIP */
func LoadGame() {
	b, err := os.ReadFile("save.dat")
	if err != nil {
		fmt.Println("LoadGame: file not found: %v", err)
		return
	}
	data := util.UncompressZip(b)
	dbuf := bytes.NewBuffer(data)

	dec := json.NewDecoder(dbuf)
	err = dec.Decode(&glob.SuperChunkMap)
	if err != nil {
		fmt.Println("LoadGame: JSON decode error: %v", err)
		return
	}

}
