package objects

import (
	"GameTest/glob"
	"GameTest/gv"
	"GameTest/util"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

/* Used to munge data into a test save file */
/* TODO: SAVE VERSION AND MAP SEED INTO FILE */
type saveMObj struct {
	P  glob.XY             `json:"p,omitempty"`
	I  uint8               `json:"i,omitempty"`
	D  uint8               `json:"d,omitempty"`
	C  []*glob.MatData     `json:"c,omitempty"`
	F  []*glob.MatData     `json:"f,omitempty"`
	KF float64             `json:"kf,omitempty"`
	K  float64             `json:"k,omitempty"`
	PO []*glob.ObjPortData `json:"po,omitempty"`
	T  uint8               `json:"t,omitempty"`
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
			for _, chunk := range sChunk.ChunkList {
				for _, mObj := range chunk.ObjList {
					tobj := &saveMObj{
						P: util.CenterXY(mObj.Pos),
						I: mObj.TypeP.TypeI,
						D: mObj.Dir,
						//C:  mObj.Contents,
						//F:  mObj.Fuel,
						KF: mObj.KGFuel,
						K:  mObj.KGHeld,
						//PO: mObj.Ports,
						T: mObj.TickCount,
					}

					for _, cont := range mObj.Contents {
						if cont == nil {
							continue
						}
						copy(tobj.C, []*glob.MatData{cont})
					}
					for c := range tobj.C {
						if tobj.C[c] == nil {
							continue
						}
						tobj.C[c].TypeI = tobj.C[c].TypeP.TypeI
						tobj.C[c].TypeP = nil
					}

					for _, po := range mObj.Ports {
						if po == nil || po.Buf.TypeP == nil {
							continue
						}
						copy(tobj.PO, []*glob.ObjPortData{po})
					}
					for p := range tobj.PO {
						if tobj.PO[p] == nil {
							continue
						}
						tobj.PO[p].Buf.TypeI = tobj.PO[p].Buf.TypeP.TypeI
						tobj.PO[p].Buf.TypeP = nil
						tobj.PO[p].Obj = nil
					}

					tempList = append(tempList, tobj)
				}
			}
		}
		fmt.Println("WALK COMPLETE:", time.Since(start).String())

		//b, err := json.Marshal(tempList)

		var b []byte
		buf := bytes.NewBuffer(b)
		enc := gob.NewEncoder(buf)
		err := enc.Encode(tempList)

		glob.SuperChunkListLock.RUnlock()
		TickListLock.Unlock()
		TockListLock.Unlock()
		fmt.Println("ENCODE DONE (WORLD UNLOCKED):", time.Since(start).String())

		if err != nil {
			fmt.Printf("SaveGame: encode error: %v\n", err)
			//return
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
	go func() {

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
		for i := range tempList {

			obj := &glob.ObjData{
				Pos:   util.UnCenterXY(tempList[i].P),
				TypeP: GameObjTypes[tempList[i].I],
				Dir:   tempList[i].D,
				//Contents:  item.C,
				KGFuel: tempList[i].KF,
				KGHeld: tempList[i].K,
				//Ports:     item.PO,
				TickCount: tempList[i].T,
			}
			copy(obj.Contents[:], tempList[i].C)
			copy(obj.Ports[:], tempList[i].PO)

			for c, cont := range obj.Contents {
				if cont == nil {
					continue
				}
				obj.Contents[c].TypeP = MatTypes[cont.TypeI]
			}

			for p, port := range obj.TypeP.Ports {
				if obj.Ports[p] == nil {
					obj.Ports[p] = &glob.ObjPortData{}
				}
				obj.Ports[p].PortDir = port
			}
			for x := 0; x < int(obj.Dir); x++ {
				util.RotatePortsCW(obj)
			}

			for p, port := range obj.Ports {
				if port == nil {
					continue
				}
				obj.Ports[p].Buf.TypeP = MatTypes[port.Buf.TypeI]
			}

			/* Relink */
			MakeChunk(util.UnCenterXY(tempList[i].P))
			chunk := util.GetChunk(util.UnCenterXY(tempList[i].P))
			obj.Parent = chunk

			obj.Parent.ObjMap[util.UnCenterXY(tempList[i].P)] = obj
			obj.Parent.ObjList = append(obj.Parent.ObjList, obj)
			chunk.Parent.PixmapDirty = true
			chunk.NumObjects++

			LinkObj(obj)

			/* Only add to list if the object calls an update function */
			if obj.TypeP.UpdateObj != nil {
				tockListAdd(obj)
			}

			if util.ObjHasPort(obj, gv.PORT_OUTPUT) {
				ticklistAdd(obj)
			}

			chunk.Parent.PixmapDirty = true

			count++
		}
		glob.VisDataDirty.Store(true)

		TickListLock.Unlock()
		TockListLock.Unlock()
		fmt.Printf("%v objects created, Completed in %v\n", count, time.Since(start).String())
	}()
}
