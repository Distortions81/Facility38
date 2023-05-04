package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
)

var (
	chatLinesTop  int
	chatLines     []chatLineData
	chatLinesLock sync.Mutex
	buildInfo     string
)

/* Init chat system */
func init() {
	resetChat()
}

/* Handles panics */
const hdFileName = "heapDump.dat"

func reportPanic(format string, args ...interface{}) {
	if r := recover(); r != nil {

		if !wasmMode {
			doLog(false, "Writing '%v' file.", hdFileName)
			f, err := os.Create(hdFileName)
			if err == nil {
				debug.WriteHeapDump(f.Fd())
				f.Close()
				defer chatDetailed("wrote heapDump", ColorRed, time.Hour)
			} else {
				doLog(false, "Failed to write '%v' file.", hdFileName)
			}
		}

		_, filename, line, _ := runtime.Caller(4)
		input := fmt.Sprintf(format, args...)
		buf := fmt.Sprintf("(GAME CRASH)\nBUILD:v%v-%v\nLabel:%v File: %v Line: %v\nError:%v\n\nStack Trace:\n%v\n", version, buildInfo, input, filepath.Base(filename), line, r, string(debug.Stack()))

		if !wasmMode {
			os.WriteFile("panic.log", []byte(buf), 0660)
			defer chatDetailed("wrote panic.log", ColorRed, time.Hour)
		}

		//DoLog(false, buf)
		chatDetailed(buf, ColorOrange, time.Hour)
		time.Sleep(time.Hour)
	}
}

/* Reset chat history */
func resetChat() {
	chatLinesLock.Lock()
	chatLines = []chatLineData{}
	chatLines = append(chatLines, chatLineData{
		text:      "",
		timestamp: time.Now(),
		lifetime:  time.Nanosecond,
		color:     ColorAqua,
		bgColor:   ColorToolTipBG,
	})
	chatLinesTop = 1
	chatLinesLock.Unlock()
}

/* WASM is single-thread, we use sleep to allow other threads to run */
func wasmSleep() {
	if wasmMode {
		time.Sleep(time.Millisecond * 10)
	}
}

/* Add coords */
func AddXY(a XY, b XY) XY {
	defer reportPanic("AddXY")
	return XY{X: a.X + b.X, Y: a.Y + b.Y}
}

/* Sub-object relative pos to real */
func GetSubPos(a XY, b XYs) XY {
	defer reportPanic("GetSubPos")
	return XY{X: uint16(int32(a.X) + int32(b.X)), Y: uint16(int32(a.Y) + int32(b.Y))}
}

/* Subtract coords */
func SubXY(a XY, b XY) XY {
	defer reportPanic("SubXY")
	return XY{X: a.X - b.X, Y: a.Y - b.Y}
}

/* Trim lines from chat */
func deleteOldLines() {
	defer reportPanic("deleteOldLines")
	var newLines []chatLineData
	var newTop int

	/* Delete 1 excess line each time */
	for l, line := range chatLines {
		if l < 1000 {
			newLines = append(newLines, line)
			newTop++
		}
	}
	chatLines = newLines
	chatLinesTop = newTop
}

/* Log with object details */
func objCD(b *buildingData, format string, args ...interface{}) {
	defer reportPanic("ObjCD")
	if !debugMode {
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

	objData := fmt.Sprintf("%v: %v: %v", b.obj.Unique.typeP.name, posToString(b.pos), text)

	/* Date, go file, go file line, text */
	buf := fmt.Sprintf("%v: %15v:%5v: %v", date, filepath.Base(filename), line, objData)
	chatDetailed(buf, ColorRed, time.Minute)
}

/* Default add lines to chat */
func chat(text string) {
	chatDetailed(text, color.White, time.Second*15)

}

/* Add to chat with options */
func chatDetailed(text string, color color.Color, life time.Duration) {
	if !mapGenerated.Load() {
		return
	}
	doLog(false, "Chat: "+text)

	/* Don't log until we are loaded into the game */
	if !mapGenerated.Load() {
		return
	}
	go func(text string) {
		chatLinesLock.Lock()
		deleteOldLines()

		sepLines := strings.Split(text, "\n")
		for _, sep := range sepLines {
			chatLines = append(chatLines, chatLineData{text: sep, color: color, bgColor: ColorToolTipBG, lifetime: life, timestamp: time.Now()})
			chatLinesTop++
		}

		chatLinesLock.Unlock()
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

/* Delete an object from a ObjData list, does not retain order (fast) */
func ObjListDelete(obj *ObjData) {
	defer reportPanic("ObjListDelete")

	obj.chunk.lock.Lock()
	defer obj.chunk.lock.Unlock()

	for index, item := range obj.chunk.objList {
		if item.Pos == obj.Pos {
			obj.chunk.objList[index] = obj.chunk.objList[len(obj.chunk.objList)-1]
			obj.chunk.objList = obj.chunk.objList[:len(obj.chunk.objList)-1]
			visDataDirty.Store(true)
			return
		}
	}
}

/* Pos XY to string */
func posToString(pos XY) string {
	defer reportPanic("PosToString")
	centerPos := CenterXY(pos)
	buf := fmt.Sprintf("(%v,%v)", humanize.Comma(int64((centerPos.X))), humanize.Comma(int64((centerPos.Y))))
	return buf
}

/* Convert an internal XY (unsigned) to a (0,0) center */
func CenterXY(pos XY) XYs {
	defer reportPanic("CenterXY")
	return XYs{X: int32(pos.X) - int32(xyCenter), Y: int32(pos.Y) - int32(xyCenter)}
}

/* Convert uncentered position to centered */
func UnCenterXY(pos XYs) XY {
	defer reportPanic("UnCenterXY")
	return XY{X: uint16(int32(pos.X) + int32(xyCenter)), Y: uint16(int32(pos.Y) + int32(xyCenter))}
}

/* Rotate DIR value clockwise */
func RotCW(dir uint8) uint8 {
	defer reportPanic("RotCW")
	if dir == DIR_ANY {
		return DIR_ANY
	}
	return uint8(PosIntMod(int(dir+1), DIR_MAX))
}

/* Rotate DIR value counter-clockwise */
func RotCCW(dir uint8) uint8 {
	defer reportPanic("RotCCW")
	if dir == DIR_ANY {
		return DIR_ANY
	}
	return uint8(PosIntMod(int(dir-1), DIR_MAX))
}

/* Rotate DIR value to x*/
func RotDir(dir uint8, add uint8) uint8 {
	defer reportPanic("RotDir")
	if dir == DIR_ANY || add == DIR_ANY {
		return DIR_ANY
	}
	return uint8(PosIntMod(int(dir-add), DIR_MAX))
}

/* give distance between two coordinates */
func Distance(xa, ya, xb, yb int) float32 {
	defer reportPanic("Distance")
	x := math.Abs(float64(xa - xb))
	y := math.Abs(float64(ya - yb))
	return float32(math.Sqrt(x*x + y*y))
}

/* Find point directly in the middle of two coordinates */
func MidPoint(x1, y1, x2, y2 int) (int, int) {
	defer reportPanic("MidPoint")
	return (x1 + x2) / 2, (y1 + y2) / 2
}

/* Get an object by XY, uses map (hash table). RLocks the given chunk */
func GetObj(pos XY, chunk *mapChunk) *buildingData {
	defer reportPanic("GetObj")

	if chunk != nil {
		chunk.lock.RLock()
		o := chunk.buildingMap[pos]
		chunk.lock.RUnlock()

		if o != nil {
			return o
		}
	}

	return nil
}

/* Get a chunk by XY, used map (hash table). RLocks the SuperChunkMap and Chunk */
func GetChunk(pos XY) *mapChunk {
	defer reportPanic("GetChunk")

	supChunkPos := posToSuperChunkPos(pos)
	chunkPos := PosToChunkPos(pos)

	superChunkMapLock.RLock()
	sChunk := superChunkMap[supChunkPos]
	superChunkMapLock.RUnlock()

	if sChunk == nil {
		return nil
	}
	sChunk.lock.RLock()
	chunk := sChunk.chunkMap[chunkPos]
	sChunk.lock.RUnlock()

	return chunk
}

/* Get a superChunk by XY, used map (hash table). RLocks the SuperChunkMap and Chunk */
func GetSuperChunk(pos XY) *mapSuperChunkData {
	defer reportPanic("GetSuperChunk")
	supChunkPos := posToSuperChunkPos(pos)

	superChunkMapLock.RLock()
	sChunk := superChunkMap[supChunkPos]
	superChunkMapLock.RUnlock()

	return sChunk
}

/* XY to Chunk XY */
func PosToChunkPos(pos XY) XY {
	defer reportPanic("PosToChunkPos")
	return XY{X: pos.X / chunkSize, Y: pos.Y / chunkSize}
}

/* Chunk XY to XY */
func ChunkPosToPos(pos XY) XY {
	defer reportPanic("ChunkPosToPos")
	return XY{X: pos.X * chunkSize, Y: pos.Y * chunkSize}
}

/* XY to SuperChunk XY */
func posToSuperChunkPos(pos XY) XY {
	defer reportPanic("PosToSuperChunkPos")
	return XY{X: pos.X / maxSuperChunk, Y: pos.Y / maxSuperChunk}
}

/* SuperChunk XY to XY */
func SuperChunkPosToPos(pos XY) XY {
	defer reportPanic("SuperChunkPosToPos")
	return XY{X: pos.X * maxSuperChunk, Y: pos.Y * maxSuperChunk}
}

/* Chunk XY to SuperChunk XY */
func ChunkPosToSuperChunkPos(pos XY) XY {
	defer reportPanic("ChunkPosToSuperChunkPos")
	return XY{X: pos.X / superChunkSize, Y: pos.Y / superChunkSize}
}

/* SuperChunk XY to Chunk XY */
func SuperChunkPosToChunkPos(pos XY) XY {
	defer reportPanic("SuperChunkPosToChunkPos")
	return XY{X: pos.X * superChunkSize, Y: pos.Y * superChunkSize}
}

/* Float (X, Y) to XY (int) */
func FloatXYToPosition(x float32, y float32) XY {
	defer reportPanic("FloatXYToPosition")
	return XY{X: uint16(x), Y: uint16(y)}
}

/* Search SuperChunk->Chunk->ObjMap hash tables to find neighboring objects in (dir) */
func getNeighborObj(src XY, dir uint8) *buildingData {
	defer reportPanic("GetNeighborObj")
	pos := src

	switch dir {
	case DIR_NORTH:
		pos.Y--
	case DIR_EAST:
		pos.X++
	case DIR_SOUTH:
		pos.Y++
	case DIR_WEST:
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
	if b.pos == src {
		return nil
	}
	return b
}

/* Convert DIR to text */
func dirToName(dir uint8) string {
	defer reportPanic("DirToName")
	switch dir {
	case DIR_NORTH:
		return "North"
	case DIR_EAST:
		return "East"
	case DIR_SOUTH:
		return "South"
	case DIR_WEST:
		return "West"
	case DIR_ANY:
		return "Any"
	}

	return "Error"
}

/* Used in debug text */
func DirToArrow(dir uint8) string {
	defer reportPanic("DirToArrow")
	switch dir {
	case DIR_NORTH:
		return "^"
	case DIR_EAST:
		return ">"
	case DIR_SOUTH:
		return "v"
	case DIR_WEST:
		return "<"
	case DIR_ANY:
		return "*"
	}

	return "Error"
}

/* Reverse Port Direction/Type */
func reverseType(t uint8) uint8 {
	defer reportPanic("ReverseType")
	switch t {
	case PORT_OUT:
		return PORT_IN
	case PORT_IN:
		return PORT_OUT
	case PORT_FIN:
		return PORT_FOUT
	case PORT_FOUT:
		return PORT_FIN
	default:
		return PORT_NONE
	}
}

/* Flop a direction */
func reverseDirection(dir uint8) uint8 {
	defer reportPanic("ReverseDirection")
	switch dir {
	case DIR_NORTH:
		return DIR_SOUTH
	case DIR_EAST:
		return DIR_WEST
	case DIR_SOUTH:
		return DIR_NORTH
	case DIR_WEST:
		return DIR_EAST
	}

	return DIR_MAX
}

/* Generic unzip []byte */
func UncompressZip(data []byte) []byte {
	defer reportPanic("UncompressZip")
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
	defer reportPanic("CompressZip")
	var b bytes.Buffer
	w, err := zlib.NewWriterLevel(&b, zlib.BestSpeed)
	if err != nil {
		doLog(true, "CompressZip: %v", err)
	}
	w.Write(data)
	w.Close()
	return b.Bytes()
}

/* Bool to text */
func BoolToOnOff(input bool) string {
	defer reportPanic("BoolToOnOff")
	if input {
		return "On"
	} else {
		return "Off"
	}
}

/* Check if a position is within a image.Rectangle */
func PosWithinRect(pos XY, rect image.Rectangle, pad uint16) bool {
	defer reportPanic("PosWithinRect")
	if int(pos.X-pad) <= rect.Max.X && int(pos.X+pad) >= rect.Min.X {
		if int(pos.Y-pad) <= rect.Max.Y && int(pos.Y+pad) >= rect.Min.Y {
			return true
		}
	}
	return false
}
