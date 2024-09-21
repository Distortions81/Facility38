package main

import (
	"fmt"
	"math"
	"time"
)

func init() {
	defer reportPanic("obj-init init")

	for i := range matTypes {
		matTypes[i].typeI = uint8(i)
	}
	for i := range worldOverlays {
		worldOverlays[i].typeI = uint8(i)
	}
	for i := range uiObjs {
		uiObjs[i].typeI = uint8(i)
	}

	/* Pre-calculate some object values */
	for i := range worldObjs {

		/* Convert mining amount to interval */
		if worldObjs[i].machineSettings.kgHourMine > 0 {
			worldObjs[i].machineSettings.kgPerCycle = ((worldObjs[i].machineSettings.kgHourMine / 60 / 60 / objectUPS) * float32(worldObjs[i].tockInterval)) * gameTimescale
		}
		/* Convert Horsepower to solid to KW and solid fuel per interval */
		if worldObjs[i].machineSettings.hp > 0 {
			kw := worldObjs[i].machineSettings.hp * HP_PER_KW
			coalKg := kw / COAL_KWH_PER_KG
			worldObjs[i].machineSettings.kgFuelPerCycle = ((coalKg / 60 / 60 / objectUPS) * float32(worldObjs[i].tockInterval)) * gameTimescale
			/* Convert KW to solid fuel per interval */
		} else if worldObjs[i].machineSettings.kw > 0 {
			coalKg := worldObjs[i].machineSettings.kw / COAL_KWH_PER_KG
			worldObjs[i].machineSettings.kgFuelPerCycle = ((coalKg / 60 / 60 / objectUPS) * float32(worldObjs[i].tockInterval)) * gameTimescale
		}

		/* Auto calculate max fuel from fuel used per interval */
		if worldObjs[i].machineSettings.kgFuelPerCycle > 0 {
			worldObjs[i].machineSettings.maxFuelKG = (worldObjs[i].machineSettings.kgFuelPerCycle * 10)
			if worldObjs[i].machineSettings.maxFuelKG < 50 {
				worldObjs[i].machineSettings.maxFuelKG = 50
			}
		}

		/* Auto calculate max contain for miners */
		if worldObjs[i].machineSettings.kgPerCycle > 0 {
			worldObjs[i].machineSettings.maxContainKG = (worldObjs[i].machineSettings.kgPerCycle * 10)
			if worldObjs[i].machineSettings.maxContainKG < 50 {
				worldObjs[i].machineSettings.maxContainKG = 50
			}
		}

		/* Flag item ports */
		for p := range worldObjs[i].ports {
			pt := worldObjs[i].ports[p].Type

			if pt == PORT_IN {
				worldObjs[i].hasInputs = true
			}
			if pt == PORT_OUT {
				worldObjs[i].hasOutputs = true
			}
			if pt == PORT_FOUT {
				worldObjs[i].hasFOut = true
			}
			if pt == PORT_FIN {
				worldObjs[i].hasFIn = true
			}

		}

		/* Flag non-square items */
		if worldObjs[i].size.X != worldObjs[i].size.Y {
			worldObjs[i].nonSquare = true
		}
		if worldObjs[i].size.X > 1 || worldObjs[i].size.Y > 1 {
			worldObjs[i].multiTile = true
		}
	}

	/* Add spaces to unit names */
	for _, mat := range matTypes {
		mat.unitName = " " + mat.unitName
	}

	//DumpItems()
}

func initSmelter(obj *ObjData) bool {
	defer reportPanic("initSmelter")
	if obj == nil {
		return false
	}

	obj.Unique.SingleContent = &MatData{}
	obj.Unique.SingleContent.typeP = matTypes[MAT_MIX_ORE]

	return true
}

func initMiner(obj *ObjData) bool {
	defer reportPanic("initMiner")
	if obj == nil {
		return false
	}

	obj.MinerData = &minerDataType{}

	foundRes := false
	/* Check for resources to mine */
	for p := 1; p < len(noiseLayers); p++ {
		var h float32 = float32(math.Abs(float64(noiseMap(float32(obj.Pos.X), float32(obj.Pos.Y), p))))

		/* We only mine solids */
		if !noiseLayers[p].typeP.isSolid {
			continue
		}
		if h > 0 {
			obj.MinerData.resources = append(obj.MinerData.resources, h)
			obj.MinerData.resourcesType = append(obj.MinerData.resourcesType, noiseLayers[p].typeI)
			obj.MinerData.resourceLayer = append(obj.MinerData.resourceLayer, uint8(p))
			obj.MinerData.resourcesCount++
			foundRes = true
		}
	}

	/* Nothing to mine here, kill all the events for this miner */
	if !foundRes {

		/* Let user know of this */
		if infoLine {
			chatDetailed(fmt.Sprintf("%v at %v: No solid resources to mine here!", obj.Unique.typeP.name, posToString(obj.Pos)),
				ColorRed, time.Minute)
		}

		obj.blocked = true
		obj.active = false
		return false /* Stop here */

	} else {

		tData := &tileData{minerData: &minerData{}}

		/* Init miner data */
		for _, sPos := range obj.Unique.typeP.subObjs {
			obj.chunk.tileMap[XY{X: uint16(sPos.X + int32(obj.Pos.X)), Y: uint16(sPos.Y + int32(obj.Pos.Y))}] = tData
		}
		obj.Tile = obj.chunk.tileMap[obj.Pos]
	}

	return true
}

func deInitMiner(obj *ObjData) {
	defer reportPanic("deInitMiner")
	/* Update resource map on remove */
	obj.chunk.parent.resourceDirty = true
}

func initBeltOver(obj *ObjData) bool {
	defer reportPanic("initBeltOver")
	obj.beltOver = &beltOverType{}
	obj.beltOver.middle = &MatData{}
	return true
}

func initSlipRoller(obj *ObjData) bool {
	defer reportPanic("initSlipRoller")
	if obj == nil {
		return false
	}

	obj.Unique.SingleContent = &MatData{}
	obj.Unique.SingleContent.typeP = matTypes[MAT_IRON_SHEET]

	return true
}
