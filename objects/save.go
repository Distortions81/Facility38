package objects

import (
	"GameTest/cwlog"
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
	Pos      world.XY                       `json:"p,omitempty"`
	TypeI    uint8                          `json:"i,omitempty"`
	Dir      uint8                          `json:"d,omitempty"`
	Contents [gv.MAT_MAX]*world.MatData     `json:"c,omitempty"`
	KGFuel   float32                        `json:"kf,omitempty"`
	KGHeld   float32                        `json:"k,omitempty"`
	Ports    [gv.DIR_MAX]*world.ObjPortData `json:"po,omitempty"`
	Ticks    uint8                          `json:"t,omitempty"`
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
			Version:  2,
			Date:     time.Now().Unix(),
			MapSeeds: seeds}
		for _, sChunk := range world.SuperChunkList {
			for _, chunk := range sChunk.ChunkList {
				for _, mObj := range chunk.ObjList {
					tobj := &saveMObj{
						Pos:   util.CenterXY(mObj.Pos),
						TypeI: mObj.TypeP.TypeI,
						Dir:   mObj.Dir,
						//C:  mObj.Contents,
						KGFuel: mObj.KGFuel,
						KGHeld: mObj.KGHeld,
						//PO: mObj.Ports,
						Ticks: mObj.TickCount,
					}

					for c := range mObj.Contents {
						if mObj.Contents[c] == nil {
							continue
						}
						tobj.Contents[c] = mObj.Contents[c]
					}
					for c := range tobj.Contents {
						if tobj.Contents[c] == nil {
							continue
						}
						tobj.Contents[c].TypeI = tobj.Contents[c].TypeP.TypeI
						tobj.Contents[c].TypeP = nil
					}

					for p := range mObj.Ports {
						if mObj.Ports[p] == nil || mObj.Ports[p].Buf.Amount == 0 {
							continue
						}
						tobj.Ports[p] = mObj.Ports[p]
					}
					for p := range tobj.Ports {
						if tobj.Ports[p] == nil {
							continue
						}
						tobj.Ports[p].Buf.TypeI = tobj.Ports[p].Buf.TypeP.TypeI
						tobj.Ports[p].Buf.TypeP = nil
						tobj.Ports[p].Obj = nil
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

		//zip := util.CompressZip(b)

		err = os.WriteFile(tempPath, b, 0644)

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
		//data := util.UncompressZip(b)
		//fmt.Println("uncompressed:", time.Since(start).String())
		dbuf := bytes.NewBuffer(b)

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

		if tempList.Version != 2 {
			cwlog.DoLog("LoadGame: Invalid save version.")
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

		PerlinNoiseInit()

		/* Needs unsafeCreateObj that can accept a starting data set */
		count := 0
		for i := range tempList.Objects {

			obj := &world.ObjData{
				Pos:       util.UnCenterXY(tempList.Objects[i].Pos),
				TypeP:     GameObjTypes[tempList.Objects[i].TypeI],
				Dir:       tempList.Objects[i].Dir,
				Contents:  tempList.Objects[i].Contents,
				KGFuel:    tempList.Objects[i].KGFuel,
				KGHeld:    tempList.Objects[i].KGHeld,
				Ports:     tempList.Objects[i].Ports,
				TickCount: tempList.Objects[i].Ticks,
			}

			for c := range obj.Contents {
				if obj.Contents[c] == nil {
					continue
				}
				obj.Contents[c].TypeP = MatTypes[obj.Contents[c].TypeI]
			}

			for p := range obj.Ports {
				if obj.Ports[p] == nil {
					continue
				}
				obj.Ports[p].Buf.TypeP = MatTypes[obj.Ports[p].Buf.TypeI]
			}

			/* Relink */
			MakeChunk(util.UnCenterXY(tempList.Objects[i].Pos))
			chunk := util.GetChunk(util.UnCenterXY(tempList.Objects[i].Pos))
			obj.Parent = chunk

			obj.Parent.ObjMap[util.UnCenterXY(tempList.Objects[i].Pos)] = obj
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
