package domain

type StreamQuery struct {
	UserID    string
	BaseURL   string
	CustomURL string
	Kind      StreamKind
}

func (s StreamQuery) Equals(o StreamQuery) bool {
	return s.UserID == o.UserID &&
		s.BaseURL == o.BaseURL &&
		s.CustomURL == o.CustomURL &&
		s.Kind == o.Kind
}

type StreamKind string

const (
	StreamKindTwitch       StreamKind = "twitch"
	StreamKindRestreamer   StreamKind = "restreamer"
	StreamKindBroadcastBox StreamKind = "broadcastbox"
)

type StreamInfo struct {
	Query        *StreamQuery
	Username     string
	Title        string
	URL          string
	ThumbnailURL string
	IsOnline     bool
}

func (s StreamInfo) Equals(o StreamInfo) bool {
	return s.IsOnline == o.IsOnline &&
		s.Title == o.Title &&
		s.URL == o.URL &&
		s.ThumbnailURL == o.ThumbnailURL
}

type Observer struct {
	ChannelID int64
	MessageID int
}

type ObservedStream struct {
	Observers              []Observer
	LatestInfo             StreamInfo
	PublishedOfflineStatus bool
}

type ChatConfig struct {
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
		BroadcastBox []struct {
			BaseURL   string `yaml:"baseurl"`
			ID        string `yaml:"id"`
			CustomURL string `yaml:"customurl"`
		} `yaml:"broadcastbox"`
	} `yaml:"streams"`
}
