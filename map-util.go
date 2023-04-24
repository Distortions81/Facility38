package main

import (
	"Facility38/gv"
	"Facility38/util"
	"Facility38/world"
	"time"
)

/* Make a superchunk */
func makeSuperChunk(pos world.XY) {
	//Make super chunk if needed

	newPos := pos
	scpos := util.PosToSuperChunkPos(newPos)

	world.SuperChunkMapLock.Lock()  //Lock Superclunk map
	world.SuperChunkListLock.Lock() //Lock Superchunk list

	if world.SuperChunkMap[scpos] == nil {

		/* Make new superchunk in map at pos */
		newSuperChunk := &world.MapSuperChunk{}

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
		world.SuperChunkMap[scpos].ChunkMap[cpos].TerrainImage = world.TempChunkImage
		world.SuperChunkMap[scpos].ChunkMap[cpos].UsingTemporary = true

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
	/* Explore some map */

	ChunksMade := 0
	area := input * gv.ChunkSize
	offx := int(pos.X) - (area / 2)
	offy := int(pos.Y) - (area / 2)
	for x := -area; x < area; x += gv.ChunkSize {
		for y := -area; y < area; y += gv.ChunkSize {
			pos := world.XY{X: uint16(offx - x), Y: uint16(offy - y)}
			MakeChunk(pos)
			ChunksMade++
			if !gv.LoadTest {
				world.MapLoadPercent = float32(ChunksMade) / float32((input*2)*(input*2)) * 100.0
			}
			if slow {
				if gv.WASMMode {
					time.Sleep(time.Nanosecond)
				} else {
					time.Sleep(time.Microsecond * 100)
				}
			}
		}
	}
	if !gv.LoadTest {
		world.MapLoadPercent = 100
	}
}
