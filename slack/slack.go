package slack

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/nlopes/slack"
)

/*
   TODO: Change @BOT_NAME to the same thing you entered when creating your Slack application.
   NOTE: command_arg_1 and command_arg_2 represent optional parameteras that you define
   in the Slack API UI
*/
const helpMessage = "type in '@BOT_NAME <command_arg_1> to get a random gif of that query, prefix your query with specific if you want the first result'"

// Struct for taking in image data from GIPHY
type ImageData struct {
	URL    string `json:"url"`
	Width  string `json:"width"`
	Height string `json:"height"`
	Size   string `json:"size"`
	Frames string `json:"frames"`
}

// Struct for taking in Gif data from GIPHY
type Gif struct {
	Type               string `json:"type"`
	Id                 string `json:"id"`
	URL                string `json:"url"`
	Tags               string `json:"tags"`
	BitlyGifURL        string `json:"bitly_gif_url"`
	BitlyFullscreenURL string `json:"bitly_fullscreen_url"`
	BitlyTiledURL      string `json:"bitly_tiled_url"`
	EmbedURL           string `json:"embed_url`
	Images             struct {
		Original               ImageData `json:"original"`
		FixedHeight            ImageData `json:"fixed_height"`
		FixedHeightStill       ImageData `json:"fixed_height_still"`
		FixedHeightDownsampled ImageData `json:"fixed_height_downsampled"`
		FixedWidth             ImageData `json:"fixed_width"`
		FixedwidthStill        ImageData `json:"fixed_width_still"`
		FixedwidthDownsampled  ImageData `json:"fixed_width_downsampled"`
	} `json:"images"`
}

// Struct for taking in pagination from GIPHY
type paginatedResults struct {
	Data       []*Gif `json:"data"`
	Pagination struct {
		TotalCount int `json:"total_count"`
	} `json:"pagination"`
}

// Struct for taking a single result from GIPHY
type singleResult struct {
	Data *Gif `json:"data"`
}

//  Creates the slack connection and the realtime manager
func CreateSlackClient(apiKey string) *slack.RTM {
	api := slack.New(apiKey)
	rtm := api.NewRTM()
	go rtm.ManageConnection() // goroutine!
	return rtm
}

//  Checks all incoming events to determine whether to respond to them
func RespondToEvents(slackClient *slack.RTM) {
	for msg := range slackClient.IncomingEvents {
		fmt.Println("Event Received: ", msg.Type)
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			botTagString := fmt.Sprintf("<@%s> ", slackClient.GetInfo().User.ID)
			if !strings.Contains(ev.Msg.Text, botTagString) {
				continue
			}
			message := strings.Replace(ev.Msg.Text, botTagString, "", -1)
			sendResponse(slackClient, message, ev.Channel)
			sendHelp(slackClient, message, ev.Channel)
		default:

		}
	}
}

// Responds with a help message in slack
func sendHelp(slackClient *slack.RTM, message, slackChannel string) {
	if strings.ToLower(message) != "help" {
		return
	}
	slackClient.SendMessage(slackClient.NewOutgoingMessage(helpMessage, slackChannel))
}

// Sends a gif in slack according to user input
func sendResponse(slackClient *slack.RTM, message, slackChannel string) {
	specific := !strings.HasPrefix(message, "random")
	message = strings.ToLower(message)
	message = strings.ReplaceAll(message, " ", "%20")

	if specific {
		message = strings.TrimPrefix(message, "random%20")
	}

	println("[RECEIVED] sendResponse:", message)
	url := fmt.Sprintf("http://api.giphy.com/v1/gifs/search?api_key=%s&q=%s&limit=5", os.Getenv("API_KEY"), message)
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Fatal("NewRequest: ", err)
		return
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return
	}

	defer resp.Body.Close()

	var data paginatedResults

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Println(err)
	}
	if specific {
		slackClient.SendMessage(slackClient.NewOutgoingMessage(data.Data[0].Images.FixedHeightDownsampled.URL, slackChannel))
	} else {
		slackClient.SendMessage(slackClient.NewOutgoingMessage(data.Data[rand.Intn(4)].Images.FixedHeightDownsampled.URL, slackChannel))
	}
}
