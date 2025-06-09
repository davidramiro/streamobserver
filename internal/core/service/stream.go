package service

import (
	"context"
	"streamobserver/internal/core/domain"
	"streamobserver/internal/core/port"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type StreamService struct {
	twitchGetter       port.StreamInfoProvider
	restreamerGetter   port.StreamInfoProvider
	broadcastboxGetter port.StreamInfoProvider
}

var _ port.StreamInfoService = (*StreamService)(nil)

func NewStreamService(getters ...port.StreamInfoProvider) *StreamService {
	srv := &StreamService{}

	for _, getter := range getters {
		switch getter.Kind() {
		case domain.StreamKindTwitch:
			srv.twitchGetter = getter
		case domain.StreamKindRestreamer:
			srv.restreamerGetter = getter
		case domain.StreamKindBroadcastBox:
			srv.broadcastboxGetter = getter
		}
	}

	return srv
}

const channelCount = 3

func (ss *StreamService) GetStreamInfos(
	ctx context.Context,
	streams []*domain.StreamQuery,
) ([]domain.StreamInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, viper.GetDuration("general.request_timeout"))
	defer cancel()

	twitchStreams := make([]*domain.StreamQuery, 0)
	restreamerStreams := make([]*domain.StreamQuery, 0)
	broadcastboxStreams := make([]*domain.StreamQuery, 0)

	for _, stream := range streams {
		switch stream.Kind {
		case domain.StreamKindTwitch:
			twitchStreams = append(twitchStreams, stream)
		case domain.StreamKindRestreamer:
			restreamerStreams = append(restreamerStreams, stream)
		case domain.StreamKindBroadcastBox:
			broadcastboxStreams = append(broadcastboxStreams, stream)
		}
	}

	log.Info().
		Int("twitchStreamCount", len(twitchStreams)).
		Int("restreamerStreamCount", len(restreamerStreams)).
		Int("broadcastboxStreamCount", len(broadcastboxStreams)).
		Msg("getting stream infos")

	wg := new(sync.WaitGroup)

	infoCh := make(chan []domain.StreamInfo, channelCount)
	errCh := make(chan error, channelCount)

	if len(twitchStreams) > 0 {
		wg.Add(1)
		go ss.twitchGetter.GetStreamInfos(ctx, twitchStreams, wg, infoCh, errCh)
	}

	if len(restreamerStreams) > 0 {
		wg.Add(1)
		go ss.restreamerGetter.GetStreamInfos(ctx, restreamerStreams, wg, infoCh, errCh)
	}

	if len(broadcastboxStreams) > 0 {
		wg.Add(1)
		go ss.broadcastboxGetter.GetStreamInfos(ctx, broadcastboxStreams, wg, infoCh, errCh)
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
