package util

import (
	"Facility38/cwlog"
	"Facility38/def"
	"Facility38/world"
	"bytes"
	"compress/zlib"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"math"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
)

var (
	ChatLinesTop  int
	ChatLines     []world.ChatLines
	ChatLinesLock sync.Mutex
)

func init() {
	ChatLines = append(ChatLines, world.ChatLines{
		Text:      "",
		Timestamp: time.Now(),
		Lifetime:  time.Nanosecond,
		Color:     world.ColorAqua,
		BGColor:   world.ColorToolTipBG,
	})
	ChatLinesTop = 1
}

func WASMSleep() {
	if world.WASMMode {
		time.Sleep(time.Nanosecond)
	}
}

func AddXY(a world.XY, b world.XY) world.XY {
	return world.XY{X: a.X + b.X, Y: a.Y + b.Y}
}

func GetSubPos(a world.XY, b world.XYs) world.XY {
	return world.XY{X: uint16(int32(a.X) + int32(b.X)), Y: uint16(int32(a.Y) + int32(b.Y))}
}

func SubXY(a world.XY, b world.XY) world.XY {
	return world.XY{X: a.X - b.X, Y: a.Y - b.Y}
}

func deleteOldLines() {

	var newLines []world.ChatLines
	var newTop int

	/* Delete 1 excess line each time */
	for l, line := range ChatLines {
		if l < 1000 {
			newLines = append(newLines, line)
			newTop++
		}
	}
	ChatLines = newLines
	ChatLinesTop = newTop
}

func ObjCD(b *world.BuildingData, format string, args ...interface{}) {
	if !world.Debug {
		return
	}
	/* Get current time */
	ctime := time.Now()
	/* Get calling function and line */
	_, filename, line, _ := runtime.Caller(1)
	/* printf conversion */
	text := fmt.Sprintf(format, args...)
	/* Add current date */
	date := fmt.Sprintf("%2v:%2v.%2v", ctime.Hour(), ctime.Minute(), ctime.Second())

	/* Add object name and position */

	objData := fmt.Sprintf("%v: %v: %v", b.Obj.Unique.TypeP.Name, PosToString(b.Pos), text)

	/* Date, go file, go file line, text */
	buf := fmt.Sprintf("%v: %15v:%5v: %v", date, filepath.Base(filename), line, objData)
	ChatDetailed(buf, world.ColorRed, time.Minute)
}

func Chat(text string) {
	go func(text string) {
		ChatLinesLock.Lock()
		deleteOldLines()

		ChatLines = append(ChatLines, world.ChatLines{Text: text, Color: color.White, BGColor: world.ColorToolTipBG, Lifetime: time.Second * 15, Timestamp: time.Now()})
		ChatLinesTop++

		ChatLinesLock.Unlock()
		cwlog.DoLog(false, "Chat: "+text)
	}(text)
}
func ChatDetailed(text string, color color.Color, life time.Duration) {
	/* Don't log until we are loaded into the game */
	if !world.MapGenerated.Load() {
		return
	}
	go func(text string) {
		ChatLinesLock.Lock()
		deleteOldLines()

		ChatLines = append(ChatLines, world.ChatLines{Text: text, Color: color, BGColor: world.ColorToolTipBG, Lifetime: life, Timestamp: time.Now()})
		ChatLinesTop++

		ChatLinesLock.Unlock()
		cwlog.DoLog(false, "Chat: "+text)
	}(text)
}

func Min(a, b float32) float32 {
	if a <= b {
		return a
	} else {
		return b
	}
}

func Max(a, b float32) float32 {
	if a >= b {
		return a
	} else {
		return b
	}
}

func MaxI(a, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

func MinI(a, b int) int {
	if a <= b {
		return a
	} else {
		return b
	}
}

func PosIntMod(d, m int) int {
	var res int = d % m
	if res < 0 && m > 0 {
		return res + m
	}
	return res
}

/* Delete an object from a world.ObjData list, does not retain order (fast) */
func ObjListDelete(obj *world.ObjData) {

	obj.Parent.Lock.Lock()
	defer obj.Parent.Lock.Unlock()
	for index, item := range obj.Parent.ObjList {
		if item.Pos == obj.Pos {
			obj.Parent.ObjList[index] = obj.Parent.ObjList[len(obj.Parent.ObjList)-1]
			obj.Parent.ObjList = obj.Parent.ObjList[:len(obj.Parent.ObjList)-1]
			world.VisDataDirty.Store(true)
			return
		}
	}
}

func PosToString(pos world.XY) string {
	centerPos := CenterXY(pos)
	buf := fmt.Sprintf("(%v,%v)", humanize.Comma(int64((centerPos.X))), humanize.Comma(int64((centerPos.Y))))
	return buf
}

/* Convert an internal XY (unsigned) to a (0,0) center */
func CenterXY(pos world.XY) world.XYs {
	return world.XYs{X: int32(pos.X) - int32(def.XYCenter), Y: int32(pos.Y) - int32(def.XYCenter)}
}

func UnCenterXY(pos world.XYs) world.XY {
	return world.XY{X: uint16(int32(pos.X) + int32(def.XYCenter)), Y: uint16(int32(pos.Y) + int32(def.XYCenter))}
}

/* Rotate consts.DIR value clockwise */
func RotCW(dir uint8) uint8 {
	return uint8(PosIntMod(int(dir+1), def.DIR_MAX))
}

/* Rotate consts.DIR value counter-clockwise */
func RotCCW(dir uint8) uint8 {
	return uint8(PosIntMod(int(dir-1), def.DIR_MAX))
}

/* Rotate consts.DIR value to x*/
func RotDir(dir uint8, add uint8) uint8 {
	return uint8(PosIntMod(int(dir-add), def.DIR_MAX))
}

/* give distance between two coordinates */
func Distance(xa, ya, xb, yb int) float32 {
	x := math.Abs(float64(xa - xb))
	y := math.Abs(float64(ya - yb))
	return float32(math.Sqrt(x*x + y*y))
}

/* Find point directly in the middle of two coordinates */
func MidPoint(x1, y1, x2, y2 int) (int, int) {
	return (x1 + x2) / 2, (y1 + y2) / 2
}

/* Get an object by XY, uses map (hashtable). RLocks the given chunk */
func GetObj(pos world.XY, chunk *world.MapChunk) *world.BuildingData {
	if chunk != nil {
		chunk.Lock.RLock()
		o := chunk.BuildingMap[pos]
		chunk.Lock.RUnlock()
		if o != nil {
			return o
		} else {
			return nil
		}
	} else {
		return nil
	}
}

/* Get a chunk by XY, used map (hashtable). RLocks the SuperChunkMap and Chunk */
func GetChunk(pos world.XY) *world.MapChunk {
	scpos := PosToSuperChunkPos(pos)
	cpos := PosToChunkPos(pos)

	world.SuperChunkMapLock.RLock()
	sChunk := world.SuperChunkMap[scpos]
	world.SuperChunkMapLock.RUnlock()

	if sChunk == nil {
		return nil
	}
	sChunk.Lock.RLock()
	chunk := sChunk.ChunkMap[cpos]
	sChunk.Lock.RUnlock()

	return chunk
}

/* Get a superchunk by XY, used map (hashtable). RLocks the SuperChunkMap and Chunk */
func GetSuperChunk(pos world.XY) *world.MapSuperChunk {
	scpos := PosToSuperChunkPos(pos)

	world.SuperChunkMapLock.RLock()
	sChunk := world.SuperChunkMap[scpos]
	world.SuperChunkMapLock.RUnlock()

	return sChunk
}

/* XY to Chunk XY */
func PosToChunkPos(pos world.XY) world.XY {
	return world.XY{X: pos.X / def.ChunkSize, Y: pos.Y / def.ChunkSize}
}

/* Chunk XY to XY */
func ChunkPosToPos(pos world.XY) world.XY {
	return world.XY{X: pos.X * def.ChunkSize, Y: pos.Y * def.ChunkSize}
}

/* XY to SuperChunk XY */
func PosToSuperChunkPos(pos world.XY) world.XY {
	return world.XY{X: pos.X / def.MaxSuperChunk, Y: pos.Y / def.MaxSuperChunk}
}

/* SuperChunk XY to XY */
func SuperChunkPosToPos(pos world.XY) world.XY {
	return world.XY{X: pos.X * def.MaxSuperChunk, Y: pos.Y * def.MaxSuperChunk}
}

/* Chunk XY to SuperChunk XY */
func ChunkPosToSuperChunkPos(pos world.XY) world.XY {
	return world.XY{X: pos.X / def.SuperChunkSize, Y: pos.Y / def.SuperChunkSize}
}

/* SuperChunk XY to Chunk XY */
func SuperChunkPosToChunkPos(pos world.XY) world.XY {
	return world.XY{X: pos.X * def.SuperChunkSize, Y: pos.Y * def.SuperChunkSize}
}

/* Float (X, Y) to world.XY (int) */
func FloatXYToPosition(x float32, y float32) world.XY {

	return world.XY{X: uint16(x), Y: uint16(y)}
}

/* Search SuperChunk->Chunk->ObjMap hashtables to find neighboring objects in (dir) */
func GetNeighborObj(src world.XY, dir uint8) *world.BuildingData {

	pos := src

	switch dir {
	case def.DIR_NORTH:
		pos.Y--
	case def.DIR_EAST:
		pos.X++
	case def.DIR_SOUTH:
		pos.Y++
	case def.DIR_WEST:
		pos.X--
	default:
		return nil
	}

	chunk := GetChunk(pos)
	if chunk == nil {
		return nil
	}
	b := GetObj(pos, chunk)
	if b == nil {
		return nil
	}
	if b.Pos == src {
		return nil
	}
	return b
}

/* Convert consts.DIR to text */
func DirToName(dir uint8) string {
	switch dir {
	case def.DIR_NORTH:
		return "North"
	case def.DIR_EAST:
		return "East"
	case def.DIR_SOUTH:
		return "South"
	case def.DIR_WEST:
		return "West"
	}

	return "Error"
}

func DirToArrow(dir uint8) string {
	switch dir {
	case def.DIR_NORTH:
		return "^"
	case def.DIR_EAST:
		return ">"
	case def.DIR_SOUTH:
		return "v"
	case def.DIR_WEST:
		return "<"
	}

	return "Error"
}

func ReverseType(t uint8) uint8 {
	switch t {
	case def.PORT_OUT:
		return def.PORT_IN
	case def.PORT_IN:
		return def.PORT_OUT
	case def.PORT_FIN:
		return def.PORT_FOUT
	case def.PORT_FOUT:
		return def.PORT_FIN
	default:
		return def.PORT_NONE
	}
}

/* Flop a consts.DIR */
func ReverseDirection(dir uint8) uint8 {
	switch dir {
	case def.DIR_NORTH:
		return def.DIR_SOUTH
	case def.DIR_EAST:
		return def.DIR_WEST
	case def.DIR_SOUTH:
		return def.DIR_NORTH
	case def.DIR_WEST:
		return def.DIR_EAST
	}

	return def.DIR_MAX
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
	w, err := zlib.NewWriterLevel(&b, zlib.BestSpeed)
	if err != nil {
		cwlog.DoLog(true, "CompressZip: %v", err)
	}
	w.Write(data)
	w.Close()
	return b.Bytes()
}

func BoolToOnOff(input bool) string {
	if input {
		return "On"
	} else {
		return "Off"
	}
}

func PosWithinRect(pos world.XY, rect image.Rectangle, pad uint16) bool {
	if int(pos.X-pad) <= rect.Max.X && int(pos.X+pad) >= rect.Min.X {
		if int(pos.Y-pad) <= rect.Max.Y && int(pos.Y+pad) >= rect.Min.Y {
			return true
		}
	}
	return false
}
