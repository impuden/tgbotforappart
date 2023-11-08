package kbbutton

import (
	"encoding/json"
	"net/http"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func FetchCatImageURL() (string, error) {
	url := "https://api.thecatapi.com/v1/images/search"

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var result []struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", err
	}

	return result[0].URL, nil
}

func Cat(bot *tgbotapi.BotAPI, chatID int64) error {
	catURL, err := FetchCatImageURL()
	if err != nil {
		return err
	}

	response, err := http.Get(catURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	image := tgbotapi.NewPhotoShare(chatID, catURL)
	_, err = bot.Send(image)

	return err
}
