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

/* Update superchunk/chunk/onj lists */
func UnsafeMakeObjLists() {

	/* Make superchunk list */
	index := 0
	var schunkTemp [10000]*glob.MapSuperChunk
	for _, sChunk := range glob.SuperChunkMap {
		if sChunk == nil {
			continue
		}
		schunkTemp[index] = sChunk
		index++
	}
	glob.SuperChunkList = schunkTemp[:index]

	/* Make chunk lists in all SuperChunks */
	var chunkTemp [10000]*glob.MapChunk
	for _, sChunk := range glob.SuperChunkMap {
		index = 0
		for _, chunk := range sChunk.ChunkMap {
			if chunk == nil {
				continue
			}
			chunkTemp[index] = chunk
			index++
		}
		sChunk.ChunkList = chunkTemp[:index]
	}

	/* Make obj lists in all chunks */
	var objTemp [10000]*glob.ObjData
	for _, sChunk := range glob.SuperChunkMap {
		for _, chunk := range sChunk.ChunkMap {
			index = 0
			for _, obj := range chunk.ObjMap {
				if obj == nil {
					continue
				}
				objTemp[index] = obj
				index++
			}
			chunk.ObjList = objTemp[:index]
		}
	}
}

/* Update event tick/tock lists */
func UnsafeMakeEventLists() {

}

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
	chunk := unsafeGetChunk(pos)

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
	unsafeLinkObj(pos, obj, dir)

	return obj
}

/* Link inputs and outputs, with output direction (newdir) */
func unsafeLinkObj(pos glob.XY, obj *glob.ObjData, newdir int) {
	unsafeLinkIn(pos, obj, newdir)
	unsafeLinkOut(pos, obj, newdir)
}

/* Link to output in (dir) */
func unsafeLinkOut(pos glob.XY, obj *glob.ObjData, dir int) {

	/* Don't bother if we don't have outputs */
	if !obj.TypeP.HasMatOutput {
		return
	}

	/* Look for object in output direction */
	neigh, _ := util.GetNeighborObj(obj, pos, dir)

	/* Did we find and obj? */
	if neigh == nil {
		return
	}
	/* Does it have inputs? */
	if neigh.TypeP.HasMatInput == 0 {
		return
	}
	/* Do they have an output? */
	if neigh.TypeP.HasMatOutput {
		/* Are we trying to connect from that direction? */
		if neigh.TypeP.Direction == util.ReverseDirection(dir) {
			return
		}
	}

	/* If we have an output already, unlink it */
	if obj.OutputObj != nil {
		/* Unlink OLD output specifically */
		unlinkOut(obj.OutputObj)
	}

	/* Make sure the object has an input initialized */
	if neigh.InputBuffer[util.ReverseDirection(dir)] != nil {
		neigh.InputBuffer[util.ReverseDirection(dir)] = &glob.MatData{}
	}

	/* Make sure our output is initalized */
	if obj.OutputBuffer == nil {
		obj.OutputBuffer = &glob.MatData{}
	}

	/* Mark target as our output */
	obj.OutputObj = neigh

	/* Put ourself in target's input list */
	neigh.InputObjs[util.ReverseDirection(dir)] = obj
}

/* Find and link inputs, set ourself to OutputObj of found objects */
func unsafeLinkIn(pos glob.XY, obj *glob.ObjData, newdir int) {

	/* Don't bother if we don't have inputs */
	if obj.TypeP.HasMatInput == 0 {
		return
	}

	for dir := consts.DIR_NORTH; dir < consts.DIR_MAX; dir++ {

		/* Don't try to connect an input the same direction as our future output */
		/* If there is an input there, remove it */
		if obj.TypeP.HasMatOutput && dir == newdir {
			unlinkInput(obj, dir)
			continue
		}

		/* Look for neighbor object */
		neigh, _ := util.GetNeighborObj(obj, pos, dir)

		/* Did we find an object? */
		if neigh == nil {
			continue
		}

		/* Does it have an output? */
		if !neigh.TypeP.HasMatOutput {
			continue
		}

		/* Is the output unoccupied? */
		if neigh.OutputObj != nil {
			/* Is it us? */
			if neigh.OutputObj != obj {
				continue
			}
		}

		/* Is the output in our direction? */
		if neigh.Direction != util.ReverseDirection(dir) {
			continue
		}

		/* Unlink old input from this direction if it exists */
		unlinkInput(obj, dir)

		/* Make sure they have an output initalized */
		if neigh.OutputBuffer == nil {
			neigh.OutputBuffer = &glob.MatData{}
		}

		/* Make sure we have a input initalized */
		if obj.InputBuffer[dir] == nil {
			obj.InputBuffer[dir] = &glob.MatData{}
		}

		/* Set ourself as their output */
		neigh.OutputObj = obj

		/* Record who is on this input */
		obj.InputObjs[util.ReverseDirection(dir)] = neigh
	}

}

/* Search SuperChunk->Chunk->ObjMap hashtables to find neighboring objects in (dir) */
func UnsafeGetNeighborObj(src *glob.ObjData, pos glob.XY, dir int) (*glob.ObjData, glob.XY) {

	switch dir {
	case consts.DIR_NORTH:
		pos.Y--
	case consts.DIR_EAST:
		pos.X++
	case consts.DIR_SOUTH:
		pos.Y++
	case consts.DIR_WEST:
		pos.X--
	default:
		return nil, glob.XY{}
	}

	chunk := unsafeGetChunk(pos)
	if chunk == nil {
		return nil, glob.XY{}
	}
	obj := unsafeGetObj(pos, chunk)
	if obj == nil {
		return nil, glob.XY{}
	}
	return obj, pos
}

/* UNSAFE, NO LOCKS: Get a chunk by XY, used map (hashtable). RLocks the SuperChunkMap and Chunk */
func unsafeGetChunk(pos glob.XY) *glob.MapChunk {
	scpos := util.PosToSuperChunkPos(pos)
	cpos := util.PosToChunkPos(pos)

	sChunk := glob.SuperChunkMap[scpos]
	if sChunk == nil {
		return nil
	}
	chunk := sChunk.ChunkMap[cpos]

	return chunk
}

/* Get an object by XY, uses map (hashtable). RLocks the given chunk */
func unsafeGetObj(pos glob.XY, chunk *glob.MapChunk) *glob.ObjData {
	if chunk != nil {
		o := chunk.ObjMap[pos]
		return o
	} else {
		return nil
	}
}
