package objects

import (
	"GameTest/gv"
	"GameTest/world"
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	maxTerrainCache     = 500
	maxTerrainCacheWASM = 50
	minTerrainTime      = time.Minute
	terrainRenderLoop   = time.Nanosecond
	debugVisualize      = false

	maxPixmapCache     = 500
	maxPixmapCacheWASM = 50
	minPixmapTime      = time.Minute
	pixmapRenderLoop   = time.Nanosecond
)

var (
	numTerrainCache int
	numPixmapCache  int
)

func SwitchLayer() {
	world.ShowMineralLayerLock.Lock()

	if world.ShowMineralLayer {
		world.ShowMineralLayer = false
	} else {
		world.ShowMineralLayer = true
	}
	world.ShowMineralLayerLock.Unlock()

	SetupTerrainCache()
}

/* Make a 'loading' temporary texture for chunk terrain */
func SetupTerrainCache() {

	tChunk := world.MapChunk{}
	renderChunkGround(&tChunk, false, world.XY{X: 0, Y: 0})
	world.TempChunkImage = tChunk.TerrainImg

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
	if chunk.Rendering {
		return
	}
	world.ShowMineralLayerLock.RLock()
	defer world.ShowMineralLayerLock.RUnlock()

	chunk.Rendering = true

	/* Make optimized background */
	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

	chunkPix := (gv.SpriteScale * gv.ChunkSize)

	var bg *ebiten.Image
	if !world.ShowMineralLayer {
		bg = TerrainTypes[0].Image
	} else {
		bg = TerrainTypes[1].Image
	}
	sx := int(float32(bg.Bounds().Size().X))
	sy := int(float32(bg.Bounds().Size().Y))
	var tImg *ebiten.Image

	if sx > 0 && sy > 0 {

		rect := image.Rectangle{}

		if world.ShowMineralLayer {
			rect.Max.X = gv.ChunkSize
			rect.Max.Y = gv.ChunkSize
		} else {
			rect.Max.X = chunkPix
			rect.Max.Y = chunkPix
		}

		tImg = ebiten.NewImageWithOptions(rect, &ebiten.NewImageOptions{Unmanaged: true})

		for i := 0; i < gv.ChunkSize; i++ {
			for j := 0; j < gv.ChunkSize; j++ {
				op.GeoM.Reset()
				op.GeoM.Translate(float64(i*sx), float64(j*sy))

				if doDetail && !world.ShowMineralLayer {
					x := (float32(cpos.X*gv.ChunkSize) + float32(i))
					y := (float32(cpos.Y*gv.ChunkSize) + float32(j))

					h := NoiseMap(x, y, 0)

					op.ColorScale.Reset()
					op.ColorScale.Scale(h, 1, 1, 1)

				} else if doDetail {
					op.ColorScale.Reset()
					x := (float32(cpos.X*gv.ChunkSize) + float32(i))
					y := (float32(cpos.Y*gv.ChunkSize) + float32(j))

					for p, nl := range NoiseLayers {
						if p == 0 {
							continue
						}
						var r, g, b float32 = 0.98, 0.98, 0.98
						h := NoiseMap(x, y, p)
						if nl.InvertValue {
							h = -h
						}

						if nl.RMod {
							r -= (h * nl.RMulti)
						}
						if nl.GMod {
							g -= (h * nl.GMulti)
						}
						if nl.BMod {
							b -= (h * nl.BMulti)
						}
						op.ColorScale.Scale(r, g, b, 1)
					}

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
	if world.ZoomScale <= gv.MapPixelThreshold && !world.ShowMineralLayer {
		if !clearedCache && gv.WASMMode {
			for _, sChunk := range world.SuperChunkList {
				for _, chunk := range sChunk.ChunkList {
					killTerrainCache(chunk, true)
				}
			}
			clearedCache = true
		}
	} else {
		clearedCache = false

		for _, sChunk := range world.SuperChunkList {
			for _, chunk := range sChunk.ChunkList {
				if chunk.Precache && chunk.UsingTemporary {
					renderChunkGround(chunk, true, chunk.Pos)
					break
				} else if !chunk.Precache {
					killTerrainCache(chunk, false)
				}
			}
		}
	}
}

/* Loop to automatically render chunk terrain, will also dispose old tiles, useses SuperChunkList*/
func RenderTerrainDaemon() {
	for {
		time.Sleep(terrainRenderLoop)

		/* If we zoom out, decallocate everything */
		if world.ZoomScale <= gv.MapPixelThreshold && !world.ShowMineralLayer {
			if !clearedCache && gv.WASMMode {
				world.SuperChunkListLock.RLock()
				for _, sChunk := range world.SuperChunkList {
					for _, chunk := range sChunk.ChunkList {
						killTerrainCache(chunk, true)
					}
				}
				world.SuperChunkListLock.RUnlock()
				clearedCache = true
			}
		} else {
			clearedCache = false

			world.SuperChunkListLock.RLock()
			for _, sChunk := range world.SuperChunkList {
				if !sChunk.Visible {
					continue
				}
				for _, chunk := range sChunk.ChunkList {
					if chunk.Precache && chunk.UsingTemporary {
						renderChunkGround(chunk, true, chunk.Pos)
					} else if !chunk.Precache {
						killTerrainCache(chunk, false)
					}
				}
			}
			world.SuperChunkListLock.RUnlock()
		}
	}
}

/* Dispose terrain cache in a chunk if needed. Always dispose: force. Locks chunk.TerrainLock */
func killTerrainCache(chunk *world.MapChunk, force bool) {

	if chunk.UsingTemporary || chunk.Rendering || chunk.TerrainImg == nil {
		return
	}

	if force ||
		(numTerrainCache > maxTerrainCache &&
			time.Since(chunk.TerrainTime) > minTerrainTime) ||
		(gv.WASMMode && numTerrainCache > maxTerrainCacheWASM) {

		chunk.TerrainLock.Lock()
		chunk.TerrainImg.Dispose()
		chunk.TerrainImg = world.TempChunkImage
		chunk.UsingTemporary = true
		numTerrainCache--
		chunk.TerrainLock.Unlock()

	}
}

/* Render pixmap images, one tile per call. Also disposes if zoom level changes. */
func PixmapRenderST() {

	if world.ZoomScale > gv.MapPixelThreshold && !world.ShowMineralLayer {

		if !pixmapCacheCleared {
			for _, sChunk := range world.SuperChunkList {
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

		for _, sChunk := range world.SuperChunkList {
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

		world.SuperChunkListLock.RLock()
		for _, sChunk := range world.SuperChunkList {

			if world.ZoomScale > gv.MapPixelThreshold && !pixmapCacheCleared {

				pixmapCacheCleared = true
				sChunk.PixLock.Lock()
				if sChunk.PixMap != nil &&
					(maxPixmapCache > numPixmapCache ||
						(gv.WASMMode && maxPixmapCacheWASM > numPixmapCache)) {

					sChunk.PixMap.Dispose()
					sChunk.PixMap = nil
					numPixmapCache--

				}
				sChunk.PixLock.Unlock()
			} else if world.ZoomScale <= gv.MapPixelThreshold && !world.ShowMineralLayer {
				pixmapCacheCleared = false

				sChunk.PixLock.Lock()
				if sChunk.PixMap == nil || sChunk.PixmapDirty {
					drawPixmap(sChunk, sChunk.Pos)
				}
				sChunk.PixLock.Unlock()
			}
		}
		world.SuperChunkListLock.RUnlock()
	}
}

/* Draw a superchunk's pixmap, allocates image if needed. */
func drawPixmap(sChunk *world.MapSuperChunk, scPos world.XY) {
	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

	/* Make Pixelmap images */
	if sChunk.PixMap == nil {
		rect := image.Rectangle{}

		rect.Max.X = gv.MaxSuperChunk
		rect.Max.Y = gv.MaxSuperChunk

		sChunk.PixMap = ebiten.NewImageWithOptions(rect, &ebiten.NewImageOptions{Unmanaged: true})
	}

	sChunk.PixMap.Fill(world.ColorCharcol)

	for _, chunk := range sChunk.ChunkList {
		if chunk.NumObjects <= 0 {
			continue
		}

		/* Draw objects in chunk */
		for _, obj := range chunk.ObjList {
			scX := (((scPos.X) * (gv.MaxSuperChunk)) - gv.XYCenter)
			scY := (((scPos.Y) * (gv.MaxSuperChunk)) - gv.XYCenter)

			x := float64((obj.Pos.X - gv.XYCenter) - scX)
			y := float64((obj.Pos.Y - gv.XYCenter) - scY)
			op.GeoM.Reset()
			op.GeoM.Translate(x, y)
			sChunk.PixMap.DrawImage(world.MiniMapTile, op)
		}
		sChunk.PixMapTime = time.Now()
		sChunk.PixmapDirty = false
		numPixmapCache++
	}
}
