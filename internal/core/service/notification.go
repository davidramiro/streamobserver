package service

import (
	"context"
	"streamobserver/internal/core/domain"
	"streamobserver/internal/core/port"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type NotificationService struct {
	notifier port.Notifier
	// TODO: combine stream getters into service agnostic interface
	streamGetter port.StreamInfoService
	streams      map[*domain.StreamQuery]domain.ObservedStream
}

var _ port.NotificationBroker = (*NotificationService)(nil)

func NewNotificationService(n port.Notifier, m port.StreamInfoService) *NotificationService {
	return &NotificationService{
		notifier:     n,
		streamGetter: m,
		streams:      make(map[*domain.StreamQuery]domain.ObservedStream),
	}
}

func (n *NotificationService) Register(target int64, query *domain.StreamQuery) {
	log.Info().Str("id", query.UserID).Int64("target", target).Msg("registering stream")

	queryFound := false
	for k, v := range n.streams {
		if k.Equals(*query) {
			queryFound = true
			targetFound := false
			for _, observer := range v.Observers {
				if observer.ChannelID == target {
					targetFound = true
				}
			}
			if !targetFound {
				v.Observers = append(v.Observers, domain.Observer{
					ChannelID: target,
				})
				n.streams[k] = v
			}
		}
	}

	if !queryFound {
		n.streams[query] = domain.ObservedStream{
			Observers: []domain.Observer{
				{
					ChannelID: target,
				},
			},
		}
	}

	log.Debug().Int("totalObserved", len(n.streams)).Msg("register successful")
}

func (n *NotificationService) StartPolling(ctx context.Context) {
	log.Debug().Msg("starting poll routine")

	for range time.Tick(viper.GetDuration("general.polling_interval")) {
		log.Debug().Msg("tick, querying streams")

		queries := make([]*domain.StreamQuery, 0)
		for k := range n.streams {
			log.Debug().Str("id", k.UserID).Msg("adding stream id to query list")
			queries = append(queries, k)
		}

		infos, err := n.streamGetter.GetStreamInfos(ctx, queries)
		if err != nil {
			log.Err(err).Msg("failed to get stream infos")
			continue
		}

		for _, info := range infos {
			s := n.streams[info.Query]

			log.Debug().Str("id", info.Query.UserID).Msg("checking if notification is needed")

			if !s.LatestInfo.Equals(info) {
				if !info.IsOnline && s.PublishedOfflineStatus {
					continue
				}
				if info.IsOnline {
					s.PublishedOfflineStatus = false
				}

				// updated info on offline streams does not contain metadata, fill and send once, clear message ID
				if !info.IsOnline && !s.PublishedOfflineStatus {
					info = s.LatestInfo
					info.IsOnline = false
					s.PublishedOfflineStatus = true
				}

				log.Info().
					Str("stream", info.Username).
					Bool("online", info.IsOnline).
					Msg("stream status update, notifying")
				go n.notify(ctx, &s, info)

				s.LatestInfo = info
				n.streams[info.Query] = s
			}
		}
	}
}

func (n *NotificationService) notify(ctx context.Context, observed *domain.ObservedStream, info domain.StreamInfo) {
	for i, observer := range observed.Observers {
		log.Info().Int64("target", observer.ChannelID).Str("stream", info.Username).Msg("notifying observer")
		if observer.MessageID == 0 {
			log.Debug().Int64("observer", observer.ChannelID).Msg("first trigger, sending info")
			id, err := n.notifier.SendStreamInfo(ctx, observer.ChannelID, info)
			if err != nil {
				log.Err(err).Int64("observer", observer.ChannelID).Msg("failed to send info")
			}
			observed.Observers[i].MessageID = id
		} else {
			log.Debug().Int64("observer", observer.ChannelID).Msg("later trigger, updating info")
			err := n.notifier.UpdateStreamInfo(ctx, observer.ChannelID, observer.MessageID, info)
			if err != nil {
				log.Err(err).Int64("observer", observer.ChannelID).Msg("failed to update info")
			}
		}
	}

	if observed.PublishedOfflineStatus {
		for i := range observed.Observers {
			observed.Observers[i].MessageID = 0
		}
	}
}
