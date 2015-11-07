package main

import (
	"flag"
	"image"
	"image/color"
	"image/gif"
	"log"
	"os"
)

const ()

var (
	debug = flag.Bool("debug", false, "Debug Flags On")
)

func init() {

}

func main() {
	log.Println("Generating")
	colList := color.Palette{
		color.RGBA{0x00, 0x00, 0x00, 0x00},
		color.RGBA{0xFF, 0x00, 0x00, 0xFF},
		color.RGBA{0x00, 0xFF, 0x00, 0xFF},
		color.RGBA{0x00, 0x00, 0xFF, 0xFF},
		color.RGBA{0x00, 0x00, 0x00, 0xFF},
		color.RGBA{0xFF, 0xFF, 0xFF, 0xFF},
	}

	newI := []*image.Paletted{
		image.NewPaletted(image.Rect(0, 0, 100, 100), colList),
		image.NewPaletted(image.Rect(0, 0, 100, 100), colList),
		image.NewPaletted(image.Rect(0, 0, 100, 100), colList),
		image.NewPaletted(image.Rect(0, 0, 100, 100), colList),
		image.NewPaletted(image.Rect(0, 0, 100, 100), colList),
	}

	for i := 0; i < len(newI); i++ {
		for x := 0; x < 100; x++ {
			for y := 0; y < 100; y++ {
				newI[i].SetColorIndex(x, y,
					uint8((((x+y)%len(colList))+i)%len(colList)))
			}
		}
	}

	gifData := gif.GIF{
		Image:     newI,
		Delay:     []int{10, 10, 10, 10, 10},
		LoopCount: -1,
		Config: image.Config{
			ColorModel: colList,
			Width:      100,
			Height:     100,
		},
		BackgroundIndex: 0,
	}

	f, e := os.Create("test.gif")
	if e != nil {
		log.Println(e)
	}

	e = gif.EncodeAll(f, &gifData)
	if e != nil {
		log.Println(e)
	}
}
