package main

import (
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	maxTerrainCache     = 500
	maxTerrainCacheWASM = 50
	maxPixmapCache      = 500
	maxPixmapCacheWASM  = 50

	minTerrainTime = time.Minute
	debugVisualize = false

	pixmapRenderLoop   = time.Millisecond * 100
	resourceRenderLoop = time.Millisecond * 100
)

var (
	numTerrainCache int
	numPixmapCache  int
)

/* Make a low-detail 'loading' temporary texture for chunk terrain */
func setupTerrainCache() {
	defer reportPanic("setupTerrainCache")
	tChunk := mapChunkData{}

	renderChunkGround(&tChunk, false, XY{X: 0, Y: 0})
	TempChunkImage = tChunk.terrainImage
	tChunk.usingTemporary = true

	superChunkListLock.RLock()
	for _, sChunk := range superChunkList {
		for _, chunk := range sChunk.chunkList {
			killTerrainCache(chunk, true)
		}
	}
	superChunkListLock.RUnlock()

	if debugVisualize {
		TempChunkImage.Fill(ColorDarkRed)
	}
}

/* Render a chunk's terrain to chunk.TerrainImg, locks chunk.TerrainLock */
func renderChunkGround(chunk *mapChunkData, doDetail bool, chunkPos XY) {
	defer reportPanic("renderChunkGround")
	chunkPix := (spriteScale * chunkSize)

	var bg *ebiten.Image = terrainTypes[0].images.main
	sx := int(float32(bg.Bounds().Size().X))
	sy := int(float32(bg.Bounds().Size().Y))
	var tImg *ebiten.Image

	if sx > 0 && sy > 0 {

		rect := image.Rectangle{}

		rect.Max.X = chunkPix
		rect.Max.Y = chunkPix

		if chunk.usingTemporary || chunk.terrainImage == nil {
			tImg = ebiten.NewImageWithOptions(rect, &ebiten.NewImageOptions{Unmanaged: true})
		}

		var opList [chunkSize * chunkSize]*ebiten.DrawImageOptions
		var opPos uint16

		for i := 0; i < chunkSize; i++ {
			for j := 0; j < chunkSize; j++ {
				var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(i*sx), float64(j*sy))

				if doDetail {
					x := (float32(chunkPos.X*chunkSize) + float32(i))
					y := (float32(chunkPos.Y*chunkSize) + float32(j))

					h := noiseMap(x, y, 0)

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

		chunk.terrainLock.Lock()
		/* Batch render */
		for _, op := range opList {
			tImg.DrawImage(bg, op)
		}
		numTerrainCache++
		chunk.terrainImage = tImg
		chunk.usingTemporary = false
		chunk.terrainTime = time.Now()
		chunk.terrainLock.Unlock()

	} else {
		panic("No valid bg texture.")
	}

}

var clearedCache bool

/* WASM single-thread version, one tile per call */
/* Disposes everything if we switch layers */
func renderTerrainST() {
	defer reportPanic("RenderTerrainST")

	/* If we zoom out, deallocate everything */
	if wasmMode && zoomScale <= mapPixelThreshold {
		if wasmMode && !clearedCache {
			for _, sChunk := range superChunkList {
				for _, chunk := range sChunk.chunkList {
					killTerrainCache(chunk, true)
				}
			}
			clearedCache = true
		}
	} else {
		clearedCache = false

		superChunkListLock.RLock()
		for _, chunk := range visChunk {
			if chunk.usingTemporary {
				renderChunkGround(chunk, true, chunk.pos)
			}
		}
		superChunkListLock.RUnlock()

		/* Kill non-visible */
		for _, sChunk := range superChunkList {
			for _, chunk := range sChunk.chunkList {
				if !chunk.visible && !chunk.usingTemporary {
					killTerrainCache(chunk, false)
				}
			}
		}
	}
}

/* Dispose terrain cache in a chunk if needed. Always dispose: force. Locks chunk.TerrainLock */
func killTerrainCache(chunk *mapChunkData, force bool) {
	defer reportPanic("killTerrainCache")
	if chunk.usingTemporary || chunk.terrainImage == nil {
		return
	}

	if force ||
		(numTerrainCache > maxTerrainCache &&
			time.Since(chunk.terrainTime) > minTerrainTime) ||
		(wasmMode && numTerrainCache > maxTerrainCacheWASM) {

		chunk.terrainLock.Lock()
		chunk.terrainImage.Dispose()
		chunk.terrainImage = nil
		chunk.terrainImage = TempChunkImage
		chunk.usingTemporary = true
		numTerrainCache--
		chunk.terrainLock.Unlock()

	}
}

/* Render pixmap images, one tile per call. Disposes everything on layer change */
var pixmapCacheCleared bool

func pixmapRenderST() {
	defer reportPanic("PixmapRenderST")
	if !showResourceLayer && zoomScale > mapPixelThreshold && !pixmapCacheCleared {

		for _, sChunk := range superChunkList {
			if sChunk.pixelMap != nil {

				sChunk.pixelMap.Dispose()
				sChunk.pixelMap = nil
				numPixmapCache--
				break

			}
		}
		pixmapCacheCleared = true

	} else if zoomScale <= mapPixelThreshold || showResourceLayer {
		pixmapCacheCleared = false

		for _, sChunk := range superChunkList {
			if sChunk.pixelMap == nil || sChunk.pixmapDirty {
				drawPixmap(sChunk, sChunk.pos)
				break
			}

		}
	}
}

/* Loop, renders and disposes superChunk to sChunk.PixMap Locks sChunk.PixLock */
func pixmapRenderDaemon() {
	defer reportPanic("PixmapRenderDaemon")

	for gameRunning {
		superChunkListLock.RLock()
		for _, sChunk := range superChunkList {
			if sChunk.numChunks == 0 {
				continue
			}

			if !showResourceLayer && zoomScale > mapPixelThreshold {
				sChunk.pixelMapLock.Lock()
				if sChunk.pixelMap != nil &&
					(maxPixmapCache > numPixmapCache) {

					sChunk.pixelMap.Dispose()
					doLog(true, "dispose pixmap %v", sChunk.pos)
					sChunk.pixelMap = nil
					numPixmapCache--

				}
				sChunk.pixelMapLock.Unlock()
			} else if zoomScale <= mapPixelThreshold || showResourceLayer {

				if sChunk.pixelMap == nil || sChunk.pixmapDirty {
					drawPixmap(sChunk, sChunk.pos)
					doLog(true, "render pixmap %v", sChunk.pos)
				}
			}
		}
		superChunkListLock.RUnlock()
		time.Sleep(pixmapRenderLoop)
	}
}

/* Loop, renders and disposes superChunk to sChunk.PixMap Locks sChunk.PixLock */
func resourceRenderDaemon() {
	defer reportPanic("resourceRenderDaemon")
	for gameRunning {

		superChunkListLock.RLock()
		for _, sChunk := range superChunkList {
			sChunk.resourceLock.Lock()
			if sChunk.resourceMap == nil || sChunk.resourceDirty {
				drawResource(sChunk)
				sChunk.resourceDirty = false
			}
			sChunk.resourceLock.Unlock()
		}
		superChunkListLock.RUnlock()

		time.Sleep(resourceRenderLoop)
	}
}

/* Render resources during render for WASM single-thread */
func resourceRenderDaemonST() {
	defer reportPanic("resourceRenderDaemonST")
	for _, sChunk := range superChunkList {
		if sChunk.resourceMap == nil || sChunk.resourceDirty {
			drawResource(sChunk)
			sChunk.resourceDirty = false
			break
		}
	}
}

/* Draw perlin noise resource channel */
func drawResource(sChunk *mapSuperChunkData) {
	defer reportPanic("drawResource")
	if sChunk == nil {
		return
	}

	if sChunk.resourceMap == nil {
		sChunk.resourceMap = make([]byte, superChunkTotal*superChunkTotal*4)
	}
	sChunkPos := SuperChunkPosToPos(sChunk.pos)

	var chunk *mapChunkData
	for x := 0; x < superChunkTotal; x++ {
		for y := 0; y < superChunkTotal; y++ {
			pixelPos := 4 * (x + y*superChunkTotal)

			worldX := float32(sChunkPos.X + uint16(x))
			worldY := float32(sChunkPos.Y + uint16(y))
			if y%chunkSize == 0 {
				chunk = sChunk.chunkMap[PosToChunkPos(XY{X: uint16(worldX), Y: uint16(worldY)})]
			}

			var r, g, b float32 = 0.01, 0.01, 0.01
			for p, nl := range noiseLayers {
				if p == 0 {
					continue
				}

				h := noiseMap(worldX, worldY, p)

				if chunk != nil {
					Tile := chunk.tileMap[XY{X: uint16(x), Y: uint16(y)}]

					if Tile != nil {
						h -= (Tile.minerData.mined[p] / 150)
					}
				}
				if nl.modRed {
					r += (h * nl.redMulti)
				}
				if nl.modGreen {
					g += (h * nl.greenMulti)
				}
				if nl.modBlue {
					b += (h * nl.blueMulti)
				}
			}
			r = Min(r, 1.0)
			g = Min(g, 1.0)
			b = Min(b, 1.0)

			r = Max(r, 0)
			g = Max(g, 0)
			b = Max(b, 0)

			sChunk.resourceMap[pixelPos] = byte(r * 255)
			sChunk.resourceMap[pixelPos+1] = byte(g * 255)
			sChunk.resourceMap[pixelPos+2] = byte(b * 255)
			sChunk.resourceMap[pixelPos+3] = 0xFF
		}
	}
	sChunk.pixmapDirty = true
}

/* Draw a superChunk's pixmap, allocates image if needed. */
func drawPixmap(sChunk *mapSuperChunkData, scPos XY) {
	defer reportPanic("drawPixmap")
	maxSize := superChunkTotal * superChunkTotal * 4
	if sChunk.itemMap == nil {
		sChunk.itemMap = make([]byte, maxSize)
	}

	didCopy := false
	sChunk.resourceLock.Lock()
	if showResourceLayer && sChunk.resourceMap != nil {
		copy(sChunk.itemMap, sChunk.resourceMap)
		didCopy = true
	}
	sChunk.resourceLock.Unlock()

	//Fill with bg and grid
	for x := 0; x < superChunkTotal; x++ {
		for y := 0; y < superChunkTotal; y++ {
			pixelPos := 4 * (x + y*superChunkTotal)

			if x%32 == 0 || y%32 == 0 {
				sChunk.itemMap[pixelPos] = 0x20
				sChunk.itemMap[pixelPos+1] = 0x20
				sChunk.itemMap[pixelPos+2] = 0x20
				sChunk.itemMap[pixelPos+3] = 0x10
			} else if !didCopy {
				sChunk.itemMap[pixelPos] = 0x05
				sChunk.itemMap[pixelPos+1] = 0x05
				sChunk.itemMap[pixelPos+2] = 0x05
				sChunk.itemMap[pixelPos+3] = 0xff
			}
		}
	}

	for _, chunk := range sChunk.chunkList {

		if chunk.numObjs <= 0 {
			continue
		}

		/* Draw objects in chunk */
		for pos := range chunk.buildingMap {
			scX := (((scPos.X) * (maxSuperChunk)) - xyCenter)
			scY := (((scPos.Y) * (maxSuperChunk)) - xyCenter)

			x := int((pos.X - xyCenter) - scX)
			y := int((pos.Y - xyCenter) - scY)

			pixelPos := 4 * (x + y*superChunkTotal)
			if pixelPos < maxSize {
				sChunk.itemMap[pixelPos] = 0xff
				sChunk.itemMap[pixelPos+1] = 0xff
				sChunk.itemMap[pixelPos+2] = 0xff
				sChunk.itemMap[pixelPos+3] = 0xff
			}
		}

	}

	sChunk.pixelMapLock.Lock()
	/* Make Pixelmap images */
	if sChunk.pixelMap == nil {
		rect := image.Rectangle{}

		rect.Max.X = superChunkTotal
		rect.Max.Y = superChunkTotal

		sChunk.pixelMap = ebiten.NewImageWithOptions(rect, &ebiten.NewImageOptions{Unmanaged: true})
		numPixmapCache++
	}
	sChunk.pixelMap.WritePixels(sChunk.itemMap)
	sChunk.pixelMapTime = time.Now()
	sChunk.pixmapDirty = false
	sChunk.pixelMapLock.Unlock()
}
