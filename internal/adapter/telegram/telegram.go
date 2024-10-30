package telegram

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
	"streamobserver/internal/core/domain"
)

const (
	twitchPrefix = "https://twitch.tv/"
	htmlSuffix   = ".html"
	liveText     = "🔴 LIVE"
	offlineText  = "❌ OFFLINE"
)

type Sender struct {
	b *bot.Bot
}

func NewTelegramSender(b *bot.Bot) *Sender {
	return &Sender{b: b}
}

// SendStreamInfo generates a message from a domain.StreamInfo and sends it to a chat ID.
func (s *Sender) SendStreamInfo(ctx context.Context, chatID int64, stream domain.StreamInfo) (int, error) {

	caption := fmt.Sprintf("%s is streaming %s\n%s\n[%s]", stream.Username, stream.Title, stream.URL, liveText)

	ret, err := s.b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:  chatID,
		Photo:   &models.InputFileString{Data: stream.ThumbnailURL},
		Caption: caption,
	})
	if err != nil {
		log.Error().Err(err).Msg("error sending telegram message")
		return -1, err
	}

	log.Debug().Interface("Message", ret).Msg("Sent message.")

	if ret.Chat.ID != chatID {
		return -1, errors.New("returned invalid chat id")
	}

	return ret.ID, nil
}

// UpdateStreamInfo generates a message from a domain.StreamInfo and sends it to a chat ID.
func (s *Sender) UpdateStreamInfo(ctx context.Context, chatID int64, messageID int, stream domain.StreamInfo) error {
	var verb string
	var status string

	if stream.IsOnline {
		verb = "is"
		status = liveText
	} else {
		verb = "was"
		status = offlineText
	}
	caption := fmt.Sprintf("%s %s streaming %s\n%s\n[%s]", stream.Username, verb, stream.Title, stream.URL, status)

	ret, err := s.b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{
		ChatID:    chatID,
		MessageID: messageID,
		Caption:   caption,
	})
	if err != nil {
		log.Error().Err(err).Msg("error sending telegram message")
		return err
	}

	log.Debug().Interface("Message", ret).Msg("Sent message.")

	if ret.Chat.ID != chatID {
		return errors.New("returned invalid chat id")
	}

	return nil
}
