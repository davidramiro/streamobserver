package twitch

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"streamobserver/internal/config"
	"streamobserver/internal/logger"
	"time"

	"github.com/rs/zerolog/log"
)

type Stream struct {
	UserName     string `json:"user_name"`
	GameName     string `json:"game_name"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnail_url"`
}
type pagination struct {
	Cursor string `json:"cursor"`
}

type streamResponse struct {
	Data []Stream `json:"data"`
}

const twitchTokenURL = "https://id.twitch.tv/oauth2/token"
const twitchStreamsURL = "https://api.twitch.tv/helix/streams"
const twitchMimeType = "application/json"

var token = struct {
	AccessToken   string `json:"access_token"`
	ExpiresIn     int    `json:"expires_in"`
	TokenType     string `json:"token_type"`
	tokenCreation time.Time
}{}

// GetStreams takes an array of Twitch usernames and returns metadata for those that are online.
func GetStreams(usernames []string) []Stream {
	authenticate()

	bearer := "Bearer " + token.AccessToken

	base, err := url.Parse(twitchStreamsURL)
	if err != nil {
		return nil
	}

	// Query params
	params := url.Values{}
	for _, s := range usernames {
		params.Add("user_login", s)
	}
	base.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", base.String(), bytes.NewBuffer(nil))
	if err != nil {
		logger.Log.Error().Err(err)
	}

	config, err := config.GetConfig()
	if err != nil {
		log.Panic().Err(err)
	}

	req.Header.Set("Authorization", bearer)
	req.Header.Add("Accept", twitchMimeType)
	req.Header.Add("client-id", config.Twitch.ClientID)

	client := &http.Client{}

	defer req.Body.Close()
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error().Err(err)
	}
	if resp.StatusCode != 200 {
		logger.Log.Error().Int("StatusCode", resp.StatusCode).Interface("Response", resp).Msg("No HTTP OK from Twitch Helix.")
	}

	var streams streamResponse

	err = json.NewDecoder(resp.Body).Decode(&streams)
	if err != nil {
		logger.Log.Fatal().Err(err)
		return nil
	}

	for _, s := range streams.Data {
		logger.Log.Debug().Interface("Info", s).Msg("Received Stream Info")
	}

	return streams.Data
}

func authenticate() {
	if token.ExpiresIn != 0 {
		logger.Log.Debug().Msg("Twitch auth token present. Checking validity.")
		expiryTime := token.tokenCreation.Add(time.Second * time.Duration(token.ExpiresIn))
		if time.Now().Before(expiryTime) {
			logger.Log.Debug().Msg("Token still valid.")
			return
		}
	}

	logger.Log.Debug().Msg("Authenticating with Twitch API")
	config, err := config.GetConfig()
	if err != nil {
		log.Panic().Err(err)
		return
	}

	values := map[string]string{
		"client_id":     config.Twitch.ClientID,
		"client_secret": config.Twitch.ClientSecret,
		"grant_type":    "client_credentials"}
	json_data, err := json.Marshal(values)

	if err != nil {
		logger.Log.Fatal().Err(err)
	}

	resp, err := http.Post(twitchTokenURL, twitchMimeType,
		bytes.NewBuffer(json_data))
	if err != nil {
		logger.Log.Fatal().Err(err)
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&token)
	token.tokenCreation = time.Now()
	if err != nil {
		logger.Log.Fatal().Err(err)
		return
	}
	logger.Log.Info().Msg("Successfully authenticated on Twitch. ")
}
