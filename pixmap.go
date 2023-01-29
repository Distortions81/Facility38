package main

import (
	"GameTest/consts"
	"GameTest/glob"
)

func renderPixMapCache() {
	for i := 0; i < gVisChunkTop; i++ {
		chunk := gVisChunks[i]
		chunkPos := gVisChunkPos[i]

		//Convert to superchunk size
		chunkPos = glob.XY{X: chunkPos.X / consts.SuperChunkSize, Y: chunkPos.Y / consts.SuperChunkSize}

		if chunk.NumObjects <= 0 {
			continue
		}

		/* Pre-calc camera superchunk position */
		scStartX := camStartX / consts.ChunkSize / consts.SuperChunkSize
		scStartY := camStartY / consts.ChunkSize / consts.SuperChunkSize
		scEndX := camEndX / consts.ChunkSize / consts.SuperChunkSize
		scEndy := camEndY / consts.ChunkSize / consts.SuperChunkSize

		/* Is this chunk in the SuperChunk? */
		if chunkPos.X < scStartX ||
			chunkPos.X > scEndX ||
			chunkPos.Y < scStartY ||
			chunkPos.Y > scEndy {
			continue
		}
	}
}
