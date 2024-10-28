package telegram

import (
	"context"
	"errors"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"streamobserver/internal/config"
	"streamobserver/internal/logger"
	"streamobserver/internal/restreamer"
	"streamobserver/internal/twitch"
	"streamobserver/internal/util"
	"strings"
)

var b *bot.Bot

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

	b, err = bot.New(config.Telegram.ApiKey)
	if err != nil {
		return err
	}

	user, err := b.GetMe(context.TODO())
	if err != nil {
		return err
	}

	logger.Log.Info().Msgf("Authorized on Telegram account %s", user.Username)
	return nil
}

// SendTwitchStreamInfo generates a message from a Twitch stream struct and sends it to a chat ID.
func SendTwitchStreamInfo(chatID int64, stream twitch.Stream) (models.Message, error) {
	if b == nil {
		logger.Log.Error().Msg("Bot not initialized.")
		return models.Message{}, errors.New("bot not initialized")
	}

	util.FormatTwitchPhotoUrl(&stream.ThumbnailURL)
	caption := stream.UserName + " is streaming " + stream.GameName + ": " + stream.Title + "\n" + twitchPrefix + stream.UserName + " [" + liveText + "]"

	photoMessage := createPhotoMessage(caption, chatID, stream.ThumbnailURL)
	ret, err := b.SendPhoto(context.TODO(), &photoMessage)
	logger.Log.Debug().Interface("Message", ret).Msg("Sent message.")

	if err != nil {
		logger.Log.Error().Err(err).Msg("error sending Telegram message")
		return models.Message{}, err
	}
	if ret.Chat.ID != chatID {
		return models.Message{}, errors.New("error sending Telegram message")
	}

	return *ret, nil
}

// SendRestreamerStreamInfo generates a message from a Restreamer stream struct and sends it to a chat ID.
func SendRestreamerStreamInfo(chatID int64, streamInfo restreamer.StreamInfo, stream restreamer.Stream) (models.Message, error) {
	if b == nil {
		logger.Log.Error().Msg("Bot not initialized.")
		return models.Message{}, errors.New("bot not initialized")
	}

	var streamLink string
	if stream.CustomURL == "" {
		streamLink = stream.BaseURL + "/" + stream.ID + htmlSuffix
	} else {
		streamLink = stream.CustomURL
	}

	caption := streamInfo.UserName + " is streaming: " + streamInfo.Description + "\n" + streamLink + " [" + liveText + "]"

	photoMessage := createPhotoMessage(caption, chatID, streamInfo.ThumbnailURL)
	ret, err := b.SendPhoto(context.TODO(), &photoMessage)
	logger.Log.Debug().Interface("Message", ret).Msg("Sent message.")

	if err != nil {
		logger.Log.Error().Err(err).Msg("error sending Telegram message")
		return models.Message{}, err
	}
	if ret.Chat.ID != chatID {
		return models.Message{}, errors.New("error sending Telegram message")
	}

	return *ret, nil
}

// SendUpdateTwitchStreamInfo updates a previously sent message with new stream info.
func SendUpdateTwitchStreamInfo(chatID int64, message models.Message, stream twitch.Stream) (models.Message, error) {
	if b == nil {
		logger.Log.Error().Msg("Bot not initialized.")
		return models.Message{}, errors.New("bot not initialized")
	}

	newcaption := stream.UserName + " is streaming " + stream.GameName + ": " + stream.Title + "\n" + twitchPrefix + stream.UserName + " [" + liveText + "]"

	ret, err := b.EditMessageCaption(context.TODO(), &bot.EditMessageCaptionParams{
		ChatID:    message.Chat.ID,
		MessageID: message.ID,
		Caption:   newcaption,
	})

	if err != nil {
		logger.Log.Error().Err(err).Msg("error updating Telegram message")
		return models.Message{}, err
	}
	if ret.Chat.ID != chatID {
		return models.Message{}, errors.New("error updating Telegram message")
	}

	return *ret, nil
}

// SendUpdateStreamOffline takes a previously sent message and edits it to reflect the changed stream status.
func SendUpdateStreamOffline(message models.Message, chatID int64) (models.Message, error) {
	logger.Log.Debug().Interface("Message", message).Msg("Updating Message")

	newtext := strings.Replace(message.Caption, liveText, offlineText, 1)
	newtext = strings.Replace(newtext, "is streaming", "was streaming", 1)

	ret, err := b.EditMessageCaption(context.TODO(), &bot.EditMessageCaptionParams{
		ChatID:    message.Chat.ID,
		MessageID: message.ID,
		Caption:   newtext,
	})

	if err != nil {
		logger.Log.Error().Err(err).Msg("error updating Telegram message")
		return models.Message{}, err
	}
	if ret.Chat.ID != chatID {
		return models.Message{}, errors.New("error updating Telegram message")
	}

	return *ret, nil
}

func createPhotoMessage(caption string, chatID int64, url string) bot.SendPhotoParams {
	return bot.SendPhotoParams{
		ChatID:  chatID,
		Photo:   &models.InputFileString{Data: url},
		Caption: caption,
	}
}
