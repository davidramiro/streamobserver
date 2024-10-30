package restreamer

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"net/http"
	"streamobserver/internal/core/domain"
	"streamobserver/internal/core/port"
	"sync"
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
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Error().Err(err).Msg("error building http request for restreamer")
		return false, err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("get stream online request failed")
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}

func fetchInfo(ctx context.Context, stream domain.StreamQuery, client *http.Client) (restreamerResponse, error) {
	url := stream.BaseURL + channelPath + stream.UserID + embedSuffix
	log.Debug().Str("URL", url).Msg("getting restreamer stream config from URL")

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Error().Err(err).Msg("error building http request for restreamer")
		return restreamerResponse{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("get stream info request failed")
		return restreamerResponse{}, err
	}

	defer resp.Body.Close()

	var restreamerInfo restreamerResponse
	err = json.NewDecoder(resp.Body).Decode(&restreamerInfo)
	if err != nil {
		log.Error().Err(err).Msg("error decoding restreamer stream info")
		return restreamerResponse{}, err
	}

	return restreamerInfo, nil
}

func (s *StreamInfoProvider) GetStreamInfos(ctx context.Context, streams []*domain.StreamQuery) ([]domain.StreamInfo, error) {
	log.Info().Int("count", len(streams)).Msg("getting info for restreamer streams")

	wg := sync.WaitGroup{}

	wg.Add(len(streams))

	infos := make([]domain.StreamInfo, 0)
	infoCh := make(chan domain.StreamInfo, len(streams))
	errCh := make(chan error, len(streams))

	client := &http.Client{}

	for _, stream := range streams {
		go fetch(ctx, stream, client, infoCh, errCh, &wg)
	}

	wg.Wait()
	close(infoCh)
	close(errCh)

	for err := range errCh {
		if err != nil {
			log.Error().Err(err).Msg("error getting stream info")
			return nil, err
		}
	}

	for info := range infoCh {
		infos = append(infos, info)
	}

	return infos, nil
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
		log.Error().Err(err).Msgf("error checking if stream %s is online", query.UserID)
		errs <- err
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
		log.Error().Err(err).Msgf("error fetching stream %s info", query.UserID)
		errs <- err
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
