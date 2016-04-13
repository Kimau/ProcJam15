package main

import (
	"log"
	"math/rand"
	"os"
	"strings"
	"unicode"

	"image"
	"image/color"
	"image/gif"

	"github.com/golang/freetype/raster"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	
	"github.com/ChimeraCoder/anaconda"
)

const (
	NOOF_COLOURS    = 6
)

type FireWork struct {
	x, y, dx, dy, t, sz, d float64
	colID                  uint8
	bundle                 []FireWork
	tweet                  *anaconda.Tweet
	text                   string
}


func genFireworkFromWord(tword string, srcFire *FireWork) FireWork {
	h := hash(tword)
	rGen := rand.New(rand.NewSource(h))

	f := FireWork{
		x:     rGen.Float64()*10.0 - 5.0,
		y:     rGen.Float64()*10.0 - 5.0,
		t:     rGen.Float64()*10.0 + 5.0,
		d:     0,
		dx:    (rGen.Float64() - 0.5) * 10.0,
		dy:    (rGen.Float64() - 0.5) * 10.0,
		sz:    float64(len(tword)) * 0.3,
		colID: srcFire.colID,
		text:  tword[0:1],
	}

	return f
}


func genFireworkFromTweet(tweets []anaconda.Tweet, i int, w float64, h float64) FireWork {

	twt := &tweets[i]
	TSstart, _ := tweets[0].CreatedAtTime()
	TSend, _ := tweets[len(tweets)-1].CreatedAtTime()

	startUnix := TSstart.Unix()
	deltaUnix := float64(TSend.Unix() - startUnix)

	wordSplitter := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && (c != '`') && (c != '\'')
	}
	words := strings.FieldsFunc(twt.Text, wordSplitter)

	timestamp, _ := twt.CreatedAtTime()
	relTime := float64(timestamp.Unix()-startUnix) / deltaUnix
	relLen := float64(len(twt.Text)) / 140.0

	f := FireWork{
		x:     float64(timestamp.Minute()) * w / 60.0,
		y:     h,
		t:     40.0 * relLen,
		d:     10.0 * relTime,
		dx:    0.0,
		dy:    -3.0,
		sz:    6.0,
		colID: uint8((twt.Id % NOOF_COLOURS) + 1),
		tweet: twt,
		text:  twt.Text[0:1],
	}

	tx := w*0.25 + w*0.5*float64(timestamp.Second())/60.0
	ty := h*0.3 - h*relLen*0.35

	f.dx = (tx - f.x) / f.t
	f.dy = (ty - f.y) / f.t

	for _, w := range words {
		f.bundle = append(f.bundle, genFireworkFromWord(w, &f))
	}

	return f
}

func genTwitterGif(tweets []anaconda.Tweet, username string, tid int64) (string, error) {
	wid := 440
	height := 220

	colList := color.Palette{
		color.RGBA{0x00, 0x00, 0x00, 0xFF},

		color.RGBA{0xFF, 0x33, 0x33, 0xFF},
		color.RGBA{0x33, 0xFF, 0x33, 0xFF},
		color.RGBA{0x33, 0x33, 0xFF, 0xFF},
		color.RGBA{0xFF, 0xFF, 0x33, 0xFF},
		color.RGBA{0xFF, 0x33, 0xFF, 0xFF},
		color.RGBA{0x33, 0xFF, 0xFF, 0xFF},
	}

	newList := []*image.Paletted{}
	delayList := []int{}
	fireworkList := []FireWork{}
	disposalList := []byte{}

	draw2d.SetFontFolder("static")

	for i := range tweets {
		f := genFireworkFromTweet(tweets, i, float64(wid), float64(height))
		fireworkList = append(fireworkList, f)
	}

	boundRect := image.Rect(0, 0, wid, height)

	for len(fireworkList) > 0 {
		rawImg := image.NewRGBA(boundRect)

		// TODO :: Create Custom Painter
		// which does blend up
		painter := raster.NewRGBAPainter(rawImg)
		gc := draw2dimg.NewGraphicContextWithPainter(rawImg, painter)

		gc.SetFontData(draw2d.FontData{
			Name: "Roboto",
		})

		gc.SetFontSize(8)

		gc.Clear()
		gc.SetFillColor(colList[0])
		gc.MoveTo(0, 0)
		gc.LineTo(0, float64(height))
		gc.LineTo(float64(wid), float64(height))
		gc.LineTo(float64(wid), 0)
		gc.Close()
		gc.Fill()

		newFList := []FireWork{}

		for _, f := range fireworkList {

			if f.d > 0 {
				f.d -= 1.0
			} else {

				gc.SetFillColor(colList[f.colID])
				gc.SetStrokeColor(colList[f.colID])

				gc.MoveTo(f.x, f.y)
				gc.FillStringAt(f.text, f.x-4, f.y+4)

				gc.MoveTo(f.x, f.y)
				gc.SetLineWidth(f.sz)
				gc.LineTo(f.x-f.dx, f.y-f.dy)
				for ns := 1.0; ns < f.sz; ns += 1.0 {
					gc.SetLineWidth(f.sz - ns)
					gc.LineTo(f.x-f.dx*ns*0.2, f.y-f.dy*ns*0.2)
				}
				gc.Stroke()

				f.x += f.dx
				f.y += f.dy
				f.t -= 1.0

				f.dy += 0.3
			}

			if f.t > 0 {
				newFList = append(newFList, f)
			} else if len(f.bundle) > 0 {
				for _, subF := range f.bundle {
					subF.x += f.x
					subF.y += f.y
					newFList = append(newFList, subF)
				}
			}
		}
		fireworkList = newFList

		// Make Pallette Image
		newImg := image.NewPaletted(boundRect, colList)
		for x := 0; x < wid; x++ {
			for y := 0; y < height; y++ {
				newImg.SetColorIndex(x, y, uint8(colList.Index(rawImg.At(x, y))))
			}
		}

		// Add Lists
		if len(newList) == 0 {
			disposalList = append(disposalList, gif.DisposalNone)
		} else {
			disposalList = append(disposalList, gif.DisposalPrevious)
		}

		newList = append(newList, newImg)
		delayList = append(delayList, 10)

	}

	log.Println("Saving gif with ", len(newList), " frames")

	gifData := gif.GIF{
		Image:           newList,
		Delay:           delayList,
		Disposal:        disposalList,
		LoopCount:       -1,
		BackgroundIndex: 0,

		Config: image.Config{
			ColorModel: colList,
			Width:      wid,
			Height:     height,
		},
	}

	fn := MakeGifFilename(username, tid)
	f, e := os.Create(fn)
	if e != nil {
		log.Println(e)
		return "", e
	}

	e = gif.EncodeAll(f, &gifData)
	if e != nil {
		log.Println(e)
		return "", e
	}

	return fn, nil
}