package telegram

import (
	"errors"
	"io"
	"net/http"
	"streamobserver/internal/config"
	"streamobserver/internal/logger"
	"streamobserver/internal/restreamer"
	"streamobserver/internal/twitch"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *tgbotapi.BotAPI

const twitchPrefix = "https://twitch.tv/"
const widthPlaceholder = "{width}"
const heightPlaceholder = "{height}"
const htmlSuffix = ".html"
const liveText = "🔴 LIVE"
const offlineText = "❌ OFFLINE"

func InitBot() {
	config, err := config.GetConfig()
	if err != nil {
		logger.Log.Panic().Err(err)
		return
	}

	bot, err = tgbotapi.NewBotAPI(config.Telegram.ApiKey)
	if err != nil {
		logger.Log.Panic().Err(err)
		return
	}

	bot.Debug = true

	logger.Log.Info().Msgf("Authorized on Telegram account %s", bot.Self.UserName)
}

func SendTwitchStreamInfo(chatID int64, stream twitch.Stream) (tgbotapi.Message, error) {
	if bot == nil {
		logger.Log.Error().Msg("Bot not initialized.")
		return tgbotapi.Message{}, nil
	}

	formattedUrl := formatTwitchPhotoUrl(stream.ThumbnailURL)
	caption := "*" + stream.UserName + "* is playing *" + stream.GameName + "*: " + stream.Title + "\n" + twitchPrefix + stream.UserName + " [" + liveText + "]"

	photoMessage, err := createPhotoMessage(caption, chatID, formattedUrl)
	if err != nil {
		return tgbotapi.Message{}, errors.New("could not send message")
	}

	ret, _ := bot.Send(photoMessage)
	logger.Log.Debug().Interface("Message", ret).Msg("Sent message.")

	return ret, nil
}

func SendRestreamerStreamInfo(chatID int64, streamInfo restreamer.StreamInfo, stream restreamer.Stream) (tgbotapi.Message, error) {
	if bot == nil {
		logger.Log.Error().Msg("Bot not initialized.")
		return tgbotapi.Message{}, errors.New("bot not initialized")
	}

	var streamLink string
	if stream.CustomURL == "" {
		streamLink = stream.BaseURL + "/" + stream.ID + htmlSuffix
	} else {
		streamLink = stream.CustomURL
	}

	caption := "*" + streamInfo.UserName + "* is streaming: " + streamInfo.Description + "\n" + streamLink + " [" + liveText + "]"

	photoMessage, err := createPhotoMessage(caption, chatID, streamInfo.ThumbnailURL)
	if err != nil {
		return tgbotapi.Message{}, errors.New("could not send message")
	}

	ret, _ := bot.Send(photoMessage)
	logger.Log.Debug().Interface("Message", ret).Msg("Sent message.")

	return ret, nil
}

func UpdateStreamMessage(message tgbotapi.Message, chatID int64) {
	logger.Log.Debug().Interface("Message", message).Msg("Updating Message")

	newtext := strings.Replace(message.Caption, liveText, offlineText, 1)
	config := tgbotapi.NewEditMessageCaption(chatID, message.MessageID, newtext)
	bot.Send(config)
}

func formatTwitchPhotoUrl(url string) string {
	resized := strings.Replace(url, heightPlaceholder, "1080", 1)
	resized = strings.Replace(resized, widthPlaceholder, "1920", 1)

	return resized
}

func createPhotoMessage(caption string, chatID int64, url string) (tgbotapi.PhotoConfig, error) {
	photoBytes, err := getPhotoFromUrl(url)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Could not send Photo on Telegram")
		return tgbotapi.PhotoConfig{}, errors.New("sending photo failed")
	}

	photoFileBytes := tgbotapi.FileBytes{
		Name:  "picture",
		Bytes: photoBytes,
	}
	config := tgbotapi.NewPhoto(chatID, photoFileBytes)
	config.Caption = caption
	config.ParseMode = tgbotapi.ModeMarkdown
	return config, nil

}

func getPhotoFromUrl(url string) ([]byte, error) {
	logger.Log.Debug().Msg("Getting image from URL")
	response, e := http.Get(url)
	if e != nil {
		logger.Log.Fatal().Err(e)
		return nil, errors.New("could not fetch image from URL")
	}
	if response.StatusCode != 200 {
		return nil, errors.New("could not fetch image from URL")
	}

	defer response.Body.Close()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Log.Fatal().Err(err)
		return nil, errors.New("could not read image bytes")
	}

	return bytes, nil
}
