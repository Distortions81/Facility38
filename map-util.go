package main

import (
	"time"
)

/* Make a superchunk */
func makeSuperChunk(pos XY) {
	defer reportPanic("makeSuperChunk")
	//Make super chunk if needed

	newPos := pos
	scpos := PosToSuperChunkPos(newPos)

	SuperChunkMapLock.Lock()  //Lock Superclunk map
	SuperChunkListLock.Lock() //Lock Superchunk list

	if SuperChunkMap[scpos] == nil {

		/* Make new superchunk in map at pos */
		newSuperChunk := &mapSuperChunkData{}

		maxSize := superChunkTotal * superChunkTotal * 4
		newSuperChunk.itemMap = make([]byte, maxSize)
		newSuperChunk.resourceMap = make([]byte, maxSize)
		newSuperChunk.resourceDirty = true

		SuperChunkMap[scpos] = newSuperChunk
		SuperChunkMap[scpos].lock.Lock() //Lock chunk

		SuperChunkList =
			append(SuperChunkList, SuperChunkMap[scpos])
		SuperChunkMap[scpos].chunkMap = make(map[XY]*mapChunk)

		/* Save position */
		SuperChunkMap[scpos].pos = scpos

		SuperChunkMap[scpos].lock.Unlock()
	}
	SuperChunkListLock.Unlock()
	SuperChunkMapLock.Unlock()
}

/* Make a chunk, insert into superchunk */
func makeChunk(pos XY) bool {
	defer reportPanic("makeChunk")
	//Make chunk if needed

	newPos := pos

	makeSuperChunk(pos)

	cpos := PosToChunkPos(newPos)
	scpos := PosToSuperChunkPos(newPos)

	SuperChunkMapLock.Lock()  //Lock Superclunk map
	SuperChunkListLock.Lock() //Lock Superchunk list

	if SuperChunkMap[scpos].chunkMap[cpos] == nil {
		/* Increase chunk count */
		SuperChunkMap[scpos].numChunks++

		/* Make a new empty chunk in the map at pos */
		SuperChunkMap[scpos].chunkMap[cpos] = &mapChunk{}
		SuperChunkMap[scpos].lock.Lock() //Lock chunk

		/* Append to chunk list */
		SuperChunkMap[scpos].chunkList =
			append(SuperChunkMap[scpos].chunkList, SuperChunkMap[scpos].chunkMap[cpos])

		SuperChunkMap[scpos].chunkMap[cpos].buildingMap = make(map[XY]*buildingData)
		SuperChunkMap[scpos].chunkMap[cpos].tileMap = make(map[XY]*tileData)

		/* Terrain img */
		SuperChunkMap[scpos].chunkMap[cpos].terrainLock.Lock()
		SuperChunkMap[scpos].chunkMap[cpos].terrainImage = TempChunkImage
		SuperChunkMap[scpos].chunkMap[cpos].usingTemporary = true
		SuperChunkMap[scpos].chunkMap[cpos].terrainLock.Unlock()

		/* Save position */
		SuperChunkMap[scpos].chunkMap[cpos].pos = cpos

		/* Save parent */
		SuperChunkMap[scpos].chunkMap[cpos].parent = SuperChunkMap[scpos]

		SuperChunkMap[scpos].lock.Unlock()

		SuperChunkListLock.Unlock()
		SuperChunkMapLock.Unlock()
		return true
	}

	SuperChunkListLock.Unlock()
	SuperChunkMapLock.Unlock()
	return false
}

/* Explore (input) chunks + and - */
func exploreMap(pos XY, input int, slow bool) {
	defer reportPanic("exploreMap")
	/* Explore some map */

	ChunksMade := 0
	area := input * chunkSize
	offx := int(pos.X) - (area / 2)
	offy := int(pos.Y) - (area / 2)
	for x := -area; x < area; x += chunkSize {
		for y := -area; y < area; y += chunkSize {
			pos := XY{X: uint16(offx - x), Y: uint16(offy - y)}
			makeChunk(pos)
			ChunksMade++

			if slow && ChunksMade%10 == 0 {
				MapLoadPercent = float32(ChunksMade) / float32((input*2)*(input*2)) * 100.0
				if WASMMode {
					time.Sleep(time.Nanosecond)
				} else {
					time.Sleep(time.Millisecond * 5)
				}
			}
		}
	}
}
