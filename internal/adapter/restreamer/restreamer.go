package restreamer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"streamobserver/internal/core/domain"
	"streamobserver/internal/core/port"
	"sync"

	"github.com/rs/zerolog/log"
)

const (
	channelPath    = "/channels/"
	internalPath   = "/memfs/"
	embedSuffix    = "/oembed.json"
	playlistSuffix = ".m3u8"
	channelSuffix  = ".html"
)

type StreamInfoProvider struct{}

var _ = (*port.StreamInfoProvider)(nil)

type restreamerResponse struct {
	Description  string `json:"description"`
	Username     string `json:"author_name"`
	ThumbnailURL string `json:"thumbnail_url"`
}

func checkOnline(ctx context.Context, stream domain.StreamQuery, client *http.Client) (bool, error) {
	url := stream.BaseURL + internalPath + stream.UserID + playlistSuffix
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("error building http request for restreamer: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("get stream online request failed: %w", err)
	}

	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

func fetchInfo(ctx context.Context, stream domain.StreamQuery, client *http.Client) (restreamerResponse, error) {
	url := stream.BaseURL + channelPath + stream.UserID + embedSuffix
	log.Debug().Str("URL", url).Msg("getting restreamer stream config from URL")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return restreamerResponse{}, fmt.Errorf("error building http request for restreamer: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return restreamerResponse{}, fmt.Errorf("get stream info request failed: %w", err)
	}

	defer resp.Body.Close()

	var restreamerInfo restreamerResponse
	err = json.NewDecoder(resp.Body).Decode(&restreamerInfo)
	if err != nil {
		return restreamerResponse{}, fmt.Errorf("error decoding restreamer stream info: %w", err)
	}

	return restreamerInfo, nil
}

func (s *StreamInfoProvider) GetStreamInfos(ctx context.Context,
	streams []*domain.StreamQuery,
	wg *sync.WaitGroup,
	infos chan<- []domain.StreamInfo,
	errorCh chan<- error) {
	defer wg.Done()

	log.Info().Int("count", len(streams)).Msg("getting info for restreamer streams")

	wg2 := new(sync.WaitGroup)
	wg2.Add(len(streams))

	streamInfos := make([]domain.StreamInfo, 0)
	infoCh := make(chan domain.StreamInfo, len(streams))
	errCh := make(chan error, len(streams))

	client := &http.Client{}

	for _, stream := range streams {
		go fetch(ctx, stream, client, infoCh, errCh, wg2)
	}

	wg2.Wait()
	close(infoCh)
	close(errCh)

	for err := range errCh {
		if err != nil {
			errorCh <- fmt.Errorf("error getting restreamer stream info: %w", err)
			return
		}
	}

	for info := range infoCh {
		streamInfos = append(streamInfos, info)
	}

	infos <- streamInfos
}

func fetch(ctx context.Context,
	query *domain.StreamQuery,
	client *http.Client,
	stream chan<- domain.StreamInfo,
	errs chan<- error,
	wg *sync.WaitGroup) {
	defer wg.Done()

	online, err := checkOnline(ctx, *query, client)
	if err != nil {
		errs <- fmt.Errorf("error checking if stream %s is online: %w", query.UserID, err)
		return
	}

	if !online {
		stream <- domain.StreamInfo{
			Query:    query,
			IsOnline: false,
		}
		return
	}

	info, err := fetchInfo(ctx, *query, client)
	if err != nil {
		errs <- fmt.Errorf("error fetching stream %s info: %w", query.UserID, err)
		return
	}

	var url string
	if query.CustomURL == "" {
		url = query.BaseURL + query.UserID + channelSuffix
	} else {
		url = query.CustomURL
	}

	stream <- domain.StreamInfo{
		Query:        query,
		Username:     info.Username,
		Title:        info.Description,
		URL:          url,
		ThumbnailURL: info.ThumbnailURL,
		IsOnline:     true,
	}
}

func (s *StreamInfoProvider) Kind() domain.StreamKind {
	return domain.StreamKindRestreamer
}
