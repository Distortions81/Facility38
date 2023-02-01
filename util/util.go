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

func CenterXY(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X - consts.XYCenter, Y: pos.Y - consts.XYCenter}
}

func RotCW(dir int) int {
	dir = dir - 1
	if dir < consts.DIR_NORTH {
		dir = consts.DIR_WEST
	}
	return dir
}
func RotCCW(dir int) int {
	dir = dir + 1
	if dir > consts.DIR_WEST {
		dir = consts.DIR_NORTH
	}
	return dir
}

func Distance(xa, ya, xb, yb int) float64 {
	x := math.Abs(float64(xa - xb))
	y := math.Abs(float64(ya - yb))
	return math.Sqrt(x*x + y*y)
}

func MidPoint(x1, y1, x2, y2 int) (int, int) {
	return (x1 + x2) / 2, (y1 + y2) / 2
}

func GetObj(pos glob.XY, chunk *glob.MapChunk) *glob.WObject {
	if chunk != nil {
		o := chunk.WObject[pos]
		return o
	} else {
		return nil
	}
}

// Automatically converts position to chunk format
func GetChunk(pos glob.XY) *glob.MapChunk {
	scpos := PosToSuperChunkPos(pos)
	cpos := PosToChunkPos(pos)

	sChunk := glob.SuperChunkMap[scpos]
	if sChunk == nil {
		return nil
	}
	chunk := sChunk.Chunks[cpos]
	return chunk
}

// Automatically converts position to superChunk format
func GetSuperChunk(pos glob.XY) *glob.MapSuperChunk {
	scpos := PosToChunkPos(pos)
	sChunk := glob.SuperChunkMap[scpos]
	return sChunk
}

func PosToChunkPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X / consts.ChunkSize, Y: pos.Y / consts.ChunkSize}
}
func ChunkPosToPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X * consts.ChunkSize, Y: pos.Y * consts.ChunkSize}
}

func PosToSuperChunkPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X / consts.SuperChunkPixels, Y: pos.Y / consts.SuperChunkPixels}
}
func SuperChunkPosToPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X * consts.SuperChunkPixels, Y: pos.Y * consts.SuperChunkPixels}
}

func ChunkPosToSuperChunkPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X / consts.SuperChunkSize, Y: pos.Y / consts.SuperChunkSize}
}
func SuperChunkPosToChunkPos(pos glob.XY) glob.XY {
	return glob.XY{X: pos.X * consts.SuperChunkSize, Y: pos.Y * consts.SuperChunkSize}
}

func FloatXYToPosition(x float64, y float64) glob.XY {

	return glob.XY{X: int(x), Y: int(y)}
}

func GetNeighborObj(src *glob.WObject, pos glob.XY, dir int) (*glob.WObject, glob.XY) {

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
	/* We are not our own neighbor */
	if src == chunk.WObject[pos] {
		return nil, glob.XY{}
	}
	return obj, pos
}

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
	case consts.DIR_UP:
		return consts.DIR_DOWN
	case consts.DIR_DOWN:
		return consts.DIR_UP
	}

	return consts.DIR_NONE
}

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
