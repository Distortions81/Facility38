package main

import (
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/remeh/sizedwaitgroup"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// World represents the game state.
type World struct {
	area   []byte
	width  int
	height int
}

// NewWorld creates a new world.
func NewWorld(width, height int, maxInitLiveCells int) *World {
	w := &World{
		area:   make([]byte, width*height),
		width:  width,
		height: height,
	}
	w.init(maxInitLiveCells)
	return w
}

// init inits world with a random state.
func (w *World) init(maxLiveCells int) {
	for i := 0; i < maxLiveCells; i++ {
		x := rand.Intn(w.width)
		y := rand.Intn(w.height)
		w.area[y*w.width+x] = 0x80
	}
}

// Update game state by one tick.
func (w *World) Update() {
	width := w.width
	height := w.height
	next := make([]byte, width*height)

	wg := sizedwaitgroup.New(runtime.NumCPU())

	for y := 0; y < height; y++ {
		wg.Add()
		go func(y int) {
			for x := 0; x < width; x++ {
				pop := neighbourCount(w.area, width, height, x, y)
				switch {
				case pop < 2:
					// rule 1. Any live cell with fewer than two live neighbours
					// dies, as if caused by under-population.
					if w.area[y*width+x] > 0x00 {
						next[y*width+x]--
					}

				case (pop == 2 || pop == 3):
					// rule 2. Any live cell with two or three live neighbours
					// lives on to the next generation.
					if w.area[y*width+x] < 0xFF {
						next[y*width+x]++
					}

				case pop > 3:
					// rule 3. Any live cell with more than three live neighbours
					// dies, as if by over-population.
					if w.area[y*width+x] > 0x00 {
						next[y*width+x]--
					}

				case pop == 3:
					// rule 4. Any dead cell with exactly three live neighbours
					// becomes a live cell, as if by reproduction.
					if w.area[y*width+x] < 0xFF {
						next[y*width+x]++
					}
				}
			}
			wg.Done()
		}(y)
	}
	wg.Wait()
	w.area = next
}

// Draw paints current game state.
func (w *World) Draw(pix []byte) {
	for i, v := range w.area {
		pix[4*i] = v
		pix[4*i+1] = v
		pix[4*i+2] = v
		pix[4*i+3] = 0xFF
	}
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// neighbourCount calculates the Moore neighborhood of (x, y).
func neighbourCount(a []byte, width, height, x, y int) int {
	c := 0

	for j := -1; j <= 1; j++ {
		for i := -1; i <= 1; i++ {
			if i == 0 && j == 0 {
				continue
			}
			x2 := x + i
			y2 := y + j
			if x2 < 0 || y2 < 0 || width <= x2 || height <= y2 {
				continue
			}
			if a[y2*width+x2] > 0x80 {
				c++
			}
		}
	}
	return c
}

//Ebiten resize handling
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

var (
	screenWidth  int = 1920
	screenHeight int = 1080
)

type Game struct {
	world  *World
	pixels []byte
}

func (g *Game) Update() error {
	g.world.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.pixels == nil {
		g.pixels = make([]byte, screenWidth*screenHeight*4)
	}
	g.world.Draw(g.pixels)
	screen.ReplacePixels(g.pixels)
}

func main() {

	g := &Game{
		world: NewWorld(screenWidth, screenHeight, int((screenWidth*screenHeight)/10)),
	}

	ebiten.SetWindowTitle("Game of Life (Ebiten Demo)")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowResizable(false)
	ebiten.SetMaxTPS(1)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOn)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
