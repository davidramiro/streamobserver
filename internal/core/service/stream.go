package service

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"streamobserver/internal/core/domain"
	"streamobserver/internal/core/port"
	"sync"
)

type StreamService struct {
	twitchGetter     port.StreamInfoProvider
	restreamerGetter port.StreamInfoProvider
}

var _ port.StreamInfoService = (*StreamService)(nil)

func NewStreamService(getters ...port.StreamInfoProvider) *StreamService {
	srv := &StreamService{}

	for _, getter := range getters {
		if getter.Kind() == domain.StreamKindTwitch {
			srv.twitchGetter = getter
		} else if getter.Kind() == domain.StreamKindRestreamer {
			srv.restreamerGetter = getter
		}
	}

	return srv
}

func (ss *StreamService) GetStreamInfos(ctx context.Context, streams []*domain.StreamQuery) ([]domain.StreamInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, viper.GetDuration("general.request_timeout"))
	defer cancel()

	twitchStreams := make([]*domain.StreamQuery, 0)
	restreamerStreams := make([]*domain.StreamQuery, 0)

	for _, stream := range streams {
		if stream.Kind == domain.StreamKindTwitch {
			twitchStreams = append(twitchStreams, stream)
		} else if stream.Kind == domain.StreamKindRestreamer {
			restreamerStreams = append(restreamerStreams, stream)
		}
	}

	log.Info().
		Int("twitchStreamCount", len(twitchStreams)).
		Int("restreamerStreamCount", len(restreamerStreams)).
		Msg("getting stream infos")

	wg := new(sync.WaitGroup)

	infoCh := make(chan []domain.StreamInfo, 2)
	errCh := make(chan error, 2)

	if len(twitchStreams) > 0 {
		wg.Add(1)
		go ss.twitchGetter.GetStreamInfos(ctx, twitchStreams, wg, infoCh, errCh)
	}

	if len(restreamerStreams) > 0 {
		wg.Add(1)
		go ss.restreamerGetter.GetStreamInfos(ctx, restreamerStreams, wg, infoCh, errCh)
	}

	wg.Wait()
	close(errCh)
	close(infoCh)

	for err := range errCh {
		if err != nil {
			log.Error().Err(err).Msg("error getting stream info")
			return nil, err
		}
	}

	infos := make([]domain.StreamInfo, 0)

	for info := range infoCh {
		infos = append(infos, info...)
	}

	return infos, nil
}
