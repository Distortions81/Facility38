package terrain

import (
	"GameTest/consts"
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/noise"
	"GameTest/objects"
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	maxTerrainCache   = 500
	minTerrainTime    = time.Minute
	terrainRenderRest = time.Millisecond * 10
	terrainRenderLoop = time.Millisecond * 100
	zoomCachePurge    = false //Always purges on WASM
	debugVisualize    = false

	maxPixmapCache   = 500
	minPixmapTime    = time.Minute
	pixmapRenderRest = time.Millisecond * 10
	pixmapRenderLoop = time.Millisecond * 100
)

var (
	numTerrainCache int
	numPixmapCache  int
)

/* Make a 'loading' temporary texture for chunk terrain */
func SetupTerrainCache() {

	tChunk := glob.MapChunk{}
	renderChunkGround(&tChunk, false, glob.XY{X: 0, Y: 0})
	glob.TempChunkImage = tChunk.TerrainImg

	if debugVisualize {
		glob.TempChunkImage.Fill(glob.ColorDarkRed)
	}
}

/* Render a chunk's terrain to chunk.TerrainImg, locks chunk.TerrainLock */
func renderChunkGround(chunk *glob.MapChunk, doDetail bool, cpos glob.XY) {

	if chunk.Rendering {
		return
	}
	chunk.Rendering = true

	/* Make optimized background */
	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

	chunkPix := (consts.SpriteScale * consts.ChunkSize)

	bg := objects.TerrainTypes[0].Image
	sx := int(float64(bg.Bounds().Size().X))
	sy := int(float64(bg.Bounds().Size().Y))
	var tImg *ebiten.Image

	if sx > 0 && sy > 0 {

		rect := image.Rectangle{}

		rect.Max.X = chunkPix
		rect.Max.Y = chunkPix

		tImg = ebiten.NewImageWithOptions(rect, &ebiten.NewImageOptions{Unmanaged: true})

		for i := 0; i < consts.ChunkSize; i++ {
			for j := 0; j < consts.ChunkSize; j++ {
				op.GeoM.Reset()
				op.GeoM.Translate(float64(i*sx), float64(j*sy))

				if doDetail {
					x := (float64(cpos.X*consts.ChunkSize) + float64(i))
					y := (float64(cpos.Y*consts.ChunkSize) + float64(j))
					h := noise.NoiseMap(x, y)

					if consts.Verbose {
						cwlog.DoLog("%.2f,%.2f: %.2f", x, y, h)
					}
					op.ColorM.Reset()
					op.ColorM.Scale(h*2, 1, 1, 1)
				}

				tImg.DrawImage(bg, op)
			}
		}

	} else {
		panic("No valid bg texture.")
	}
	chunk.TerrainLock.Lock()
	numTerrainCache++
	chunk.TerrainImg = tImg
	chunk.UsingTemporary = false
	chunk.Rendering = false
	chunk.TerrainTime = time.Now()
	chunk.TerrainLock.Unlock()
}

var clearedCache bool

/* WASM single-thread version, one tile per call */
func RenderTerrainST() {

	/* If we zoom out, decallocate everything */
	if glob.ZoomScale <= consts.MapPixelThreshold {
		if !clearedCache && (zoomCachePurge || glob.WASMMode) {
			for _, sChunk := range glob.SuperChunkList {
				for _, chunk := range sChunk.ChunkList {
					killTerrainCache(chunk, true)
				}
			}
			clearedCache = true
		}
	} else {
		clearedCache = false

		glob.SuperChunkListLock.RLock()
		for _, sChunk := range glob.SuperChunkList {
			if !sChunk.Visible {
				continue
			}
			for _, chunk := range sChunk.ChunkList {
				if chunk.Precache && chunk.UsingTemporary {
					renderChunkGround(chunk, true, chunk.Pos)
					break
				} else if !chunk.Precache &&
					numTerrainCache > maxTerrainCache {
					killTerrainCache(chunk, false)
				}
			}
		}
	}
}

/* Loop to automatically render chunk terrain, will also dispose old tiles, uses glob.VisChunks */
func RenderTerrainDaemon() {
	for {
		time.Sleep(terrainRenderLoop)

		/* If we zoom out, decallocate everything */
		if glob.ZoomScale <= consts.MapPixelThreshold {
			if !clearedCache && zoomCachePurge {
				glob.SuperChunkListLock.RLock()
				for _, sChunk := range glob.SuperChunkList {
					for _, chunk := range sChunk.ChunkList {
						killTerrainCache(chunk, true)
						time.Sleep(terrainRenderRest)
					}
				}
				glob.SuperChunkListLock.RUnlock()
				clearedCache = true
			}
		} else {
			clearedCache = false

			glob.SuperChunkListLock.RLock()
			for _, sChunk := range glob.SuperChunkList {
				if !sChunk.Visible {
					continue
				}
				for _, chunk := range sChunk.ChunkList {
					if chunk.Precache && chunk.UsingTemporary {
						renderChunkGround(chunk, true, chunk.Pos)
						time.Sleep(terrainRenderRest)
					} else if !chunk.Precache &&
						numTerrainCache > maxTerrainCache {
						killTerrainCache(chunk, false)
						time.Sleep(terrainRenderRest)
					}
				}
			}
			glob.SuperChunkListLock.RUnlock()
		}
	}
}

/* Dispose terrain cache in a chunk if needed. Always dispose: force. Locks chunk.TerrainLock */
func killTerrainCache(chunk *glob.MapChunk, force bool) {

	if chunk.UsingTemporary || chunk.TerrainImg == nil {
		return
	}
	if force || numTerrainCache > maxTerrainCache &&
		time.Since(chunk.TerrainTime) > minTerrainTime {
		chunk.TerrainLock.Lock()
		chunk.TerrainImg.Dispose()
		chunk.TerrainImg = glob.TempChunkImage
		chunk.UsingTemporary = true
		numTerrainCache--
		chunk.TerrainLock.Unlock()
	}
}

/* Render pixmap images, one tile per call. Also disposes if zoom level changes. */
func PixmapRenderST() {

	if glob.ZoomScale > consts.MapPixelThreshold {

		if !pixmapCacheCleared {
			for _, sChunk := range glob.SuperChunkList {
				if sChunk.PixMap != nil {

					sChunk.PixMap.Dispose()
					sChunk.PixMap = nil
					numPixmapCache--
					break

				}
			}
			pixmapCacheCleared = true
		}
	} else {
		pixmapCacheCleared = false

		for _, sChunk := range glob.SuperChunkList {
			if sChunk.PixMap == nil || sChunk.PixmapDirty {
				drawPixmap(sChunk, sChunk.Pos)
				break
			}

		}
	}
}

var pixmapCacheCleared bool

/* Loop, renders and disposes superchunk to sChunk.PixMap Locks sChunk.PixLock */
func PixmapRenderDaemon() {

	for {
		time.Sleep(pixmapRenderLoop)

		glob.SuperChunkListLock.RLock()
		for _, sChunk := range glob.SuperChunkList {

			if glob.ZoomScale > consts.MapPixelThreshold && !pixmapCacheCleared {

				pixmapCacheCleared = true
				sChunk.PixLock.Lock()
				if sChunk.PixMap != nil && maxPixmapCache > numPixmapCache {

					sChunk.PixMap.Dispose()
					sChunk.PixMap = nil
					numPixmapCache--
					time.Sleep(pixmapRenderRest)

				}
				sChunk.PixLock.Unlock()
			} else if glob.ZoomScale <= consts.MapPixelThreshold {
				pixmapCacheCleared = false

				sChunk.PixLock.Lock()
				if sChunk.PixMap == nil || sChunk.PixmapDirty {
					drawPixmap(sChunk, sChunk.Pos)
					time.Sleep(pixmapRenderRest)
				}
				sChunk.PixLock.Unlock()
			}
		}
		glob.SuperChunkListLock.RUnlock()
	}
}

/* Draw a superchunk's pixmap, allocates image if needed. */

func drawPixmap(sChunk *glob.MapSuperChunk, scPos glob.XY) {
	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

	/* Make Pixelmap images */
	if sChunk.PixMap == nil {
		rect := image.Rectangle{}

		rect.Max.X = consts.MaxSuperChunk
		rect.Max.Y = consts.MaxSuperChunk

		sChunk.PixMap = ebiten.NewImageWithOptions(rect, &ebiten.NewImageOptions{Unmanaged: true})
	}

	sChunk.PixMap.Fill(glob.ColorCharcol)

	for _, chunk := range sChunk.ChunkList {
		if chunk.NumObjects <= 0 {
			continue
		}

		/* Draw objects in chunk */
		for _, obj := range chunk.ObjList {
			scX := (((scPos.X) * (consts.MaxSuperChunk)) - consts.XYCenter)
			scY := (((scPos.Y) * (consts.MaxSuperChunk)) - consts.XYCenter)

			x := float64((obj.Pos.X - consts.XYCenter) - scX)
			y := float64((obj.Pos.Y - consts.XYCenter) - scY)
			op.GeoM.Reset()
			op.GeoM.Translate(x, y)
			sChunk.PixMap.DrawImage(glob.MiniMapTile, op)
		}
		sChunk.PixMapTime = time.Now()
		sChunk.PixmapDirty = false
		numPixmapCache++
	}
}
