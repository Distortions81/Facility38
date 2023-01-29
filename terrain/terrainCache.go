package terrain

import (
	"GameTest/consts"
	"GameTest/glob"
	"GameTest/noise"
	"GameTest/objects"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/remeh/sizedwaitgroup"
)

const (
	cChunkGroundCacheTime = time.Second * 15
	cCacheMax             = 100
)

var (
	gNumChunkImage int
)

func SetupTerrainCache() {
	/* Temp tile to use when rendering a new chunk */
	tChunk := glob.MapChunk{}
	renderChunkGround(&tChunk, false, glob.XY{X: 0, Y: 0})
	glob.TempChunkImage = tChunk.GroundImg
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

		chunk.GroundLock.Lock()
		tImg = ebiten.NewImage(chunkPix, chunkPix)
		gNumChunkImage++
		chunk.UsingTemporary = false
		chunk.GroundLock.Unlock()

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
	chunk.GroundLock.Lock()
	chunk.GroundImg = tImg
	chunk.GroundLock.Unlock()
}

/* Wasm single-thread version, one tile per frame */
func STCacheUpdate() {
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
			if chunk.GroundImg == nil {
				continue
			}
			if chunk.Visible || chunk.GroundImg == nil {
				if chunk.UsingTemporary {
					renderChunkGround(chunk, true, cpos)
					continue
				}
			} else {
				if gNumChunkImage > cCacheMax &&
					time.Since(chunk.LastSaw) > cChunkGroundCacheTime {
					killTerrainCache(chunk)
					continue
				}
			}
		}
	}
}

func TerrainCacheDaemon() {
	time.Sleep(time.Second)
	wg := sizedwaitgroup.New(objects.NumWorkers)

	for {
		glob.SuperChunkMapLock.Lock()
		tmpWorld := glob.SuperChunkMap
		glob.SuperChunkMapLock.Unlock()

		/* If we zoom out, decallocate everything */
		if glob.ZoomScale < consts.MapPixelThreshold {
			for _, sChunk := range tmpWorld {
				for _, chunk := range sChunk.Chunks {
					killTerrainCache(chunk)
				}
			}
			continue
		}

		for _, sChunk := range tmpWorld {
			for cpos, chunk := range sChunk.Chunks {
				if chunk.GroundImg == nil {
					continue
				}
				wg.Add()
				go func(chunk *glob.MapChunk, cpos glob.XY) {
					glob.SuperChunkMapLock.Lock()

					if chunk.Visible || chunk.GroundImg == nil {
						if chunk.UsingTemporary {
							renderChunkGround(chunk, true, cpos)
						}
					} else {
						if gNumChunkImage > cCacheMax &&
							time.Since(chunk.LastSaw) > cChunkGroundCacheTime {
							killTerrainCache(chunk)
						}
					}

					glob.SuperChunkMapLock.Unlock()
					wg.Done()
				}(chunk, cpos)
			}
		}
		wg.Wait()
		time.Sleep(time.Millisecond * 100)
	}
}

func killTerrainCache(chunk *glob.MapChunk) {
	if chunk.UsingTemporary || chunk.GroundImg == nil {
		return
	}
	chunk.GroundLock.Lock()
	chunk.GroundImg.Dispose()
	chunk.GroundImg = nil
	gNumChunkImage--
	chunk.GroundLock.Unlock()
}
