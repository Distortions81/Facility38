package main

import (
	"time"
)

/* Make a superChunk */
func makeSuperChunk(pos XY) {
	defer reportPanic("makeSuperChunk")
	//Make superChunk if needed

	newPos := pos
	supChunkPos := posToSuperChunkPos(newPos)

	superChunkMapLock.Lock()  //Lock superChunk map
	superChunkListLock.Lock() //Lock superChunk list

	if superChunkMap[supChunkPos] == nil {

		/* Make new superChunk in map at pos */
		newSuperChunk := &mapSuperChunkData{}

		maxSize := superChunkTotal * superChunkTotal * 4
		newSuperChunk.itemMap = make([]byte, maxSize)
		newSuperChunk.resourceMap = make([]byte, maxSize)
		newSuperChunk.resourceDirty = true

		superChunkMap[supChunkPos] = newSuperChunk
		superChunkMap[supChunkPos].lock.Lock() //Lock chunk

		superChunkList =
			append(superChunkList, superChunkMap[supChunkPos])
		superChunkMap[supChunkPos].chunkMap = make(map[XY]*mapChunk)

		/* Save position */
		superChunkMap[supChunkPos].pos = supChunkPos

		superChunkMap[supChunkPos].lock.Unlock()
	}
	superChunkListLock.Unlock()
	superChunkMapLock.Unlock()
}

/* Make a chunk, insert into superChunk */
func makeChunk(pos XY) bool {
	defer reportPanic("makeChunk")
	//Make chunk if needed

	newPos := pos

	makeSuperChunk(pos)

	chunkPos := PosToChunkPos(newPos)
	supChunkPos := posToSuperChunkPos(newPos)

	superChunkMapLock.Lock()  //Lock superChunk map
	superChunkListLock.Lock() //Lock superChunk list

	if superChunkMap[supChunkPos].chunkMap[chunkPos] == nil {
		/* Increase chunk count */
		superChunkMap[supChunkPos].numChunks++

		/* Make a new empty chunk in the map at pos */
		superChunkMap[supChunkPos].chunkMap[chunkPos] = &mapChunk{}
		superChunkMap[supChunkPos].lock.Lock() //Lock chunk

		/* Append to chunk list */
		superChunkMap[supChunkPos].chunkList =
			append(superChunkMap[supChunkPos].chunkList, superChunkMap[supChunkPos].chunkMap[chunkPos])

		superChunkMap[supChunkPos].chunkMap[chunkPos].buildingMap = make(map[XY]*buildingData)
		superChunkMap[supChunkPos].chunkMap[chunkPos].tileMap = make(map[XY]*tileData)

		/* Terrain img */
		superChunkMap[supChunkPos].chunkMap[chunkPos].terrainLock.Lock()
		superChunkMap[supChunkPos].chunkMap[chunkPos].terrainImage = TempChunkImage
		superChunkMap[supChunkPos].chunkMap[chunkPos].usingTemporary = true
		superChunkMap[supChunkPos].chunkMap[chunkPos].terrainLock.Unlock()

		/* Save position */
		superChunkMap[supChunkPos].chunkMap[chunkPos].pos = chunkPos

		/* Save parent */
		superChunkMap[supChunkPos].chunkMap[chunkPos].parent = superChunkMap[supChunkPos]

		superChunkMap[supChunkPos].lock.Unlock()

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
	offX := int(pos.X) - (area / 2)
	offY := int(pos.Y) - (area / 2)
	for x := -area; x < area; x += chunkSize {
		for y := -area; y < area; y += chunkSize {
			pos := XY{X: uint16(offX - x), Y: uint16(offY - y)}
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
