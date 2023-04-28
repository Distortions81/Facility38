package main

import (
	"Facility38/def"
	"Facility38/util"
	"Facility38/world"
	"fmt"
	"math"
	"time"
)

func init() {
	defer util.ReportPanic("obj-init init")

	for i := range MatTypes {
		MatTypes[i].TypeI = uint8(i)
	}
	for i := range WorldOverlays {
		WorldOverlays[i].TypeI = uint8(i)
	}
	for i := range UIObjs {
		UIObjs[i].TypeI = uint8(i)
	}

	/* Pre-calculate some object values */
	for i := range WorldObjs {

		/* Convert mining amount to interval */
		if WorldObjs[i].MachineSettings.KgHourMine > 0 {
			WorldObjs[i].MachineSettings.KgPerCycle = ((WorldObjs[i].MachineSettings.KgHourMine / 60 / 60 / world.ObjectUPS) * float32(WorldObjs[i].TockInterval)) * def.TIMESCALE_MULTI
		}
		/* Convert Horsepower to solid to KW and solid fuel per interval */
		if WorldObjs[i].MachineSettings.HP > 0 {
			KW := WorldObjs[i].MachineSettings.HP * def.HP_PER_KW
			COALKG := KW / def.COAL_KWH_PER_KG
			WorldObjs[i].MachineSettings.KgFuelPerCycle = ((COALKG / 60 / 60 / world.ObjectUPS) * float32(WorldObjs[i].TockInterval)) * def.TIMESCALE_MULTI
			/* Convert KW to solid fuel per interval */
		} else if WorldObjs[i].MachineSettings.KW > 0 {
			COALKG := WorldObjs[i].MachineSettings.KW / def.COAL_KWH_PER_KG
			WorldObjs[i].MachineSettings.KgFuelPerCycle = ((COALKG / 60 / 60 / world.ObjectUPS) * float32(WorldObjs[i].TockInterval)) * def.TIMESCALE_MULTI
		}

		/* Auto calculate max fuel from fuel used per interval */
		if WorldObjs[i].MachineSettings.KgFuelPerCycle > 0 {
			WorldObjs[i].MachineSettings.MaxFuelKG = (WorldObjs[i].MachineSettings.KgFuelPerCycle * 10)
			if WorldObjs[i].MachineSettings.MaxFuelKG < 50 {
				WorldObjs[i].MachineSettings.MaxFuelKG = 50
			}
		}

		/* Auto calculate max contain for miners */
		if WorldObjs[i].MachineSettings.KgPerCycle > 0 {
			WorldObjs[i].MachineSettings.MaxContainKG = (WorldObjs[i].MachineSettings.KgPerCycle * 10)
			if WorldObjs[i].MachineSettings.MaxContainKG < 50 {
				WorldObjs[i].MachineSettings.MaxContainKG = 50
			}
		}

		/* Flag item ports */
		for p := range WorldObjs[i].Ports {
			pt := WorldObjs[i].Ports[p].Type

			if pt == def.PORT_IN {
				WorldObjs[i].HasInputs = true
			}
			if pt == def.PORT_OUT {
				WorldObjs[i].HasOutputs = true
			}
			if pt == def.PORT_FOUT {
				WorldObjs[i].HasFOut = true
			}
			if pt == def.PORT_FIN {
				WorldObjs[i].HasFIn = true
			}

		}

		/* Flag non-square items */
		if WorldObjs[i].Size.X != WorldObjs[i].Size.Y {
			WorldObjs[i].NonSquare = true
		}
		if WorldObjs[i].Size.X > 1 || WorldObjs[i].Size.Y > 1 {
			WorldObjs[i].MultiTile = true
		}
	}

	/* Add spaces to unit names */
	for _, mat := range MatTypes {
		mat.UnitName = " " + mat.UnitName
	}

	//DumpItems()
}

func initSmelter(obj *world.ObjData) bool {
	defer util.ReportPanic("initSmelter")
	if obj == nil {
		return false
	}

	obj.Unique.SingleContent = &world.MatData{}
	obj.Unique.SingleContent.TypeP = MatTypes[def.MAT_MIX_ORE]

	return true
}

func initMiner(obj *world.ObjData) bool {
	defer util.ReportPanic("initMiner")
	if obj == nil {
		return false
	}

	obj.MinerData = &world.MinerDataType{}

	foundRes := false
	/* Check for resources to mine */
	for p := 1; p < len(NoiseLayers); p++ {
		var h float32 = float32(math.Abs(float64(NoiseMap(float32(obj.Pos.X), float32(obj.Pos.Y), p))))

		/* We only mine solids */
		if !NoiseLayers[p].TypeP.IsSolid {
			continue
		}
		if h > 0 {
			obj.MinerData.Resources = append(obj.MinerData.Resources, h)
			obj.MinerData.ResourcesType = append(obj.MinerData.ResourcesType, NoiseLayers[p].TypeI)
			obj.MinerData.ResourcesCount++
			foundRes = true
		}
	}

	/* Nothing to mine here, kill all the events for this miner */
	if !foundRes {

		/* Let user know of this */
		util.ChatDetailed(fmt.Sprintf("%v at %v: No resources to mine here!", obj.Unique.TypeP.Name, util.PosToString(obj.Pos)),
			world.ColorRed, time.Minute)

		obj.Blocked = true
		obj.Active = false
		return false /* Stop here */

	} else {

		/* Init miner data */
		obj.Chunk.TileMap[obj.Pos] = &world.TileData{MinerData: &world.MinerData{}}
		obj.Tile = obj.Chunk.TileMap[obj.Pos]
	}

	return true
}

func deinitMiner(obj *world.ObjData) {
	defer util.ReportPanic("deinitMiner")
	/* Update resource map on remove */
	obj.Chunk.Parent.ResourceDirty = true
}

func initBeltOver(obj *world.ObjData) bool {
	defer util.ReportPanic("initBeltOver")
	obj.BeltOver = &world.BeltOverType{}
	obj.BeltOver.Middle = &world.MatData{}
	return true
}

func initSlipRoller(obj *world.ObjData) bool {
	defer util.ReportPanic("initSlipRoller")
	if obj == nil {
		return false
	}

	obj.Unique.SingleContent = &world.MatData{}
	obj.Unique.SingleContent.TypeP = MatTypes[def.MAT_IRON_SHEET]

	return true
}
