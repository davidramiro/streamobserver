package port

import (
	"context"
	"streamobserver/internal/core/domain"
)

type StreamInfoProvider interface {
	// GetStreamInfos takes an array of streams for a single stream service and returns metadata for those that are online.
	GetStreamInfos(ctx context.Context, streams []*domain.StreamQuery) ([]domain.StreamInfo, error)
	// Kind returns the streaming service fetched by this provider
	Kind() domain.StreamKind
}

type Notifier interface {
	// SendStreamInfo sends a message with stream info to a target channel ID
	SendStreamInfo(ctx context.Context, target int64, stream domain.StreamInfo) (messageID int, err error)
	// UpdateStreamInfo updates a previously sent message ID with stream info
	UpdateStreamInfo(ctx context.Context, chatID int64, messageID int, stream domain.StreamInfo) error
}

type NotificationBroker interface {
	// Register adds a target channel ID and a stream to observe NotificationBroker
	Register(target int64, query *domain.StreamQuery)
	// StartPolling starts the notification routine
	StartPolling(ctx context.Context)
}

type StreamInfoService interface {
	// GetStreamInfos retrieves stream info for different providers
	GetStreamInfos(ctx context.Context, streams []*domain.StreamQuery) ([]domain.StreamInfo, error)
}
