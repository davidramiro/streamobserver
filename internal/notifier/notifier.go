package notifier

import (
	"errors"
	"os"
	"path/filepath"
	"streamobserver/internal/logger"
	"streamobserver/internal/restreamer"
	"streamobserver/internal/telegram"
	"streamobserver/internal/twitch"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/yaml.v3"
)

type telegramChat struct {
	twitchStreams     []twitchStream
	restreamerStreams []restreamerStream
	chatID            int64
}

type twitchStream struct {
	username            string
	title               string
	game                string
	status              bool
	notified            bool
	notificationMessage tgbotapi.Message
}

type restreamerStream struct {
	stream              restreamer.Stream
	status              bool
	notified            bool
	notificationMessage tgbotapi.Message
}

type notifierConfig struct {
	Chats []*chatConfig `yaml:"chats"`
}

type chatConfig struct {
	ChatID  int64 `yaml:"chatid"`
	Streams struct {
		Twitch []struct {
			Username string `yaml:"username"`
		} `yaml:"twitch"`

		Restreamer []struct {
			BaseURL   string `yaml:"baseurl"`
			ID        string `yaml:"id"`
			CustomURL string `yaml:"customurl"`
		} `yaml:"restreamer"`
	} `yaml:"streams"`
}

var chats = []telegramChat{}

// PopulateObservers parses the streams config file.
func PopulateObservers() error {
	config := &notifierConfig{}

	// Open streams.yml
	p := filepath.FromSlash("./streams.yml")
	logger.Log.Debug().Str("Path", p).Msg("Reading streams config from disk")
	file, err := os.Open(p)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode YAML
	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return err
	}

	// Chats from Config
	for _, s := range config.Chats {
		chat := &telegramChat{}
		chat.chatID = s.ChatID
		// Twitch streams for each chat
		for _, tw := range s.Streams.Twitch {
			twitchData := new(twitchStream)
			twitchData.username = tw.Username
			twitchData.notified = false
			twitchData.status = false

			chat.twitchStreams = append(chat.twitchStreams, *twitchData)
		}

		// Restreamer streams for each chat
		for _, rs := range s.Streams.Restreamer {
			restreamerData := new(restreamerStream)
			restreamerData.stream = restreamer.Stream{
				BaseURL:   rs.BaseURL,
				ID:        rs.ID,
				CustomURL: rs.CustomURL,
			}
			restreamerData.notified = false
			restreamerData.status = false
			chat.restreamerStreams = append(chat.restreamerStreams, *restreamerData)
		}

		chats = append(chats, *chat)
	}

	if len(chats) == 0 || (len(chats[0].restreamerStreams) == 0 && len(chats[0].twitchStreams) == 0) {
		return errors.New("no chats/streams loaded")
	}

	return nil
}

// Notify starts the notification process and checks the configured streams for updates.
func Notify() {
	logger.Log.Debug().Msg("Started notify iteration")
	for i := range chats {
		for j := range chats[i].twitchStreams {
			checkAndNotifyTwitch(&chats[i].twitchStreams[j], chats[i].chatID)
		}
		for j := range chats[i].restreamerStreams {
			checkAndNotifyRestreamer(&chats[i].restreamerStreams[j], chats[i].chatID)
		}
	}

}

func checkAndNotifyTwitch(streamToCheck *twitchStream, chatID int64) {
	logger.Log.Info().Str("Channel", streamToCheck.username).Msg("Twitch: Checking status")

	// Get stream status and metadata from API
	streamResponse, err := twitch.GetStreams([]string{streamToCheck.username})
	if err != nil {
		logger.Log.Error().Err(errors.New("error getting info from Twitch API"))
		return
	}

	// List populated means stream is online
	if len(streamResponse) > 0 {
		if !streamToCheck.status {
			logger.Log.Debug().Str("Channel", streamToCheck.username).Msg("Online and status has changed.")
			streamToCheck.status = true
			if !streamToCheck.notified {
				logger.Log.Debug().Str("Channel", streamToCheck.username).Msg("Status change and notification needed.")
				var err error
				streamToCheck.notificationMessage, err = telegram.SendTwitchStreamInfo(chatID, streamResponse[0])
				if err != nil {
					logger.Log.Error().Int64("ChatID", chatID).Str("Channel", streamToCheck.username).Err(err).Msg("Could not notify about channel status.")
					streamToCheck.notified = false
					streamToCheck.status = false
				}
				streamToCheck.notified = true
				streamToCheck.game = streamResponse[0].GameName
				streamToCheck.title = streamResponse[0].Title
			}
		} else {
			logger.Log.Debug().Str("Channel", streamToCheck.username).Msg("Online and status has not changed.")
			if streamToCheck.game != streamResponse[0].GameName || streamToCheck.title != streamResponse[0].Title {
				telegram.SendUpdateTwitchStreamInfo(chatID, streamToCheck.notificationMessage, streamResponse[0])
			}

		}
	} else {
		if streamToCheck.status && streamToCheck.notified && streamToCheck.notificationMessage.MessageID != 0 {
			telegram.SendUpdateStreamOffline(streamToCheck.notificationMessage, chatID)
		}
		logger.Log.Debug().Str("Channel", streamToCheck.username).Msg("Channel offline.")
		streamToCheck.status = false
		streamToCheck.notified = false
	}
}

func checkAndNotifyRestreamer(streamToCheck *restreamerStream, chatID int64) {
	logger.Log.Info().Str("Channel", streamToCheck.stream.ID).Msg("Restreamer: Checking status")
	online, err := restreamer.CheckStreamLive(streamToCheck.stream)
	if err != nil {
		logger.Log.Error().Err(err).Msg("error getting stream status from restreamer")
		return
	}
	if online {
		streamInfo, err := restreamer.GetStreamInfo(streamToCheck.stream)
		if err != nil {
			logger.Log.Error().Err(err).Msg("error getting stream info from restreamer")
			return
		}
		if !streamToCheck.status {
			logger.Log.Debug().Str("Channel", streamInfo.UserName).Msg("Online and status has changed.")
			streamToCheck.status = true
			if !streamToCheck.notified {
				logger.Log.Debug().Str("Channel", streamInfo.UserName).Msg("Status change and notification needed.")
				var err error
				streamToCheck.notificationMessage, err = telegram.SendRestreamerStreamInfo(chatID, streamInfo, streamToCheck.stream)
				if err != nil {
					logger.Log.Error().Int64("ChatID", chatID).Str("Channel", streamToCheck.stream.ID).Err(err).Msg("Could not notify about channel status.")
					streamToCheck.notified = false
					streamToCheck.status = false
				}
				streamToCheck.notified = true
			}
		} else {
			logger.Log.Debug().Str("Channel", streamInfo.UserName).Msg("Online and status has not changed.")
		}
	} else {
		if streamToCheck.status && streamToCheck.notified && streamToCheck.notificationMessage.MessageID != 0 {
			telegram.SendUpdateStreamOffline(streamToCheck.notificationMessage, chatID)
		}
		logger.Log.Debug().Str("Channel", streamToCheck.stream.ID).Msg("Channel offline.")
		streamToCheck.status = false
		streamToCheck.notified = false
	}
}
