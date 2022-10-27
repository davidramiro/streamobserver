package notifier

import (
	"os"
	"path/filepath"
	"streamobserver/internal/logger"
	"streamobserver/internal/restreamer"
	"streamobserver/internal/telegram"
	"streamobserver/internal/twitch"

	"gopkg.in/yaml.v2"
)

type TelegramChat struct {
	TwitchStreams     []TwitchStream
	RestreamerStreams []RestreamerStream
	ChatID            int64
}

type TwitchStream struct {
	username string
	Status   bool
	Notified bool
}

type RestreamerStream struct {
	Stream   restreamer.Stream
	Status   bool
	Notified bool
}

type NotifierConfig struct {
	Chats []*ChatConfig `yaml:"chats"`
}

type ChatConfig struct {
	ChatID  int64 `yaml:"chatid"`
	Streams struct {
		Twitch []struct {
			Username string `yaml:"username"`
		} `yaml:"twitch"`

		Restreamer []struct {
			BaseURL string `yaml:"baseurl"`
			ID      string `yaml:"id"`
		} `yaml:"restreamer"`
	} `yaml:"streams"`
}

var streams = make([]TelegramChat, 1)

func PopulateObservers() {
	config := &NotifierConfig{}

	// Open streams.yml
	p := filepath.FromSlash("./streams.yml")
	logger.Log.Debug().Str("Path", p).Msg("Reading streams config from disk")
	file, err := os.Open(p)
	if err != nil {
		logger.Log.Error().Err(err)
	}
	defer file.Close()

	// Decode YAML
	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		logger.Log.Error().Err(err)
	}

	// Chats from Config
	for _, s := range config.Chats {
		chat := &TelegramChat{}
		chat.ChatID = s.ChatID
		// Twitch streams for each chat
		for _, tw := range s.Streams.Twitch {
			twitchData := new(TwitchStream)
			twitchData.username = tw.Username
			twitchData.Notified = false
			twitchData.Status = false

			chat.TwitchStreams = append(chat.TwitchStreams, *twitchData)
		}

		// Restreamer streams for each chat
		for _, rs := range s.Streams.Restreamer {
			restreamerData := new(RestreamerStream)
			restreamerData.Stream = restreamer.Stream{
				BaseURL: rs.BaseURL,
				ID:      rs.ID,
			}
			restreamerData.Notified = false
			restreamerData.Status = false
			chat.RestreamerStreams = append(chat.RestreamerStreams, *restreamerData)
		}

		streams = append(streams, *chat)
	}
}

func Notify() {
	logger.Log.Debug().Msg("Started notify iteration")
	for i := range streams {
		for j := range streams[i].TwitchStreams {
			checkAndNotifyTwitch(&streams[i].TwitchStreams[j], streams[i].ChatID)
		}
		for j := range streams[i].RestreamerStreams {
			checkAndNotifyRestreamer(&streams[i].RestreamerStreams[j], streams[i].ChatID)
		}
	}

}

func checkAndNotifyTwitch(streamToCheck *TwitchStream, chatID int64) {
	logger.Log.Info().Str("Channel", streamToCheck.username).Msg("Twitch: Checking status")

	// Get stream status and metadata from API
	streamResponse := twitch.GetStreams([]string{streamToCheck.username})
	// List populated means stream is online
	if len(streamResponse) > 0 {
		if !streamToCheck.Status {
			logger.Log.Debug().Str("Channel", streamToCheck.username).Msg("Online and status has changed.")
			streamToCheck.Status = true
			if !streamToCheck.Notified {
				logger.Log.Debug().Str("Channel", streamToCheck.username).Msg("Status change and notification needed.")
				streamToCheck.Notified = true
				telegram.SendTwitchStreamInfo(chatID, streamResponse[0])
			}
		} else {
			logger.Log.Debug().Str("Channel", streamToCheck.username).Msg("Online and status has not changed.")
		}
	} else {
		logger.Log.Debug().Str("Channel", streamToCheck.username).Msg("Channel offline.")
		streamToCheck.Status = false
		streamToCheck.Notified = false
	}
}

func checkAndNotifyRestreamer(streamToCheck *RestreamerStream, chatID int64) {
	logger.Log.Info().Str("Channel", streamToCheck.Stream.ID).Msg("Restreamer: Checking status")
	online := restreamer.CheckStreamLive(streamToCheck.Stream)
	if online {
		streamInfo, err := restreamer.GetStreamInfo(streamToCheck.Stream)
		if err != nil {
			streamToCheck.Notified = false
			streamToCheck.Status = false
			logger.Log.Error().Err(err).Msg("Restreamer appears to be online, could not fetch info")
		}
		if !streamToCheck.Status {
			logger.Log.Debug().Str("Channel", streamInfo.UserName).Msg("Online and status has changed.")
			streamToCheck.Status = true
			if !streamToCheck.Notified {
				logger.Log.Debug().Str("Channel", streamInfo.UserName).Msg("Status change and notification needed.")
				streamToCheck.Notified = true
				telegram.SendRestreamerStreamInfo(chatID, streamInfo, streamToCheck.Stream)
			}
		} else {
			logger.Log.Debug().Str("Channel", streamInfo.UserName).Msg("Online and status has not changed.")
		}
	} else {
		logger.Log.Debug().Str("Channel", streamToCheck.Stream.ID).Msg("Channel offline.")
		streamToCheck.Status = false
		streamToCheck.Notified = false
	}
}
