package util

import (
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/gv"
	"bytes"
	"compress/zlib"
	"io"
	"log"
	"math"

	"github.com/dustin/go-humanize"
)

func RotatePortsCW(obj *glob.ObjData) {
	var newPorts [gv.DIR_MAX]glob.ObjPortData
	for i := 0; i < gv.DIR_MAX; i++ {
		//Copy to array, rotated with modulo
		p := int(PosIntMod((i + 1), gv.DIR_MAX))
		newPorts[p] = *obj.Ports[i]
	}
	for i := 0; i < gv.DIR_MAX; i++ {
		//Copy back to object
		obj.Ports[i] = &newPorts[i]
	}
}

func PosIntMod(d, m int) int {
	var res int = d % m
	if res < 0 && m > 0 {
		return res + m
	}
	return res
}

func RotatePortsCCW(obj *glob.ObjData) {
	var newPorts [gv.DIR_MAX]glob.ObjPortData
	for i := 0; i < gv.DIR_MAX; i++ {
		//Copy to array, rotated with modulo
		p := int(PosIntMod((i - 1), gv.DIR_MAX))
		newPorts[p] = *obj.Ports[i]
	}
	for i := 0; i < gv.DIR_MAX; i++ {
		//Copy back to object
		obj.Ports[i] = &newPorts[i]
	}
}

func ObjHasPort(obj *glob.ObjData, portDir uint8) bool {
	for p, _ := range obj.Ports {
		if obj.TypeP.Ports[p] == portDir {
			return true
		}
	}
	return false
}

/* Delete an object from a glob.ObjData list, does not retain order (fast) */
func ObjListDelete(obj *glob.ObjData) {
	oPos := CenterXY(obj.Pos)

	for index, item := range obj.Parent.ObjList {
		if item.Pos == obj.Pos {
			cwlog.DoLog("ObjListDelete: Deleted %v at index: %v (%v,%v)", obj.TypeP.Name, index, oPos.X, oPos.Y)
			obj.Parent.ObjList[index] = obj.Parent.ObjList[len(obj.Parent.ObjList)-1]
			obj.Parent.ObjList = obj.Parent.ObjList[:len(obj.Parent.ObjList)-1]
			return
		}
	}

	cwlog.DoLog("ObjListDelete: %v (%v,%v) not found in chunk ObjList", obj.TypeP.Name, oPos.X, oPos.Y)
}

/* Convert an internal XY (unsigned) to a (0,0) center */
func CenterXY(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X - gv.XYCenter, Y: pos.Y - gv.XYCenter}
}

/* Rotate consts.DIR value clockwise */
func RotCW(dir uint8) uint8 {
	return uint8(PosIntMod(int(dir+1), gv.DIR_MAX))
}

/* Rotate consts.DIR value counter-clockwise */
func RotCCW(dir uint8) uint8 {
	return uint8(PosIntMod(int(dir-1), gv.DIR_MAX))
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
	return glob.XY{X: pos.X / gv.ChunkSize, Y: pos.Y / gv.ChunkSize}
}

/* Chunk XY to XY */
func ChunkPosToPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X * gv.ChunkSize, Y: pos.Y * gv.ChunkSize}
}

/* XY to SuperChunk XY */
func PosToSuperChunkPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X / gv.MaxSuperChunk, Y: pos.Y / gv.MaxSuperChunk}
}

/* SuperChunk XY to XY */
func SuperChunkPosToPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X * gv.MaxSuperChunk, Y: pos.Y * gv.MaxSuperChunk}
}

/* Chunk XY to SuperChunk XY */
func ChunkPosToSuperChunkPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X / gv.SuperChunkSize, Y: pos.Y / gv.SuperChunkSize}
}

/* SuperChunk XY to Chunk XY */
func SuperChunkPosToChunkPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X * gv.SuperChunkSize, Y: pos.Y * gv.SuperChunkSize}
}

/* Float (X, Y) to glob.XY (int) */
func FloatXYToPosition(x float64, y float64) glob.XY {

	return glob.XY{X: int(x), Y: int(y)}
}

/* Search SuperChunk->Chunk->ObjMap hashtables to find neighboring objects in (dir) */
func GetNeighborObj(src *glob.ObjData, dir uint8) *glob.ObjData {

	pos := src.Pos

	switch dir {
	case gv.DIR_NORTH:
		pos.Y--
	case gv.DIR_EAST:
		pos.X++
	case gv.DIR_SOUTH:
		pos.Y++
	case gv.DIR_WEST:
		pos.X--
	default:
		return nil
	}

	chunk := GetChunk(pos)
	if chunk == nil {
		return nil
	}
	obj := GetObj(pos, chunk)
	if obj == nil {
		return nil
	}
	return obj
}

/* Convert consts.DIR to text */
func DirToName(dir uint8) string {
	switch dir {
	case gv.DIR_NORTH:
		return "North"
	case gv.DIR_EAST:
		return "East"
	case gv.DIR_SOUTH:
		return "South"
	case gv.DIR_WEST:
		return "West"
	}

	return "Error"
}

/* Flop a consts.DIR */
func ReverseDirection(dir uint8) uint8 {
	switch dir {
	case gv.DIR_NORTH:
		return gv.DIR_SOUTH
	case gv.DIR_EAST:
		return gv.DIR_WEST
	case gv.DIR_SOUTH:
		return gv.DIR_NORTH
	case gv.DIR_WEST:
		return gv.DIR_EAST
	}

	return gv.DIR_MAX
}

func ReversePort(port uint8) uint8 {
	if port == gv.PORT_INPUT {
		return gv.PORT_OUTPUT
	} else if port == gv.PORT_OUTPUT {
		return gv.PORT_INPUT
	}
	return gv.PORT_NONE
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
