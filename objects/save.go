package objects

import (
	"GameTest/glob"
	"GameTest/gv"
	"GameTest/util"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

/* Used to munge data into a test save file */
type saveMObj struct {
	P  glob.XY
	I  uint8
	D  uint8
	C  [gv.MAT_MAX]*glob.MatData
	F  [gv.MAT_MAX]*glob.MatData
	KF float64
	K  float64

	PO [gv.DIR_MAX]*glob.ObjPortData
	T  uint8
}

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

		tempList := []*saveMObj{}
		for _, sChunk := range glob.SuperChunkList {
			fmt.Println("sc:", sChunk.Pos)
			for _, chunk := range sChunk.ChunkList {
				for _, mObj := range chunk.ObjList {
					tempList = append(tempList, &saveMObj{
						P:  mObj.Pos,
						I:  mObj.TypeP.TypeI,
						D:  mObj.Dir,
						C:  mObj.Contents,
						F:  mObj.Fuel,
						KF: mObj.KGFuel,
						K:  mObj.KGHeld,
						PO: mObj.Ports,
						T:  mObj.TickCount,
					})
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
			fmt.Printf("SaveGame: encode error: %v\n", err)
			return
		}

		_, err = os.Create(tempPath)

		if err != nil {
			fmt.Printf("SaveGame: os.Create error: %v\n", err)
			return
		}

		zip := util.CompressZip(b)

		err = os.WriteFile(tempPath, zip, 0644)

		if err != nil {
			fmt.Printf("SaveGame: os.WriteFile error: %v\n", err)
		}

		err = os.Rename(tempPath, finalPath)

		if err != nil {
			fmt.Printf("SaveGame: couldn't rename save file: %v\n", err)
			return
		}

		fmt.Println("COMPRESS & WRITE COMPLETE:", time.Since(start).String())
	}()
}

/* WIP */
func LoadGame() {

	start := time.Now()

	b, err := os.ReadFile("save.dat")
	if err != nil {
		fmt.Printf("LoadGame: file not found: %v\n", err)
		return
	}

	fmt.Println("save read:", time.Since(start).String())
	data := util.UncompressZip(b)
	fmt.Println("uncompressed:", time.Since(start).String())
	dbuf := bytes.NewBuffer(data)

	dec := json.NewDecoder(dbuf)

	/* Pause the whole world ... */
	glob.SuperChunkListLock.RLock()
	TickListLock.Lock()
	TockListLock.Lock()
	tempList := []*saveMObj{}
	err = dec.Decode(&tempList)
	if err != nil {
		fmt.Printf("LoadGame: JSON decode error: %v\n", err)
		return
	}
	fmt.Println("json decoded:", time.Since(start).String())
	glob.SuperChunkListLock.RUnlock()

	/* Needs unsafeCreateObj that can accept a starting data set */
	count := 0
	for _, item := range tempList {

		obj := glob.ObjData{
			Pos:       item.P,
			TypeP:     GameObjTypes[item.I],
			Dir:       item.D,
			Contents:  item.C,
			Fuel:      item.F,
			KGFuel:    item.KF,
			KGHeld:    item.K,
			Ports:     item.PO,
			TickCount: item.T,
		}
		/* Relink */
		MakeChunk(item.P)
		chunk := util.GetChunk(item.P)
		obj.Parent = chunk

		chunk.ObjList = append(chunk.ObjList, &obj)
		chunk.ObjMap[item.P] = &obj

		LinkObj(&obj)

		/* Only add to list if the object calls an update function */
		if obj.TypeP.UpdateObj != nil {
			tockListAdd(&obj)
		}

		if util.ObjHasPort(&obj, gv.PORT_OUTPUT) {
			ticklistAdd(&obj)
		}

		chunk.Parent.PixmapDirty = true
		chunk.NumObjects++
		count++
	}
	glob.VisDataDirty.Store(true)

	TickListLock.Unlock()
	TockListLock.Unlock()
	fmt.Printf("%v objects created, Completed in %v\n", count, time.Since(start).String())

}
