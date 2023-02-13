package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
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
	P  world.XY             `json:"p,omitempty"`
	I  uint8                `json:"i,omitempty"`
	D  uint8                `json:"d,omitempty"`
	C  []*world.MatData     `json:"c,omitempty"`
	F  []*world.MatData     `json:"f,omitempty"`
	KF float32              `json:"kf,omitempty"`
	K  float32              `json:"k,omitempty"`
	PO []*world.ObjPortData `json:"po,omitempty"`
	T  uint8                `json:"t,omitempty"`
}

/* WIP */
func SaveGame() {

	if world.WASMMode {
		return
	}

	go func() {
		tempPath := "save.dat.tmp"
		finalPath := "save.dat"

		//start := time.Now()
		//("Save starting.")

		/* Pause the whole world ... */
		world.SuperChunkListLock.RLock()
		world.TickListLock.Lock()
		world.TockListLock.Lock()

		var seeds MapSeedsData
		for _, nl := range NoiseLayers {
			switch nl.TypeI {
			case gv.MAT_NONE:
				seeds.Grass = nl.Seed
			case gv.MAT_OIL:
				seeds.Oil = nl.Seed
			case gv.MAT_GAS:
				seeds.Gas = nl.Seed
			case gv.MAT_COAL:
				seeds.Coal = nl.Seed
			case gv.MAT_IRON_ORE:
				seeds.Iron = nl.Seed
			case gv.MAT_COPPER:
				seeds.Copper = nl.Seed
			case gv.MAT_STONE:
				seeds.Stone = nl.Seed
			}
		}

		tempList := gameSave{
			Version:  1,
			Date:     time.Now().Unix(),
			MapSeeds: seeds}
		for _, sChunk := range world.SuperChunkList {
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
						copy(tobj.C, []*world.MatData{cont})
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
						copy(tobj.PO, []*world.ObjPortData{po})
					}
					for p := range tobj.PO {
						if tobj.PO[p] == nil {
							continue
						}
						tobj.PO[p].Buf.TypeI = tobj.PO[p].Buf.TypeP.TypeI
						tobj.PO[p].Buf.TypeP = nil
						tobj.PO[p].Obj = nil
					}

					tempList.Objects = append(tempList.Objects, tobj)
				}
			}
		}
		//fmt.Println("WALK COMPLETE:", time.Since(start).String())

		b, err := json.Marshal(tempList)

		world.SuperChunkListLock.RUnlock()
		world.TickListLock.Unlock()
		world.TockListLock.Unlock()
		//fmt.Println("ENCODE DONE (WORLD UNLOCKED):", time.Since(start).String())

		if err != nil {
			//fmt.Printf("SaveGame: encode error: %v\n", err)
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

		//fmt.Println("COMPRESS & WRITE COMPLETE:", time.Since(start).String())
	}()
}

/* WIP */
func LoadGame() {

	if world.WASMMode {
		return
	}

	go func() {

		//start := time.Now()

		b, err := os.ReadFile("save.dat")
		if err != nil {
			fmt.Printf("LoadGame: file not found: %v\n", err)
			return
		}

		//fmt.Println("save read:", time.Since(start).String())
		data := util.UncompressZip(b)
		//fmt.Println("uncompressed:", time.Since(start).String())
		dbuf := bytes.NewBuffer(data)

		dec := json.NewDecoder(dbuf)

		/* Pause the whole world ... */
		world.SuperChunkListLock.RLock()
		world.TickListLock.Lock()
		world.TockListLock.Lock()
		tempList := gameSave{}
		err = dec.Decode(&tempList)
		if err != nil {
			fmt.Printf("LoadGame: JSON decode error: %v\n", err)
			return
		}
		//fmt.Println("json decoded:", time.Since(start).String())
		world.SuperChunkListLock.RUnlock()
		for n, nl := range NoiseLayers {
			switch nl.TypeI {
			case gv.MAT_NONE:
				NoiseLayers[n].Seed = tempList.MapSeeds.Grass
			case gv.MAT_OIL:
				NoiseLayers[n].Seed = tempList.MapSeeds.Oil
			case gv.MAT_GAS:
				NoiseLayers[n].Seed = tempList.MapSeeds.Gas
			case gv.MAT_COAL:
				NoiseLayers[n].Seed = tempList.MapSeeds.Coal
			case gv.MAT_IRON_ORE:
				NoiseLayers[n].Seed = tempList.MapSeeds.Iron
			case gv.MAT_COPPER:
				NoiseLayers[n].Seed = tempList.MapSeeds.Copper
			case gv.MAT_STONE:
				NoiseLayers[n].Seed = tempList.MapSeeds.Stone
			}
		}

		NoiseInit()

		/* Needs unsafeCreateObj that can accept a starting data set */
		count := 0
		for i := range tempList.Objects {

			obj := &world.ObjData{
				Pos:   util.UnCenterXY(tempList.Objects[i].P),
				TypeP: GameObjTypes[tempList.Objects[i].I],
				Dir:   tempList.Objects[i].D,
				//Contents:  item.C,
				KGFuel: tempList.Objects[i].KF,
				KGHeld: tempList.Objects[i].K,
				//Ports:     item.PO,
				TickCount: tempList.Objects[i].T,
			}
			copy(obj.Contents[:], tempList.Objects[i].C)
			copy(obj.Ports[:], tempList.Objects[i].PO)

			for c, cont := range obj.Contents {
				if cont == nil {
					continue
				}
				obj.Contents[c].TypeP = MatTypes[cont.TypeI]
			}

			for p, port := range obj.TypeP.Ports {
				if obj.Ports[p] == nil {
					obj.Ports[p] = &world.ObjPortData{}
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
			MakeChunk(util.UnCenterXY(tempList.Objects[i].P))
			chunk := util.GetChunk(util.UnCenterXY(tempList.Objects[i].P))
			obj.Parent = chunk

			obj.Parent.ObjMap[util.UnCenterXY(tempList.Objects[i].P)] = obj
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

			if obj.TypeP.InitObj != nil {
				obj.TypeP.InitObj(obj)
			}

			chunk.Parent.PixmapDirty = true

			count++
		}

		/* Refresh minerals */
		for _, sChunk := range world.SuperChunkList {
			drawMineral(sChunk)
		}
		world.VisDataDirty.Store(true)

		world.TickListLock.Unlock()
		world.TockListLock.Unlock()

		//fmt.Printf("%v objects created, Completed in %v\n", count, time.Since(start).String())
	}()
}
