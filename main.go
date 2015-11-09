package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"unicode"

	"image"
	"image/color"
	"image/gif"

	"github.com/ChimeraCoder/anaconda"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

const ()

var (
	debug = flag.Bool("debug", false, "Debug Flags On")
)

func init() {

}

type ClientSecret struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`

	AccessToken  string `json:"access_token"`
	AccessSecret string `json:"access_token_secret"`
}

func loadClientSecret(filename string) (*ClientSecret, error) {
	jsonBlob, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cs ClientSecret
	err = json.Unmarshal(jsonBlob, &cs)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

func startTwitterAPI() (*anaconda.TwitterApi, error) {
	secret, err := loadClientSecret("_secret.json")
	if err != nil {
		log.Fatalln("Secret Missing: %s", err)
		return nil, err
	}

	anaconda.SetConsumerKey(secret.Key)
	anaconda.SetConsumerSecret(secret.Secret)
	api := anaconda.NewTwitterApi(secret.AccessToken, secret.AccessSecret)
	return api, nil
}

type FireWork struct {
	x, y, dx, dy, t, sz, d float64
	colID                  uint8
	bundle                 []FireWork
	tweet                  *anaconda.Tweet
	text                   string
}

func hash(s string) int64 {
	h := fnv.New64()
	h.Write([]byte(s))
	return int64(h.Sum64())
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
		colID: uint8((twt.Id % 3) + 1),
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

func genTwitterGif(tweets []anaconda.Tweet) {
	wid := 440
	height := 220

	colList := color.Palette{
		color.RGBA{0x00, 0x00, 0x00, 0x00},
		color.RGBA{0xFF, 0x00, 0x00, 0xFF},
		color.RGBA{0x00, 0xFF, 0x00, 0xFF},
		color.RGBA{0x00, 0x00, 0xFF, 0xFF},
		color.RGBA{0x00, 0x00, 0x00, 0xFF},
		color.RGBA{0xFF, 0xFF, 0xFF, 0xFF},
	}

	newList := []*image.Paletted{}
	delayList := []int{}
	fireworkList := []FireWork{}

	draw2d.SetFontFolder("static")

	for i, _ := range tweets {
		f := genFireworkFromTweet(tweets, i, float64(wid), float64(height))
		fireworkList = append(fireworkList, f)
	}

	boundRect := image.Rect(0, 0, wid, height)

	for len(fireworkList) > 0 {
		rawImg := image.NewRGBA(boundRect)
		gc := draw2dimg.NewGraphicContext(rawImg)

		gc.SetFontData(draw2d.FontData{
			Name: "Roboto",
		})

		gc.SetFontSize(8)

		gc.Clear()

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

		newImg := image.NewPaletted(boundRect, colList)
		for x := 0; x < wid; x++ {
			for y := 0; y < height; y++ {
				newImg.SetColorIndex(x, y, uint8(colList.Index(rawImg.At(x, y))))
			}
		}

		newList = append(newList, newImg)
		delayList = append(delayList, 10)
		fireworkList = newFList
	}

	log.Println("Saving gif with ", len(newList), " frames")

	gifData := gif.GIF{
		Image:           newList,
		Delay:           delayList,
		LoopCount:       -1,
		BackgroundIndex: 0,
		Config: image.Config{
			ColorModel: colList,
			Width:      wid,
			Height:     height,
		},
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

func main() {

	api, _ := startTwitterAPI()
	// Search
	{
		search_result, err := api.GetSearch("golang", nil)
		if err != nil {
			panic(err)
		}
		for _, tweet := range search_result.Statuses {
			fmt.Println(">>", tweet.Text)
		}
	}

	// Homeline
	{
		v := url.Values{}
		v.Set("screen_name", "evilkimau")
		v.Set("count", "30")
		search_result, err := api.GetUserTimeline(v)

		if err != nil {
			panic(err)
		}

		for _, tweet := range search_result {
			fmt.Println(">>", tweet.User.Name, ":", tweet.Text, "\n\t", tweet.FavoriteCount, tweet.RetweetCount)
		}
		genTwitterGif(search_result)
	}

	log.Println("Generating")

}
