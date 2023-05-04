package main

import (
	"time"
)

/* Make a superchunk */
func makeSuperChunk(pos XY) {
	defer reportPanic("makeSuperChunk")
	//Make super chunk if needed

	newPos := pos
	scpos := posToSuperChunkPos(newPos)

	superChunkMapLock.Lock()  //Lock Superclunk map
	superChunkListLock.Lock() //Lock Superchunk list

	if superChunkMap[scpos] == nil {

		/* Make new superchunk in map at pos */
		newSuperChunk := &mapSuperChunkData{}

		maxSize := superChunkTotal * superChunkTotal * 4
		newSuperChunk.itemMap = make([]byte, maxSize)
		newSuperChunk.resourceMap = make([]byte, maxSize)
		newSuperChunk.resourceDirty = true

		superChunkMap[scpos] = newSuperChunk
		superChunkMap[scpos].lock.Lock() //Lock chunk

		superChunkList =
			append(superChunkList, superChunkMap[scpos])
		superChunkMap[scpos].chunkMap = make(map[XY]*mapChunk)

		/* Save position */
		superChunkMap[scpos].pos = scpos

		superChunkMap[scpos].lock.Unlock()
	}
	superChunkListLock.Unlock()
	superChunkMapLock.Unlock()
}

/* Make a chunk, insert into superchunk */
func makeChunk(pos XY) bool {
	defer reportPanic("makeChunk")
	//Make chunk if needed

	newPos := pos

	makeSuperChunk(pos)

	cpos := PosToChunkPos(newPos)
	scpos := posToSuperChunkPos(newPos)

	superChunkMapLock.Lock()  //Lock Superclunk map
	superChunkListLock.Lock() //Lock Superchunk list

	if superChunkMap[scpos].chunkMap[cpos] == nil {
		/* Increase chunk count */
		superChunkMap[scpos].numChunks++

		/* Make a new empty chunk in the map at pos */
		superChunkMap[scpos].chunkMap[cpos] = &mapChunk{}
		superChunkMap[scpos].lock.Lock() //Lock chunk

		/* Append to chunk list */
		superChunkMap[scpos].chunkList =
			append(superChunkMap[scpos].chunkList, superChunkMap[scpos].chunkMap[cpos])

		superChunkMap[scpos].chunkMap[cpos].buildingMap = make(map[XY]*buildingData)
		superChunkMap[scpos].chunkMap[cpos].tileMap = make(map[XY]*tileData)

		/* Terrain img */
		superChunkMap[scpos].chunkMap[cpos].terrainLock.Lock()
		superChunkMap[scpos].chunkMap[cpos].terrainImage = TempChunkImage
		superChunkMap[scpos].chunkMap[cpos].usingTemporary = true
		superChunkMap[scpos].chunkMap[cpos].terrainLock.Unlock()

		/* Save position */
		superChunkMap[scpos].chunkMap[cpos].pos = cpos

		/* Save parent */
		superChunkMap[scpos].chunkMap[cpos].parent = superChunkMap[scpos]

		superChunkMap[scpos].lock.Unlock()

		superChunkListLock.Unlock()
		superChunkMapLock.Unlock()
		return true
	}

	superChunkListLock.Unlock()
	superChunkMapLock.Unlock()
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
				mapLoadPercent = float32(ChunksMade) / float32((input*2)*(input*2)) * 100.0
				if wasmMode {
					time.Sleep(time.Nanosecond)
				} else {
					time.Sleep(time.Millisecond * 5)
				}
			}
		}
	}
}
