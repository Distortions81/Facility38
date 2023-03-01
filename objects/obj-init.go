package objects

import (
	"GameTest/gv"
	"GameTest/util"
	"GameTest/world"
	"fmt"
	"math"
	"time"
)

func initMiner(obj *world.ObjData) {
	if obj == nil {
		return
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
		if h > 0.01 {
			obj.MinerData.Resources = append(obj.MinerData.Resources, h)
			obj.MinerData.ResourcesType = append(obj.MinerData.ResourcesType, NoiseLayers[p].TypeI)
			obj.MinerData.ResourcesCount++
			foundRes = true
		}
	}

	/* Nothing to mine here, kill all the events for this miner */
	if !foundRes {

		/* Let user know of this */
		oPos := util.CenterXY(obj.Pos)
		util.ChatDetailed(fmt.Sprintf("%v at (%v,%v): No resources to mine here!", obj.TypeP.Name, oPos.X, oPos.Y),
			world.ColorRed, time.Minute)

		obj.Blocked = true
		obj.Active = false
		EventQueueAdd(obj, gv.QUEUE_TYPE_TICK, true)
		EventQueueAdd(obj, gv.QUEUE_TYPE_TOCK, true)

		return /* Stop here */
	}

	/* Init miner data */
	obj.Parent.TileMap[obj.Pos] = &world.TileData{MinerData: &world.MinerData{}}
	obj.Tile = obj.Parent.TileMap[obj.Pos]
}

func deinitMiner(obj *world.ObjData) {

	/* Update resource map on remove */
	obj.Parent.Parent.ResouceDirty = true
}
