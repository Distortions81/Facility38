package main

import (
	"Facility38/def"
	"Facility38/util"
	"Facility38/world"
	"time"
)

/* Make a superchunk */
func makeSuperChunk(pos world.XY) {
	defer util.ReportPanic("makeSuperChunk")
	//Make super chunk if needed

	newPos := pos
	scpos := util.PosToSuperChunkPos(newPos)

	world.SuperChunkMapLock.Lock()  //Lock Superclunk map
	world.SuperChunkListLock.Lock() //Lock Superchunk list

	if world.SuperChunkMap[scpos] == nil {

		/* Make new superchunk in map at pos */
		newSuperChunk := &world.MapSuperChunk{}

		maxSize := def.SuperChunkTotal * def.SuperChunkTotal * 4
		newSuperChunk.ItemMap = make([]byte, maxSize)
		newSuperChunk.ResourceMap = make([]byte, maxSize)
		newSuperChunk.ResourceDirty = true

		world.SuperChunkMap[scpos] = newSuperChunk
		world.SuperChunkMap[scpos].Lock.Lock() //Lock chunk

		world.SuperChunkList =
			append(world.SuperChunkList, world.SuperChunkMap[scpos])
		world.SuperChunkMap[scpos].ChunkMap = make(map[world.XY]*world.MapChunk)

		/* Save position */
		world.SuperChunkMap[scpos].Pos = scpos

		world.SuperChunkMap[scpos].Lock.Unlock()
	}
	world.SuperChunkListLock.Unlock()
	world.SuperChunkMapLock.Unlock()
}

/* Make a chunk, insert into superchunk */
func MakeChunk(pos world.XY) bool {
	defer util.ReportPanic("MakeChunk")
	//Make chunk if needed

	newPos := pos

	makeSuperChunk(pos)

	cpos := util.PosToChunkPos(newPos)
	scpos := util.PosToSuperChunkPos(newPos)

	world.SuperChunkMapLock.Lock()  //Lock Superclunk map
	world.SuperChunkListLock.Lock() //Lock Superchunk list

	if world.SuperChunkMap[scpos].ChunkMap[cpos] == nil {
		/* Increase chunk count */
		world.SuperChunkMap[scpos].NumChunks++

		/* Make a new empty chunk in the map at pos */
		world.SuperChunkMap[scpos].ChunkMap[cpos] = &world.MapChunk{}
		world.SuperChunkMap[scpos].Lock.Lock() //Lock chunk

		/* Append to chunk list */
		world.SuperChunkMap[scpos].ChunkList =
			append(world.SuperChunkMap[scpos].ChunkList, world.SuperChunkMap[scpos].ChunkMap[cpos])

		world.SuperChunkMap[scpos].ChunkMap[cpos].BuildingMap = make(map[world.XY]*world.BuildingData)
		world.SuperChunkMap[scpos].ChunkMap[cpos].TileMap = make(map[world.XY]*world.TileData)

		/* Terrain img */
		world.SuperChunkMap[scpos].ChunkMap[cpos].TerrainLock.Lock()
		world.SuperChunkMap[scpos].ChunkMap[cpos].TerrainImage = world.TempChunkImage
		world.SuperChunkMap[scpos].ChunkMap[cpos].UsingTemporary = true
		world.SuperChunkMap[scpos].ChunkMap[cpos].TerrainLock.Unlock()

		/* Save position */
		world.SuperChunkMap[scpos].ChunkMap[cpos].Pos = cpos

		/* Save parent */
		world.SuperChunkMap[scpos].ChunkMap[cpos].Parent = world.SuperChunkMap[scpos]

		world.SuperChunkMap[scpos].Lock.Unlock()

		world.SuperChunkListLock.Unlock()
		world.SuperChunkMapLock.Unlock()
		return true
	}

	world.SuperChunkListLock.Unlock()
	world.SuperChunkMapLock.Unlock()
	return false
}

/* Explore (input) chunks + and - */
func ExploreMap(pos world.XY, input int, slow bool) {
	defer util.ReportPanic("ExploreMap")
	/* Explore some map */

	ChunksMade := 0
	area := input * def.ChunkSize
	offx := int(pos.X) - (area / 2)
	offy := int(pos.Y) - (area / 2)
	for x := -area; x < area; x += def.ChunkSize {
		for y := -area; y < area; y += def.ChunkSize {
			pos := world.XY{X: uint16(offx - x), Y: uint16(offy - y)}
			MakeChunk(pos)
			ChunksMade++

			if slow && ChunksMade%10 == 0 {
				world.MapLoadPercent = float32(ChunksMade) / float32((input*2)*(input*2)) * 100.0
				if world.WASMMode {
					time.Sleep(time.Nanosecond)
				} else {
					time.Sleep(time.Millisecond * 5)
				}
			}
		}
	}
}
