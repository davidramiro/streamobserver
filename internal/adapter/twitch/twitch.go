package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"net/url"
	"streamobserver/internal/core/domain"
	"streamobserver/internal/core/port"
	"strings"
	"sync"
	"time"
)

const (
	widthPlaceholder  = "{width}"
	heightPlaceholder = "{height}"
	twitchTokenURL    = "https://id.twitch.tv/oauth2/token"
	twitchStreamsURL  = "https://api.twitch.tv/helix/streams"
	twitchBaseURL     = "https://twitch.tv"
	twitchMimeType    = "application/json"
)

type StreamInfoProvider struct {
	token authToken
}

var _ = (*port.StreamInfoProvider)(nil)

type twitchResponse struct {
	Data []struct {
		Username     string `json:"user_login"`
		GameName     string `json:"game_name"`
		Title        string `json:"title"`
		ThumbnailURL string `json:"thumbnail_url"`
	} `json:"data,omitempty"`
}

type authToken struct {
	AccessToken   string `json:"access_token"`
	ExpiresIn     int    `json:"expires_in"`
	tokenCreation time.Time
}

func formatTwitchPhotoUrl(url string) string {
	url = strings.Replace(url, heightPlaceholder, "1080", 1)
	return strings.Replace(url, widthPlaceholder, "1920", 1)
}

func (s *StreamInfoProvider) GetStreamInfos(ctx context.Context,
	streams []*domain.StreamQuery,
	wg *sync.WaitGroup,
	infos chan<- []domain.StreamInfo,
	errCh chan<- error) {
	defer wg.Done()

	log.Info().Int("count", len(streams)).Msg("getting info for twitch streams")

	err := s.authenticate(ctx)
	if err != nil {
		log.Err(err).Msg("error authenticating with twitch")
		errCh <- err
		return
	}

	bearer := "Bearer " + s.token.AccessToken

	base, err := url.Parse(twitchStreamsURL)
	if err != nil {
		errCh <- err
		return
	}

	// Query params
	params := url.Values{}
	params.Add("type", "all")
	params.Add("first", "100")
	for _, s := range streams {
		params.Add("user_login", s.UserID)
	}
	base.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		log.Err(err).Msg("error building request for twitch")
		errCh <- err
		return
	}

	req.Header.Set("Authorization", bearer)
	req.Header.Add("Accept", twitchMimeType)
	req.Header.Add("client-id", viper.GetString("twitch.client_id"))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("error making request to twitch")
		errCh <- err
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Error().Int("StatusCode", resp.StatusCode).Interface("Response", resp).
			Msg("No HTTP OK from Twitch Helix.")
		errCh <- err
		return
	}

	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Err(err).Msg("error getting bytes from twitch json response")
		errCh <- err
		return
	}

	buffer := new(bytes.Buffer)
	err = json.Compact(buffer, responseBytes)
	if err != nil {
		log.Err(err).Msg("error compacting twitch json response")
		errCh <- err
		return
	}

	var response twitchResponse
	err = json.NewDecoder(buffer).Decode(&response)
	if err != nil {
		log.Err(err).Msg("error decoding response from twitch")
		errCh <- err
		return
	}

	streamInfos := make([]domain.StreamInfo, len(streams))

	for i, s := range streams {
		online := false
		log.Debug().Str("id", s.UserID).Msg("checking if stream in response")
		for _, data := range response.Data {
			if strings.EqualFold(data.Username, s.UserID) {
				log.Debug().Msg("found, setting info")
				streamInfos[i] = domain.StreamInfo{
					Query:        s,
					Username:     data.Username,
					Title:        fmt.Sprintf("%s: %s", data.GameName, data.Title),
					URL:          fmt.Sprintf("%s/%s", twitchBaseURL, data.Username),
					ThumbnailURL: formatTwitchPhotoUrl(data.ThumbnailURL),
					IsOnline:     true,
				}
				online = true
			}
		}
		if !online {
			log.Debug().Msg("not found, setting offline info")
			streamInfos[i] = domain.StreamInfo{
				Query:    s,
				Username: s.UserID,
				IsOnline: false,
			}
		}
	}

	infos <- streamInfos
}

func (s *StreamInfoProvider) authenticate(ctx context.Context) error {
	log.Debug().Msg("authenticating with twitch API")
	if s.token.AccessToken != "" {
		log.Debug().Msg("twitch auth token present, checking validity")
		expiryTime := s.token.tokenCreation.Add(time.Second * time.Duration(s.token.ExpiresIn))
		if time.Now().Before(expiryTime) {
			log.Debug().Msg("token still valid.")
			return nil
		}
	}

	base, err := url.Parse(twitchTokenURL)
	if err != nil {
		return err
	}

	// Query params
	params := url.Values{}
	params.Add("client_id", viper.GetString("twitch.client_id"))
	params.Add("client_secret", viper.GetString("twitch.client_secret"))
	params.Add("grant_type", "client_credentials")
	base.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base.String(), nil)
	if err != nil {
		log.Err(err).Msg("error creating twitch auth request")
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Err(err).Msg("error executing twitch auth request")
		return err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&s.token)
	s.token.tokenCreation = time.Now()
	if err != nil {
		log.Err(err).Msg("error parsing twitch auth response")
		return err
	}

	log.Debug().Msg("successfully authenticated on twitch")
	return nil
}

func (s *StreamInfoProvider) Kind() domain.StreamKind {
	return domain.StreamKindTwitch
}
