package twitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"streamobserver/internal/config"
	"streamobserver/internal/logger"
	"time"
)

type Stream struct {
	UserName     string `json:"user_name"`
	GameName     string `json:"game_name"`
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnail_url"`
}
type pagination struct {
	Cursor string `json:"cursor"`
}

type streamResponse struct {
	Data []Stream `json:"data"`
}

type authToken struct {
	AccessToken   string `json:"access_token"`
	ExpiresIn     int    `json:"expires_in"`
	TokenType     string `json:"token_type"`
	tokenCreation time.Time
}

const (
	twitchTokenURL   = "https://id.twitch.tv/oauth2/token"
	twitchStreamsURL = "https://api.twitch.tv/helix/streams"
	twitchMimeType   = "application/json"
)

var token *authToken
var configuration *config.Twitch

// GetStreams takes an array of Twitch usernames and returns metadata for those that are online.
func GetStreams(usernames []string) ([]Stream, error) {
	authenticate()

	bearer := "Bearer " + token.AccessToken

	base, err := url.Parse(twitchStreamsURL)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	req.Header.Set("Authorization", bearer)
	req.Header.Add("Accept", twitchMimeType)
	req.Header.Add("client-id", configuration.ClientID)

	client := &http.Client{}

	defer req.Body.Close()
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error().Err(err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		logger.Log.Error().Int("StatusCode", resp.StatusCode).Interface("Response", resp).Msg("No HTTP OK from Twitch Helix.")
		return nil, err
	}

	var streams streamResponse

	err = json.NewDecoder(resp.Body).Decode(&streams)
	if err != nil {
		logger.Log.Fatal().Err(err)
		return nil, err
	}

	for _, s := range streams.Data {
		logger.Log.Debug().Interface("Info", s).Msg("Received Stream Info")
	}

	return streams.Data, nil
}

func authenticate() {

	if configuration == nil {
		readConfig()
	}

	logger.Log.Debug().Msg("Authenticating with Twitch API")
	if token != nil && token.ExpiresIn != 0 {
		logger.Log.Debug().Msg("Twitch auth token present. Checking validity.")
		expiryTime := token.tokenCreation.Add(time.Second * time.Duration(token.ExpiresIn))
		if time.Now().Before(expiryTime) {
			logger.Log.Debug().Msg("Token still valid.")
			return
		}
	}

	values := map[string]string{
		"client_id":     configuration.ClientID,
		"client_secret": configuration.ClientSecret,
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

func readConfig() bool {

	var err error
	config, err := config.GetConfig()
	if err != nil {
		logger.Log.Panic().Err(err)
		return false
	}

	if config == nil {
		logger.Log.Panic().Err(errors.New("got empty config"))
		return false
	}

	configuration = &config.Twitch

	return true
}
