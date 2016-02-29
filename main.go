package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"image"
	"image/color"
	"image/gif"

	"github.com/ChimeraCoder/anaconda"
	"github.com/golang/freetype/raster"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

const (
	NOOF_COLOURS    = 6
	MINUTES_TO_WAIT = 2
)

var (
	debug = flag.Bool("debug", false, "Debug Flags On")
	live  = flag.Bool("live", true, "Actual Live Tweet")
)

func init() {
	flag.Parse()
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

	fn := fmt.Sprintf("%s_%d.gif", username, tid)
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

func postImageTweet(api *anaconda.TwitterApi, gifFile string, t *anaconda.Tweet) error {
	// Post

	data, err := ioutil.ReadFile(gifFile)
	if err != nil {
		return err
	}

	mediaResponse, err := api.UploadMedia(base64.StdEncoding.EncodeToString(data))
	if err != nil {
		return err
	}

	v := url.Values{}
	v.Set("media_ids", strconv.FormatInt(mediaResponse.MediaID, 10))
	v.Set("in_reply_to_status_id", t.IdStr)

	tweetString := fmt.Sprintf("@%s here are your fireworks", t.User.ScreenName)

	_, err = api.PostTweet(tweetString, v)
	if err != nil {
		return err
	} else {
		// fmt.Println(result)
	}

	return nil
}

func Exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func GenerateFireworkFor(api *anaconda.TwitterApi, t *anaconda.Tweet) error {

	checkFile := fmt.Sprintf("%s_%d.gif", t.User.ScreenName, t.Id)
	if Exists(checkFile) {
		fmt.Println("Already Exsists")
		return nil
	}

	v := url.Values{}
	v.Set("screen_name", t.User.ScreenName)
	v.Set("count", "30")
	search_result, err := api.GetUserTimeline(v)
	if err != nil {
		return err
	}

	gifFile, e := genTwitterGif(search_result, t.User.ScreenName, t.Id)
	if e != nil {
		return e
	}

	if *live {
		return postImageTweet(api, gifFile, t)
	} else {
		fmt.Println("Not live: ", live, t.User.ScreenName, gifFile)
	}

	return nil
}

func main() {

	api, _ := startTwitterAPI()

	var startTime = time.Now()

	// Homeline
	/*
	v := url.Values{}
	v.Set("count", "30")
	search_result, e2 := api.GetHomeTimeline(v)
	if e2 != nil {
		fmt.Println(e2)
	}

	for _, tweet := range search_result {		fmt.Println(">>", tweet.User.ScreenName, ":", "\t", tweet.FavoriteCount, tweet.RetweetCount)}
*/
	
	// Start Up Tweet
	{
		v := url.Values{}
		_, err := api.PostTweet("@evilkimau Good Morning. I had to restart.", v)
		if err != nil {
			fmt.Println(err)
			return
		} else {
		}
	}

	// Refresh Loop
	var lastId int64 = 0
	var err error
	var hasNewBits bool = true
	var loopMe = true
	for loopMe {
		// Sleep
		time.Sleep(time.Minute * MINUTES_TO_WAIT)

		if hasNewBits {
			fmt.Printf("\nRefreshing")
			hasNewBits = false
		}

		// Get Mentions
		v := url.Values{}
		v.Set("count", "15")
		if lastId != 0 {
			v.Set("since_id", strconv.FormatInt(lastId, 10))
		}

		// Tweets
		var tweets []anaconda.Tweet
		tweets, err = api.GetMentionsTimeline(v)
		if len(tweets) > 0 {
			fmt.Printf("\nRetrieved %d mentions. \n", len(tweets))
			hasNewBits = true
		} else {
			fmt.Printf(".")
		}
		if err != nil {
			fmt.Println(err)
			continue
		}

		mentionMap := make(map[string]int64)

		for _, t := range tweets {
			// Get Last ID
			if lastId < t.Id {
				lastId = t.Id
			}

			ttime, _ := t.CreatedAtTime()
			timeDiff := startTime.Sub(ttime)

			if timeDiff > 0 {
				// Old Tweet
				if timeDiff > time.Hour {
					fmt.Printf("Ignoring tweet from %s because its from %.0f hours ago \n", t.User.ScreenName, timeDiff.Hours())
				} else if timeDiff > time.Minute {
					fmt.Printf("Ignoring tweet from %s because its from %.0f minutes ago \n", t.User.ScreenName, timeDiff.Minutes())
				} else {
					fmt.Printf("Ignoring tweet from %s because its from %.0f seconds ago \n", t.User.ScreenName, timeDiff.Seconds())
				}
				continue
			}

			fmt.Printf("%s is %d \n", ttime.Sub(startTime) < 0, ttime.Sub(startTime))

			v, ok := mentionMap[t.User.ScreenName]
			if ok && v >= t.Id {
				// Already Gen
			} else {
				// Generate Fireworks
				mentionMap[t.User.ScreenName] = t.Id
				fmt.Println("Generate Fireworks for ", t.User.ScreenName, t.Id)
				err = GenerateFireworkFor(api, &t)
			}

			// NEXT Twet
		}

		// Next Round
	}
}
