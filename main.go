package main

import (
	"context"
	"streamobserver/internal/adapter/broadcastbox"
	"streamobserver/internal/adapter/restreamer"
	"streamobserver/internal/adapter/telegram"
	"streamobserver/internal/adapter/twitch"
	"streamobserver/internal/core/domain"
	"streamobserver/internal/core/service"

	"github.com/go-telegram/bot"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func main() {
	log.Info().Str("author", "davidramiro").Msg("starting streamobserver")

	log.Info().Msg("initializing telegram bot")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Panic().Err(err).Msg("failed to read config")
	}

	if viper.GetBool("general.debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	token := viper.GetString("telegram.apikey")

	b, err := bot.New(token)
	if err != nil {
		log.Panic().Err(err).Msg("failed initializing telegram bot")
	}

	sender := telegram.NewTelegramSender(b)
	ta := &twitch.StreamInfoProvider{}
	ra := &restreamer.StreamInfoProvider{}
	bb := &broadcastbox.StreamInfoProvider{}

	streamService := service.NewStreamService(ta, ra, bb)

	notificationService := service.NewNotificationService(sender, streamService)

	var chats []domain.ChatConfig
	err = viper.UnmarshalKey("chats", &chats)
	if err != nil {
		log.Panic().Err(err).Msg("failed to unmarshal config")
	}

	for _, chat := range chats {
		for _, restreamerConfig := range chat.Streams.Restreamer {
			notificationService.Register(chat.ChatID, &domain.StreamQuery{
				UserID:    restreamerConfig.ID,
				BaseURL:   restreamerConfig.BaseURL,
				CustomURL: restreamerConfig.CustomURL,
				Kind:      domain.StreamKindRestreamer,
			})
		}
		for _, twitchConfig := range chat.Streams.Twitch {
			notificationService.Register(chat.ChatID, &domain.StreamQuery{
				UserID: twitchConfig.Username,
				Kind:   domain.StreamKindTwitch,
			})
		}
		for _, broadcastboxConfig := range chat.Streams.BroadcastBox {
			notificationService.Register(chat.ChatID, &domain.StreamQuery{
				UserID:  broadcastboxConfig.ID,
				BaseURL: broadcastboxConfig.BaseURL,
				Kind:    domain.StreamKindBroadcastBox,
			})
		}
	}

	notificationService.StartPolling(context.Background())
}
