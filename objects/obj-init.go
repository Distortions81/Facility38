package objects

import (
	"GameTest/util"
	"GameTest/world"
	"fmt"
	"math"
	"time"
)

func initSmelter(obj *world.ObjData) bool {
	if obj == nil {
		return false
	}

	obj.SingleContent = &world.MatData{}

	return true
}

func initMiner(obj *world.ObjData) bool {
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
		util.ChatDetailed(fmt.Sprintf("%v at %v: No resources to mine here!", obj.TypeP.Name, util.PosToString(obj.Pos)),
			world.ColorRed, time.Minute)

		obj.Blocked = true
		obj.Active = false
		return false /* Stop here */

	} else {

		/* Init miner data */
		obj.Parent.TileMap[obj.Pos] = &world.TileData{MinerData: &world.MinerData{}}
		obj.Tile = obj.Parent.TileMap[obj.Pos]
	}

	return true
}

func deinitMiner(obj *world.ObjData) {

	/* Update resource map on remove */
	obj.Parent.Parent.ResourceDirty = true
}
