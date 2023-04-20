package main

import (
	"Facility38/gv"
	"Facility38/objects"
	"Facility38/util"
	"Facility38/world"
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	maxTerrainCache     = 500
	maxTerrainCacheWASM = 50
	minTerrainTime      = time.Minute
	terrainRenderLoop   = time.Millisecond
	debugVisualize      = false

	maxPixmapCache     = 500
	maxPixmapCacheWASM = 50
	minPixmapTime      = time.Minute
	pixmapRenderLoop   = time.Millisecond * 100
	resourceRenderLoop = time.Second
)

var (
	numTerrainCache int
	numPixmapCache  int
)

/* Make a 'loading' temporary texture for chunk terrain */
func SetupTerrainCache() {

	tChunk := world.MapChunk{}
	renderChunkGround(&tChunk, false, world.XY{X: 0, Y: 0})
	world.TempChunkImage = tChunk.TerrainImage

	world.SuperChunkListLock.RLock()
	for _, sChunk := range world.SuperChunkList {
		for _, chunk := range sChunk.ChunkList {
			killTerrainCache(chunk, true)
		}
	}
	world.SuperChunkListLock.RUnlock()

	if debugVisualize {
		world.TempChunkImage.Fill(world.ColorDarkRed)
	}
}

/* Render a chunk's terrain to chunk.TerrainImg, locks chunk.TerrainLock */
func renderChunkGround(chunk *world.MapChunk, doDetail bool, cpos world.XY) {
	chunkPix := (gv.SpriteScale * gv.ChunkSize)

	var bg *ebiten.Image = objects.TerrainTypes[0].Images.Main
	sx := int(float32(bg.Bounds().Size().X))
	sy := int(float32(bg.Bounds().Size().Y))
	var tImg *ebiten.Image

	if sx > 0 && sy > 0 {

		rect := image.Rectangle{}

		rect.Max.X = chunkPix
		rect.Max.Y = chunkPix

		tImg = ebiten.NewImageWithOptions(rect, &ebiten.NewImageOptions{Unmanaged: true})

		var opList [gv.ChunkSize * gv.ChunkSize]*ebiten.DrawImageOptions
		var opPos uint16

		for i := 0; i < gv.ChunkSize; i++ {
			for j := 0; j < gv.ChunkSize; j++ {
				var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
				op.GeoM.Translate(float64(i*sx), float64(j*sy))

				if doDetail {
					x := (float32(cpos.X*gv.ChunkSize) + float32(i))
					y := (float32(cpos.Y*gv.ChunkSize) + float32(j))

					h := objects.NoiseMap(x, y, 0)

					op.ColorScale.Reset()
					op.ColorScale.Scale((h)-1.5, (h)-1.5, (h)-1.5, 1)
				} else {
					op.ColorScale.Reset()
					op.ColorScale.Scale(0.4, 0.4, 0.4, 1)
				}
				opList[opPos] = op
				opPos++
			}
		}
		/* Batch render */
		for _, op := range opList {
			tImg.DrawImage(bg, op)
		}

	} else {
		panic("No valid bg texture.")
	}

	chunk.TerrainLock.Lock()

	numTerrainCache++
	chunk.TerrainImage = tImg
	chunk.UsingTemporary = false
	chunk.TerrainTime = time.Now()
	chunk.TerrainLock.Unlock()
}

var clearedCache bool

/* WASM single-thread version, one tile per call */
func RenderTerrainST() {

	/* If we zoom out, decallocate everything */
	if world.ZoomScale <= gv.MapPixelThreshold {
		if gv.WASMMode && !clearedCache {
			for _, sChunk := range world.SuperChunkList {
				for _, chunk := range sChunk.ChunkList {
					killTerrainCache(chunk, true)
				}
			}
			clearedCache = true
		}
	} else {
		clearedCache = false

		world.SuperChunkListLock.RLock()
		for _, chunk := range VisChunk {
			if chunk.UsingTemporary {
				renderChunkGround(chunk, true, chunk.Pos)
			}
		}
		world.SuperChunkListLock.RUnlock()

		/* Kill non-visible */
		for _, sChunk := range world.SuperChunkList {
			for _, chunk := range sChunk.ChunkList {
				if !chunk.Visible && !chunk.UsingTemporary {
					killTerrainCache(chunk, false)
				}
			}
		}
	}
}

/* Dispose terrain cache in a chunk if needed. Always dispose: force. Locks chunk.TerrainLock */
func killTerrainCache(chunk *world.MapChunk, force bool) {

	if chunk.UsingTemporary || chunk.TerrainImage == nil {
		return
	}

	if force ||
		(numTerrainCache > maxTerrainCache &&
			time.Since(chunk.TerrainTime) > minTerrainTime) ||
		(gv.WASMMode && numTerrainCache > maxTerrainCacheWASM) {

		chunk.TerrainLock.Lock()
		chunk.TerrainImage.Dispose()
		chunk.TerrainImage = world.TempChunkImage
		chunk.UsingTemporary = true
		numTerrainCache--
		chunk.TerrainLock.Unlock()

	}
}

/* Render pixmap images, one tile per call. Also disposes if zoom level changes. */
func PixmapRenderST() {

	if !world.ShowResourceLayer && world.ZoomScale > gv.MapPixelThreshold && !pixmapCacheCleared {

		for _, sChunk := range world.SuperChunkList {
			if sChunk.PixelMap != nil {

				sChunk.PixelMap.Dispose()
				sChunk.PixelMap = nil
				numPixmapCache--
				break

			}
		}
		pixmapCacheCleared = true

	} else if world.ZoomScale <= gv.MapPixelThreshold || world.ShowResourceLayer {
		pixmapCacheCleared = false

		for _, sChunk := range world.SuperChunkList {
			if sChunk.PixelMap == nil || sChunk.PixmapDirty {
				drawPixmap(sChunk, sChunk.Pos)
				break
			}

		}
	}
}

var pixmapCacheCleared bool

/* Loop, renders and disposes superchunk to sChunk.PixMap Locks sChunk.PixLock */
func PixmapRenderDaemon() {

	for objects.GameRunning {
		time.Sleep(pixmapRenderLoop)

		world.SuperChunkListLock.RLock()
		for _, sChunk := range world.SuperChunkList {
			if sChunk.NumChunks == 0 {
				continue
			}

			if !world.ShowResourceLayer && world.ZoomScale > gv.MapPixelThreshold && !pixmapCacheCleared {

				pixmapCacheCleared = true
				sChunk.PixelMapLock.Lock()
				if sChunk.PixelMap != nil &&
					(maxPixmapCache > numPixmapCache ||
						(gv.WASMMode && maxPixmapCacheWASM > numPixmapCache)) {

					sChunk.PixelMap.Dispose()
					sChunk.PixelMap = nil
					numPixmapCache--

				}
				sChunk.PixelMapLock.Unlock()
			} else if world.ZoomScale <= gv.MapPixelThreshold || world.ShowResourceLayer {
				pixmapCacheCleared = false

				if sChunk.PixelMap == nil || sChunk.PixmapDirty {
					drawPixmap(sChunk, sChunk.Pos)
				}
			}
		}
		world.SuperChunkListLock.RUnlock()
	}
}

/* Loop, renders and disposes superchunk to sChunk.PixMap Locks sChunk.PixLock */
func ResourceRenderDaemon() {

	for objects.GameRunning {

		world.SuperChunkListLock.RLock()
		for _, sChunk := range world.SuperChunkList {
			sChunk.ResourceLock.Lock()
			if sChunk.ResourceMap == nil || sChunk.ResourceDirty {
				drawResource(sChunk)
				sChunk.ResourceDirty = false
			}
			sChunk.ResourceLock.Unlock()
		}
		world.SuperChunkListLock.RUnlock()

		time.Sleep(resourceRenderLoop)
	}
}

func ResourceRenderDaemonST() {

	for _, sChunk := range world.SuperChunkList {
		if sChunk.ResourceMap == nil || sChunk.ResourceDirty {
			drawResource(sChunk)
			sChunk.ResourceDirty = false
			break
		}
	}
}

func drawResource(sChunk *world.MapSuperChunk) {
	if sChunk == nil {
		return
	}

	if sChunk.ResourceMap == nil {
		sChunk.ResourceMap = make([]byte, gv.SuperChunkTotal*gv.SuperChunkTotal*4)
	}

	for x := 0; x < gv.SuperChunkTotal; x++ {
		for y := 0; y < gv.SuperChunkTotal; y++ {
			ppos := 4 * (x + y*gv.SuperChunkTotal)

			worldX := float32((sChunk.Pos.X * gv.SuperChunkTotal) + uint16(x))
			worldY := float32((sChunk.Pos.Y * gv.SuperChunkTotal) + uint16(y))

			var r, g, b float32 = 0.01, 0.01, 0.01
			for p, nl := range objects.NoiseLayers {
				if p == 0 {
					continue
				}

				h := objects.NoiseMap(worldX, worldY, p)

				Chunk := sChunk.ChunkMap[util.PosToChunkPos(world.XY{X: uint16(worldX), Y: uint16(worldY)})]
				if Chunk != nil {
					Tile := Chunk.TileMap[world.XY{X: uint16(x), Y: uint16(y)}]

					if Tile != nil {
						h -= (Tile.MinerData.Mined[p] / 150)
					}
				}
				if nl.ModRed {
					r += (h * nl.RedMulti)
				}
				if nl.ModGreen {
					g += (h * nl.GreenMulti)
				}
				if nl.ModBlue {
					b += (h * nl.BlueMulti)
				}
			}
			r = util.Min(r, 1.0)
			g = util.Min(g, 1.0)
			b = util.Min(b, 1.0)

			r = util.Max(r, 0)
			g = util.Max(g, 0)
			b = util.Max(b, 0)

			sChunk.ResourceMap[ppos] = byte(r * 255)
			sChunk.ResourceMap[ppos+1] = byte(g * 255)
			sChunk.ResourceMap[ppos+2] = byte(b * 255)
			sChunk.ResourceMap[ppos+3] = 0xFF
		}
	}
	sChunk.PixmapDirty = true
}

/* Draw a superchunk's pixmap, allocates image if needed. */
func drawPixmap(sChunk *world.MapSuperChunk, scPos world.XY) {
	/* Make Pixelmap images */
	if sChunk.PixelMap == nil {
		rect := image.Rectangle{}

		rect.Max.X = gv.SuperChunkTotal
		rect.Max.Y = gv.SuperChunkTotal

		sChunk.PixelMap = ebiten.NewImageWithOptions(rect, &ebiten.NewImageOptions{Unmanaged: true})
	}

	maxSize := gv.SuperChunkTotal * gv.SuperChunkTotal * 4
	var ObjPix []byte = make([]byte, maxSize)

	didCopy := false
	sChunk.ResourceLock.Lock()
	if world.ShowResourceLayer && sChunk.ResourceMap != nil {
		copy(ObjPix, sChunk.ResourceMap)
		didCopy = true
	}
	sChunk.ResourceLock.Unlock()

	//Fill with bg and grid
	for x := 0; x < gv.SuperChunkTotal; x++ {
		for y := 0; y < gv.SuperChunkTotal; y++ {
			ppos := 4 * (x + y*gv.SuperChunkTotal)

			if x%32 == 0 || y%32 == 0 {
				ObjPix[ppos] = 0x20
				ObjPix[ppos+1] = 0x20
				ObjPix[ppos+2] = 0x20
				ObjPix[ppos+3] = 0x10
			} else if !didCopy {
				ObjPix[ppos] = 0x05
				ObjPix[ppos+1] = 0x05
				ObjPix[ppos+2] = 0x05
				ObjPix[ppos+3] = 0xff
			}
		}
	}

	for _, chunk := range sChunk.ChunkList {
		if chunk.NumObjs <= 0 {
			continue
		}

		/* Draw objects in chunk */
		for pos := range chunk.BuildingMap {
			scX := (((scPos.X) * (gv.MaxSuperChunk)) - gv.XYCenter)
			scY := (((scPos.Y) * (gv.MaxSuperChunk)) - gv.XYCenter)

			x := int((pos.X - gv.XYCenter) - scX)
			y := int((pos.Y - gv.XYCenter) - scY)

			ppos := 4 * (x + y*gv.SuperChunkTotal)
			if ppos < maxSize {
				ObjPix[ppos] = 0xff
				ObjPix[ppos+1] = 0xff
				ObjPix[ppos+2] = 0xff
				ObjPix[ppos+3] = 0xff
			}
		}

	}
	sChunk.PixelMapLock.Lock()
	sChunk.PixelMap.WritePixels(ObjPix)
	sChunk.PixelMapTime = time.Now()
	sChunk.PixmapDirty = false
	numPixmapCache++
	sChunk.PixelMapLock.Unlock()
}
