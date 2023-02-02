package util

import (
	"GameTest/consts"
	"GameTest/cwlog"
	"GameTest/glob"
	"bytes"
	"compress/zlib"
	"io"
	"log"
	"math"

	"github.com/dustin/go-humanize"
)

/* Delete an object from a glob.ObjData list, does not retain order (fast) */
func ObjListDelete(obj *glob.ObjData) {
	for index, item := range obj.Parent.ObjList {
		if item.Pos == obj.Pos {
			obj.Parent.ObjList[index] = obj.Parent.ObjList[len(obj.Parent.ObjList)-1]
			obj.Parent.ObjList = obj.Parent.ObjList[:len(obj.Parent.ObjList)-1]
			break
		}
	}
}

/* Convert an internal XY (unsigned) to a (0,0) center */
func CenterXY(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X - consts.XYCenter, Y: pos.Y - consts.XYCenter}
}

/* Rotate consts.DIR value clockwise */
func RotCW(dir int) int {
	dir = dir - 1
	if dir < consts.DIR_NORTH {
		dir = consts.DIR_WEST
	}
	return dir
}

/* Rotate consts.DIR value counter-clockwise */
func RotCCW(dir int) int {
	dir = dir + 1
	if dir > consts.DIR_WEST {
		dir = consts.DIR_NORTH
	}
	return dir
}

/* give distance between two coordinates */
func Distance(xa, ya, xb, yb int) float64 {
	x := math.Abs(float64(xa - xb))
	y := math.Abs(float64(ya - yb))
	return math.Sqrt(x*x + y*y)
}

/* Find point directly in the middle of two coordinates */
func MidPoint(x1, y1, x2, y2 int) (int, int) {
	return (x1 + x2) / 2, (y1 + y2) / 2
}

/* Get an object by XY, uses map (hashtable). RLocks the given chunk */
func GetObj(pos glob.XY, chunk *glob.MapChunk) *glob.ObjData {
	if chunk != nil {
		chunk.Lock.RLock()
		o := chunk.ObjMap[pos]
		chunk.Lock.RUnlock()
		return o
	} else {
		return nil
	}
}

/* Get a chunk by XY, used map (hashtable). RLocks the SuperChunkMap and Chunk */
func GetChunk(pos glob.XY) *glob.MapChunk {
	scpos := PosToSuperChunkPos(pos)
	cpos := PosToChunkPos(pos)

	glob.SuperChunkMapLock.RLock()
	sChunk := glob.SuperChunkMap[scpos]
	glob.SuperChunkMapLock.RUnlock()

	if sChunk == nil {
		return nil
	}
	sChunk.Lock.RLock()
	chunk := sChunk.ChunkMap[cpos]
	sChunk.Lock.RUnlock()

	return chunk
}

/* UNSAFE, NO LOCKS: Get a chunk by XY, used map (hashtable). RLocks the SuperChunkMap and Chunk */
func UnsafeGetChunk(pos glob.XY) *glob.MapChunk {
	scpos := PosToSuperChunkPos(pos)
	cpos := PosToChunkPos(pos)

	sChunk := glob.SuperChunkMap[scpos]
	if sChunk == nil {
		return nil
	}
	chunk := sChunk.ChunkMap[cpos]

	return chunk
}

/* Get a superchunk by XY, used map (hashtable). RLocks the SuperChunkMap and Chunk */
func GetSuperChunk(pos glob.XY) *glob.MapSuperChunk {
	scpos := PosToChunkPos(pos)

	glob.SuperChunkMapLock.RLock()
	sChunk := glob.SuperChunkMap[scpos]
	glob.SuperChunkMapLock.RUnlock()

	return sChunk
}

/* XY to Chunk XY */
func PosToChunkPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X / consts.ChunkSize, Y: pos.Y / consts.ChunkSize}
}

/* Chunk XY to XY */
func ChunkPosToPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X * consts.ChunkSize, Y: pos.Y * consts.ChunkSize}
}

/* XY to SuperChunk XY */
func PosToSuperChunkPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X / consts.SuperChunkPixels, Y: pos.Y / consts.SuperChunkPixels}
}

/* SuperChunk XY to XY */
func SuperChunkPosToPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X * consts.SuperChunkPixels, Y: pos.Y * consts.SuperChunkPixels}
}

/* Chunk XY to SuperChunk XY */
func ChunkPosToSuperChunkPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X / consts.SuperChunkSize, Y: pos.Y / consts.SuperChunkSize}
}

/* SuperChunk XY to Chunk XY */
func SuperChunkPosToChunkPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X * consts.SuperChunkSize, Y: pos.Y * consts.SuperChunkSize}
}

/* Float (X, Y) to glob.XY (int) */
func FloatXYToPosition(x float64, y float64) glob.XY {

	return glob.XY{X: int(x), Y: int(y)}
}

/* Search SuperChunk->Chunk->ObjMap hashtables to find neighboring objects in (dir) */
func GetNeighborObj(src *glob.ObjData, pos glob.XY, dir int) (*glob.ObjData, glob.XY) {

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

	chunk := GetChunk(pos)
	if chunk == nil {
		return nil, glob.XY{}
	}
	obj := GetObj(pos, chunk)
	if obj == nil {
		return nil, glob.XY{}
	}
	return obj, pos
}

/* Convert consts.DIR to text */
func DirToName(dir int) string {
	switch dir {
	case consts.DIR_NORTH:
		return "North"
	case consts.DIR_EAST:
		return "East"
	case consts.DIR_SOUTH:
		return "South"
	case consts.DIR_WEST:
		return "West"
	}

	return "Error"
}

/* Flop a consts.DIR */
func ReverseDirection(dir int) int {
	switch dir {
	case consts.DIR_NORTH:
		return consts.DIR_SOUTH
	case consts.DIR_EAST:
		return consts.DIR_WEST
	case consts.DIR_SOUTH:
		return consts.DIR_NORTH
	case consts.DIR_WEST:
		return consts.DIR_EAST
	}

	return consts.DIR_MAX
}

/* Generic unzip []byte */
func UncompressZip(data []byte) []byte {

	b := bytes.NewReader(data)

	log.Println("Uncompressing: ", humanize.Bytes(uint64(len(data))))
	z, err := zlib.NewReader(b)
	if err != nil {
		log.Println("Error: ", err)
		return nil
	}
	defer z.Close()

	p, err := io.ReadAll(z)
	if err != nil {
		log.Println("Error: ", err)
		return nil
	}
	log.Print("Uncompressed: ", humanize.Bytes(uint64(len(p))))
	return p
}

/* Generic zip []byte */
func CompressZip(data []byte) []byte {
	var b bytes.Buffer
	w, err := zlib.NewWriterLevel(&b, zlib.BestCompression)
	if err != nil {
		cwlog.DoLog("CompressZip: %v", err)
	}
	w.Write(data)
	w.Close()
	return b.Bytes()
}
