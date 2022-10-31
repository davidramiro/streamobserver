package telegram

import (
	"errors"
	"streamobserver/internal/config"
	"streamobserver/internal/logger"
	"streamobserver/internal/restreamer"
	"streamobserver/internal/twitch"
	"streamobserver/internal/util"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *tgbotapi.BotAPI

const (
	twitchPrefix = "https://twitch.tv/"
	htmlSuffix   = ".html"
	liveText     = "🔴 LIVE"
	offlineText  = "❌ OFFLINE"
)

// InitBot initializes the Telegram bot to send updates with.
func InitBot(debug bool) error {
	config, err := config.GetConfig()
	if err != nil {
		return err
	}

	bot, err = tgbotapi.NewBotAPI(config.Telegram.ApiKey)
	if err != nil {
		return err
	}

	bot.Debug = debug
	logger.Log.Info().Msgf("Authorized on Telegram account %s", bot.Self.UserName)
	return nil
}

// SendTwitchStreamInfo generates a message from a Twitch stream struct and sends it to a chat ID.
func SendTwitchStreamInfo(chatID int64, stream twitch.Stream) (tgbotapi.Message, error) {
	if bot == nil {
		logger.Log.Error().Msg("Bot not initialized.")
		return tgbotapi.Message{}, errors.New("bot not initialized")
	}

	util.FormatTwitchPhotoUrl(&stream.ThumbnailURL)
	caption := "<b>" + stream.UserName + "</b> is streaming <b>" + stream.GameName + "</b>: " + stream.Title + "\n" + twitchPrefix + stream.UserName + " [" + liveText + "]"

	photoMessage, err := createPhotoMessage(caption, chatID, stream.ThumbnailURL)
	if err != nil {
		return tgbotapi.Message{}, errors.New("could not send message")
	}

	ret, err := bot.Send(photoMessage)
	logger.Log.Debug().Interface("Message", ret).Msg("Sent message.")

	if err != nil {
		logger.Log.Error().Err(err).Msg("error sending Telegram message")
		return tgbotapi.Message{}, err
	}
	if ret.Chat.ID != chatID {
		return tgbotapi.Message{}, errors.New("error sending Telegram message")
	}

	return ret, nil
}

// SendRestreamerStreamInfo generates a message from a Restreamer stream struct and sends it to a chat ID.
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

	caption := "<b>" + streamInfo.UserName + "</b> is streaming: " + streamInfo.Description + "\n" + streamLink + " [" + liveText + "]"

	photoMessage, err := createPhotoMessage(caption, chatID, streamInfo.ThumbnailURL)
	if err != nil {
		return tgbotapi.Message{}, errors.New("could not send message")
	}

	ret, err := bot.Send(photoMessage)
	logger.Log.Debug().Interface("Message", ret).Msg("Sent message.")

	if err != nil {
		logger.Log.Error().Err(err).Msg("error sending Telegram message")
		return tgbotapi.Message{}, err
	}
	if ret.Chat.ID != chatID {
		return tgbotapi.Message{}, errors.New("error sending Telegram message")
	}

	return ret, nil
}

// SendUpdateTwitchStreamInfo updates a previously sent message with new stream info.
func SendUpdateTwitchStreamInfo(chatID int64, message tgbotapi.Message, stream twitch.Stream) (tgbotapi.Message, error) {
	if bot == nil {
		logger.Log.Error().Msg("Bot not initialized.")
		return tgbotapi.Message{}, errors.New("bot not initialized")
	}

	newcaption := "<b>" + stream.UserName + "</b> is streaming <b>" + stream.GameName + "</b>: " + stream.Title + "\n" + twitchPrefix + stream.UserName + " [" + liveText + "]"

	config := tgbotapi.NewEditMessageCaption(chatID, message.MessageID, newcaption)
	config.ParseMode = tgbotapi.ModeHTML
	ret, err := bot.Send(config)

	if err != nil {
		logger.Log.Error().Err(err).Msg("error updating Telegram message")
		return tgbotapi.Message{}, err
	}
	if ret.Chat.ID != chatID {
		return tgbotapi.Message{}, errors.New("error updating Telegram message")
	}

	return ret, nil
}

// SendUpdateStreamOffline takes a previously sent message and edits it to reflect the changed stream status.
func SendUpdateStreamOffline(message tgbotapi.Message, chatID int64) (tgbotapi.Message, error) {
	logger.Log.Debug().Interface("Message", message).Msg("Updating Message")

	newtext := strings.Replace(message.Caption, liveText, offlineText, 1)
	newtext = strings.Replace(newtext, "is streaming", "was streaming", 1)
	config := tgbotapi.NewEditMessageCaption(chatID, message.MessageID, newtext)
	config.ParseMode = tgbotapi.ModeHTML
	config.CaptionEntities = message.CaptionEntities

	ret, err := bot.Send(config)

	if err != nil {
		logger.Log.Error().Err(err).Msg("error updating Telegram message")
		return tgbotapi.Message{}, err
	}
	if ret.Chat.ID != chatID {
		return tgbotapi.Message{}, errors.New("error updating Telegram message")
	}

	return ret, nil
}

func createPhotoMessage(caption string, chatID int64, url string) (tgbotapi.PhotoConfig, error) {
	photoBytes, err := util.GetPhotoFromUrl(url)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Could not send photo on Telegram")
		return tgbotapi.PhotoConfig{}, errors.New("could not retrieve photo")
	}

	photoFileBytes := tgbotapi.FileBytes{
		Name:  "picture",
		Bytes: photoBytes,
	}
	config := tgbotapi.NewPhoto(chatID, photoFileBytes)
	config.Caption = caption
	config.ParseMode = tgbotapi.ModeHTML
	return config, nil

}
