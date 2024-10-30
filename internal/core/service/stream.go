package service

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"streamobserver/internal/core/domain"
	"streamobserver/internal/core/port"
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

	infos := make([]domain.StreamInfo, 0)
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

	if len(twitchStreams) > 0 {
		twitchInfos, err := ss.twitchGetter.GetStreamInfos(ctx, twitchStreams)
		if err != nil {
			log.Err(err).Msg("unable to fetch twitch streams")
			return nil, err
		}
		infos = append(infos, twitchInfos...)
	}

	if len(restreamerStreams) > 0 {
		restreamerInfos, err := ss.restreamerGetter.GetStreamInfos(ctx, restreamerStreams)
		if err != nil {
			log.Err(err).Msg("unable to fetch restream streams")
			return nil, err
		}
		infos = append(infos, restreamerInfos...)
	}

	return infos, nil
}
