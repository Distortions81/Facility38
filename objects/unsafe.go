package objects

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/util"
)

/*
 *
 * Only use to create or load maps when tick/tock are not running
 * Maually make superchunk, chunk, objlist, ticklist and tocklist at end for speed
 * Appending is slow
 *
 */

/* Make a super chunk if it does not exist, unsafe map load version */
func unsafeMakeSuperChunk(pos glob.XY) {

	newPos := pos
	scpos := util.PosToSuperChunkPos(newPos)

	//Make super chunk if needed

	SuperChunkTmp := glob.SuperChunkMap[scpos]

	if SuperChunkTmp == nil {
		/* Make new superchunk in map at pos */
		glob.SuperChunkMap[scpos] = &glob.MapSuperChunk{}
		SuperChunkTmp = glob.SuperChunkMap[scpos]

		SuperChunkTmp.ChunkMap = make(map[glob.XY]*glob.MapChunk)

		/* Save position */
		SuperChunkTmp.Pos = scpos
	}

}

/* Make a chunk if it does not exist, unsafe map load version */
func unsafeMakeChunk(pos glob.XY) {
	//Make chunk if needed
	unsafeMakeSuperChunk(pos)

	cpos := util.PosToChunkPos(pos)
	scpos := util.PosToSuperChunkPos(pos)

	SuperChunkTmp := glob.SuperChunkMap[scpos]
	ChunkTmp := SuperChunkTmp.ChunkMap[cpos]

	if ChunkTmp == nil {

		/* Increase chunk count */
		SuperChunkTmp.NumChunks++

		/* Make a new empty chunk in the map at pos */
		SuperChunkTmp.ChunkMap[cpos] = &glob.MapChunk{}
		ChunkTmp = SuperChunkTmp.ChunkMap[cpos]

		ChunkTmp.ObjMap = make(map[glob.XY]*glob.ObjData)

		/* Save position */
		ChunkTmp.Pos = cpos

		/* Save parent */
		ChunkTmp.Parent = SuperChunkTmp
	}
}

/* Make a obj, unsafe map load version */
func UnsafeCreateObj(pos glob.XY, mtype int, dir int) *glob.ObjData {

	//Make chunk if needed
	unsafeMakeChunk(pos)
	chunk := util.UnsafeGetChunk(pos)

	obj := &glob.ObjData{}

	obj.Pos = pos
	obj.Parent = chunk

	obj.TypeP = GameObjTypes[mtype]

	obj.Contents = [consts.MAT_MAX]*glob.MatData{}
	if obj.TypeP.HasMatOutput {
		obj.Direction = dir
	}

	obj.Parent.ObjMap[pos] = obj
	obj.Parent.Parent.PixmapDirty = true
	obj.Parent.NumObjects++
	LinkObj(pos, obj, dir)

	return obj
}
