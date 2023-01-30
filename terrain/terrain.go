package terrain

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/noise"
	"GameTest/objects"

	"github.com/hajimehoshi/ebiten/v2"
)

func SetupTerrainCache() {
	/* Temp tile to use when rendering a new chunk */
	tChunk := glob.MapChunk{}
	renderChunkGround(&tChunk, false, glob.XY{X: 0, Y: 0})
	glob.TempChunkImage = tChunk.TerrainImg
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

					//fmt.Printf("%.2f,%.2f: %.2f\n", x, y, h)
					op.ColorM.Reset()
					op.ColorM.Scale(h*2, 1, 1, 1)
				}

				tImg.DrawImage(bg, op)
			}
		}

	} else {
		panic("No valid bg texture.")
	}
	chunk.TerrainImg = tImg
}

/* Wasm single-thread version, one tile per frame */
func RenderTerrain() {
	tmpWorld := glob.SuperChunkMap

	/* If we zoom out, decallocate everything */
	if glob.ZoomScale < consts.MapPixelThreshold {
		for _, sChunk := range tmpWorld {
			for _, chunk := range sChunk.Chunks {
				killTerrainCache(chunk)
			}
		}
		return
	}

	for _, sChunk := range tmpWorld {
		for cpos, chunk := range sChunk.Chunks {
			if chunk.TerrainImg == nil {
				continue
			}
			if chunk.Visible && glob.ZoomScale > consts.MapPixelThreshold {
				if chunk.UsingTemporary {
					renderChunkGround(chunk, true, cpos)
					continue
				}
			} else {
				killTerrainCache(chunk)
			}
		}
	}
}

func killTerrainCache(chunk *glob.MapChunk) {
	if chunk.UsingTemporary || chunk.TerrainImg == nil {
		return
	}
	chunk.TerrainImg.Dispose()
	chunk.TerrainImg = nil
}
