package telegram

import (
	"context"
	"errors"
	"fmt"
	"streamobserver/internal/core/domain"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

const (
	liveText    = "🔴 LIVE"
	offlineText = "❌ OFFLINE"
)

type Sender struct {
	b *bot.Bot
}

func NewTelegramSender(b *bot.Bot) *Sender {
	return &Sender{b: b}
}

// SendStreamInfo generates a message from a domain.StreamInfo and sends it to a chat ID.
func (s *Sender) SendStreamInfo(ctx context.Context, chatID int64, stream domain.StreamInfo) (int, error) {
	var viewerInfo string
	if stream.ViewerCount > -1 {
		viewerInfo = fmt.Sprintf("for %d viewers", stream.ViewerCount)
	}

	caption := fmt.Sprintf("%s is streaming %s %s\n%s\n[%s]",
		stream.Username, stream.Title, viewerInfo, stream.URL, liveText)

	var message *models.Message
	var err error
	if stream.ThumbnailURL == "" {
		message, err = s.b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   caption,
		})
		if err != nil {
			return -1, fmt.Errorf("error sending telegram message: %w", err)
		}
	} else {
		message, err = s.b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID: chatID,
			Photo: &models.InputFileString{
				Data: fmt.Sprintf(
					"%s?time=%d",
					stream.ThumbnailURL,
					time.Now().Unix()),
			},
			Caption: caption,
		})
		if err != nil {
			return -1, fmt.Errorf("error sending telegram photo: %w", err)
		}
	}

	log.Debug().Interface("Message", message).Msg("Sent message.")

	if message.Chat.ID != chatID {
		return -1, errors.New("returned invalid chat id")
	}

	return message.ID, nil
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
	var viewerInfo string
	if stream.ViewerCount > -1 {
		viewerInfo = fmt.Sprintf("for %d viewers", stream.ViewerCount)
	}

	caption := fmt.Sprintf("%s %s streaming %s %s\n%s\n[%s]", stream.Username, verb, stream.Title,
		viewerInfo, stream.URL, status)

	var message *models.Message
	var err error
	if stream.ThumbnailURL == "" {
		message, err = s.b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: messageID,
			Text:      caption,
		})
		if err != nil {
			return fmt.Errorf("error editing telegram message: %w", err)
		}
	} else {
		message, err = s.b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{
			ChatID:    chatID,
			MessageID: messageID,
			Caption:   caption,
		})
		if err != nil {
			return fmt.Errorf("error editing telegram photo: %w", err)
		}
	}

	log.Debug().Interface("Message", *message).Msg("Sent message.")

	if message.Chat.ID != chatID {
		return errors.New("returned invalid chat id")
	}

	return nil
}
