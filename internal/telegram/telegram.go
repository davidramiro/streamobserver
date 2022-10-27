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
	"github.com/rs/zerolog/log"
)

var bot *tgbotapi.BotAPI

const twitchPrefix = "https://twitch.tv/"
const widthPlaceholder = "{width}"
const heightPlaceholder = "{height}"
const htmlSuffix = ".html"

func InitBot() {
	config, err := config.GetConfig()
	if err != nil {
		log.Panic().Err(err)
		return
	}

	bot, err = tgbotapi.NewBotAPI(config.Telegram.ApiKey)
	if err != nil {
		log.Panic().Err(err)
		return
	}

	bot.Debug = true

	logger.Log.Info().Msgf("Authorized on Telegram account %s", bot.Self.UserName)
}

func SendMessage(msg string) {
	if bot == nil {
		logger.Log.Error().Msg("Bot not initialized.")
		return
	}
	msgToSend := tgbotapi.NewMessage(434289657, msg)
	_, err := bot.Send(msgToSend)
	if err != nil {
		logger.Log.Error().Err(err)
	}
}

func SendTwitchStreamInfo(chatID int64, stream twitch.Stream) {
	if bot == nil {
		logger.Log.Error().Msg("Bot not initialized.")
		return
	}

	// newphoto := tgbotapi.NewPhoto(434289657, )

	formattedUrl := formatTwitchPhotoUrl(stream.ThumbnailURL)
	caption := "*" + stream.UserName + "* is playing *" + stream.GameName + "*: " + stream.Title + "\n" + twitchPrefix + stream.UserName

	photoMessage := createPhotoMessage(caption, chatID, formattedUrl)

	bot.Send(photoMessage)
}

func SendRestreamerStreamInfo(chatID int64, streamInfo restreamer.StreamInfo, stream restreamer.Stream) {
	if bot == nil {
		logger.Log.Error().Msg("Bot not initialized.")
		return
	}

	var streamLink string
	if stream.CustomURL == "" {
		streamLink = stream.BaseURL + "/" + stream.ID + htmlSuffix
	} else {
		streamLink = stream.CustomURL
	}

	caption := "*" + streamInfo.UserName + "* is streaming: " + streamInfo.Description + "\n" + streamLink

	photoMessage := createPhotoMessage(caption, chatID, streamInfo.ThumbnailURL)

	bot.Send(photoMessage)
}

func formatTwitchPhotoUrl(url string) string {
	resized := strings.Replace(url, heightPlaceholder, "1080", 1)
	resized = strings.Replace(resized, widthPlaceholder, "1920", 1)

	return resized
}

func createPhotoMessage(caption string, chatID int64, url string) tgbotapi.PhotoConfig {
	photoBytes, err := getPhotoFromUrl(url)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Could not send Photo on Telegram")
	}

	photoFileBytes := tgbotapi.FileBytes{
		Name:  "picture",
		Bytes: photoBytes,
	}
	config := tgbotapi.NewPhoto(chatID, photoFileBytes)
	config.Caption = caption
	config.ParseMode = tgbotapi.ModeMarkdown
	return config

}

func getPhotoFromUrl(url string) ([]byte, error) {
	logger.Log.Debug().Msg("Getting image from URL")
	response, e := http.Get(url)
	if e != nil {
		logger.Log.Fatal().Err(e)
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
