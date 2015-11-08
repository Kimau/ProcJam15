package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"image"
	"image/color"
	"image/gif"

	"github.com/ChimeraCoder/anaconda"
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

func main() {

	api, _ := startTwitterAPI()
	search_result, err := api.GetSearch("golang", nil)
	if err != nil {
		panic(err)
	}
	for _, tweet := range search_result.Statuses {
		fmt.Println(">>", tweet.Text)
	}

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
		Image:           newI,
		Delay:           []int{10, 10, 10, 10, 10},
		LoopCount:       -1,
		BackgroundIndex: 0,
		Config: image.Config{
			ColorModel: colList,
			Width:      100,
			Height:     100,
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
