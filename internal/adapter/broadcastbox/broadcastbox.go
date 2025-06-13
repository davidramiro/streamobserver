package broadcastbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"streamobserver/internal/core/domain"
	"streamobserver/internal/core/port"
	"sync"

	"github.com/rs/zerolog/log"
)

const (
	apiURL = "/api/status"
)

type StreamInfoProvider struct{}

var _ = (*port.StreamInfoProvider)(nil)

type broadcastboxResponse struct {
	StreamKey    string `json:"streamKey"`
	VideoStreams []struct {
		RID string `json:"rid"`
	} `json:"videoStreams"`
	WhepSessions []struct {
		ID string `json:"id"`
	} `json:"whepSessions"`
}

func checkOnline(ctx context.Context, stream domain.StreamQuery, client *http.Client) (bool, int, error) {
	url := stream.BaseURL + apiURL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, 0, fmt.Errorf("error building http request for broadcastbox: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, 0, fmt.Errorf("get stream online request failed: %w", err)
	}
	defer resp.Body.Close()

	var respBody []broadcastboxResponse
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, 0, fmt.Errorf("error reading response body: %w", err)
	}

	err = json.Unmarshal(b, &respBody)
	if err != nil {
		return false, 0, fmt.Errorf("error decoding response body: %w", err)
	}

	for _, s := range respBody {
		if s.StreamKey == stream.UserID {
			return true, len(s.WhepSessions), nil
		}
	}

	return false, 0, nil
}

func (s *StreamInfoProvider) GetStreamInfos(ctx context.Context,
	streams []*domain.StreamQuery,
	wg *sync.WaitGroup,
	infos chan<- []domain.StreamInfo,
	errorCh chan<- error) {
	defer wg.Done()

	log.Info().Int("count", len(streams)).Msg("getting info for bb streams")

	streamInfos := make([]domain.StreamInfo, 0)

	client := &http.Client{}

	for _, stream := range streams {
		on, viewers, err := checkOnline(ctx, *stream, client)
		if err != nil {
			errorCh <- fmt.Errorf("error checking if stream %s is online: %w", stream.UserID, err)
			return
		}

		if on {
			var url string
			if stream.CustomURL == "" {
				url = stream.BaseURL + "/" + stream.UserID
			} else {
				url = stream.CustomURL
			}

			streamInfos = append(streamInfos, domain.StreamInfo{
				Username:    stream.UserID,
				URL:         url,
				Query:       stream,
				ViewerCount: viewers,
				IsOnline:    true,
			})
		} else {
			streamInfos = append(streamInfos, domain.StreamInfo{
				Query:    stream,
				IsOnline: false,
			})
		}
	}

	infos <- streamInfos
}

func (s *StreamInfoProvider) Kind() domain.StreamKind {
	return domain.StreamKindBroadcastBox
}
