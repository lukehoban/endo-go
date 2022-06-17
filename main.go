package main

import (
	"endo/pkg/endo"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"runtime/pprof"
)

func main() {

	if true {
		f, err := os.Create("out.prof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	prefix := ""
	if len(os.Args) >= 2 {
		prefix = os.Args[1]
	}
	bitmap, err := endo.Render(prefix)
	img := image.NewRGBA(image.Rect(0, 0, 600, 600))
	for j := 0; j < 600; j++ {
		for i := 0; i < 600; i++ {
			pixel := bitmap[i][j]
			img.Set(i, j, color.RGBA{
				R: uint8(pixel.RGB.R),
				G: uint8(pixel.RGB.G),
				B: uint8(pixel.RGB.B),
				A: 255,
			})
		}
	}

	f, err := os.Create(fmt.Sprintf("images/img%s.png", prefix))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, img)

}
