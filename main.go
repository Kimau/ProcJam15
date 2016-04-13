package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/ChimeraCoder/anaconda"
)

const (
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

func hash(s string) int64 {
	h := fnv.New64()
	h.Write([]byte(s))
	return int64(h.Sum64())
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

func MakeGifFilename(screenname string, tweetId int64) string {
  return fmt.Sprintf("_%s_%d.gif", screenname, tweetId)
}

func GenerateFireworkFor(api *anaconda.TwitterApi, t *anaconda.Tweet) error {

	checkFile := MakeGifFilename(t.User.ScreenName, t.Id)
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
