package terrain

import (
	"GameTest/consts"
	"GameTest/cwlog"
	"GameTest/glob"
	"GameTest/noise"
	"GameTest/objects"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	maxTerrainCache = 50
	renderRest      = time.Millisecond
	renderLoop      = time.Millisecond * 10
	zoomCachePurge  = false //Always purges on WASM
	debugVisualize  = false
)

var (
	numTerrainCache int
)

func SetupTerrainCache() {
	/* Temp tile to use when rendering a new chunk */
	tChunk := glob.MapChunk{}
	renderChunkGround(&tChunk, false, glob.XY{X: 0, Y: 0})
	glob.TempChunkImage = tChunk.TerrainImg

	if debugVisualize {
		glob.TempChunkImage.Fill(glob.ColorDarkRed)
	}
}

func renderChunkGround(chunk *glob.MapChunk, doDetail bool, cpos glob.XY) {
	/* Make optimized background */
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

	chunkPix := (consts.SpriteScale * consts.ChunkSize)

	bg := objects.TerrainTypes[0].Image
	sx := int(float64(bg.Bounds().Size().X))
	sy := int(float64(bg.Bounds().Size().Y))
	var tImg *ebiten.Image

	if sx > 0 && sy > 0 {

		tImg = ebiten.NewImage(chunkPix, chunkPix)
		chunk.UsingTemporary = false

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
	chunk.TerrainLock.Unlock()
}

/* Wasm single-thread version, one tile per frame */
func RenderTerrainST() {

	/* If we zoom out, decallocate everything */
	if glob.ZoomScale <= consts.MapPixelThreshold {
		if !clearedCache {
			for i := 0; i < glob.VisChunkTop; i++ {
				chunk := glob.VisChunks[i]
				killTerrainCache(chunk, true)
			}
			clearedCache = true
		}
	} else {
		clearedCache = false
		for i := 0; i < glob.VisChunkTop; i++ {

			cpos := glob.VisChunkPos[i]
			chunk := glob.VisChunks[i]

			if chunk.TerrainImg == nil {
				continue
			}
			if chunk.Precache && chunk.UsingTemporary {
				renderChunkGround(chunk, true, cpos)
				break
			} else if !chunk.Precache {
				killTerrainCache(chunk, false)
			}
		}
	}

}

var clearedCache bool

func RenderTerrainDaemon() {
	for {
		time.Sleep(renderLoop)

		/* If we zoom out, decallocate everything */
		if glob.ZoomScale <= consts.MapPixelThreshold {
			if !clearedCache && zoomCachePurge {
				for i := 0; i < glob.VisChunkTop; i++ {
					time.Sleep(renderRest)
					chunk := glob.VisChunks[i]
					killTerrainCache(chunk, true)
				}
				clearedCache = true
			}
		} else {
			clearedCache = false
			for i := 0; i < glob.VisChunkTop; i++ {

				cpos := glob.VisChunkPos[i]
				chunk := glob.VisChunks[i]

				if chunk.TerrainImg == nil {
					continue
				}
				time.Sleep(renderRest)
				if chunk.Precache && chunk.UsingTemporary {
					renderChunkGround(chunk, true, cpos)
				} else if !chunk.Precache {
					killTerrainCache(chunk, false)
				}
			}
		}
	}
}

func killTerrainCache(chunk *glob.MapChunk, force bool) {
	if chunk.UsingTemporary || chunk.TerrainImg == nil {
		return
	}
	if force || numTerrainCache > maxTerrainCache {
		chunk.TerrainLock.Lock()
		chunk.TerrainImg.Dispose()
		chunk.TerrainImg = nil
		numTerrainCache--
		chunk.TerrainLock.Unlock()
	}
}
