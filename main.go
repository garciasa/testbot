package main

// import "github.com/go-telegram-bot-api/telegram-bot-api"
import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"gopkg.in/telegram-bot-api.v4"
)

const bannerMsg = `Welcome to *Beaches.ie* Bot
I'm here to help you ☺️
type /help to see commands
`

const helpMsg = `
*Commands*
/list - Beaches near you
/restrictions - Restrictions near you
/search text - Search beaches by name
/county name - List beaches in a county
/favourites id1,id2,... - Adding favourites beaches to your profile
/jelly - take a photo, send it to us and our AI software is going to recognize it
/report - take a photo and report an beach incident
`

type beachesResponse struct {
	Name   string
	Code   string
	County string `json:"CountyName"`
}

func main() {

	apikey := os.Getenv("tgapikey")

	if apikey == "" {
		log.Panic("ApiKey not defined")
	}

	bot, err := tgbotapi.NewBotAPI(apikey)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			log.Printf("[Command] %s", update.Message.Command())
			msg, err := generateResponseToCmd(*update.Message)
			if err != nil {
				log.Printf("[ERROR] - %s", err)
				continue
			}

			bot.Send(msg)
			continue
		}

		if update.Message.Location != nil {
			// Response a location button
			log.Printf("[LOCATION] %f %f", update.Message.Location.Latitude, update.Message.Location.Longitude)
			// Api Call for getting the closest beaches for that location
			// TODO: go routine
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Thank you, we're checking...")
			bot.Send(msg)
			beaches, err := listBeachesNearMe(update.Message.Chat.ID, update.Message.Location.Latitude, update.Message.Location.Longitude)
			if err != nil {
				log.Printf("[ERROR] - %s", err)
			}

			for _, b := range beaches {
				txt := fmt.Sprintf("*Name:* %s\n*County:* %s\n*Code:* %s", b.Name, b.County, b.Code)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, txt)
				msg.ParseMode = "MARKDOWN"
				bot.Send(msg)
			}

			continue

		}

		// log.Printf("[%s] %s", update.Message.From.FirstName, update.Message.Text)

		// msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		// msg.ReplyToMessageID = update.Message.MessageID

		// bot.Send(msg)
	}

}

func generateResponseToCmd(cmd tgbotapi.Message) (msg tgbotapi.MessageConfig, err error) {
	switch action := cmd.Command(); action {
	case "start":
		msg = tgbotapi.NewMessage(cmd.Chat.ID, bannerMsg)
		msg.ParseMode = "MARKDOWN"
	case "help":
		msg = tgbotapi.NewMessage(cmd.Chat.ID, helpMsg)
		msg.ParseMode = "MARKDOWN"
	case "list":
		msg = tgbotapi.NewMessage(cmd.Chat.ID, "Give me your location")
		cardLocation := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButtonLocation("Get Location"),
			),
		)
		cardLocation.OneTimeKeyboard = true
		msg.ReplyMarkup = cardLocation
	default:
		return tgbotapi.MessageConfig{}, errors.New("No command found")
	}

	return msg, nil
}

func listBeachesNearMe(chatID int64, latitude float64, longitude float64) (beaches []beachesResponse, err error) {
	url := fmt.Sprintf("https://api.beaches.ie/api/beach/nearme/%f/%f/5", latitude, longitude)
	log.Printf("[INFO] - %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(resp.Body).Decode(&beaches)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] - Total %d", len(beaches))

	return beaches, nil

}
