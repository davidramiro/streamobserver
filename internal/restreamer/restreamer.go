package restreamer

import (
	"encoding/json"
	"net/http"
	"streamobserver/internal/logger"
)

// StreamInfo contains Restreamer stream metadata
type StreamInfo struct {
	Description  string `json:"description"`
	UserName     string `json:"author_name"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// Stream contains the basic info used to poll a stream from Restreamer
type Stream struct {
	BaseURL   string
	ID        string
	CustomURL string
}

const channelPath = "/channels/"
const internalPath = "/memfs/"
const embedSuffix = "/oembed.json"
const playlistSuffix = ".m3u8"

// CheckStreamLive returns if a Restreamer stream is online
func CheckStreamLive(stream Stream) (bool, error) {
	url := stream.BaseURL + internalPath + stream.ID + playlistSuffix
	logger.Log.Debug().Str("URL", url).Msg("Restreamer: Checking URL for HTTP OK")
	resp, err := http.Get(url)
	if err != nil {
		logger.Log.Fatal().Err(err)
		return false, err
	}

	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

// GetStreamInfo returns the metadata of a stream, if online
func GetStreamInfo(stream Stream) (StreamInfo, error) {

	url := stream.BaseURL + channelPath + stream.ID + embedSuffix
	logger.Log.Debug().Str("URL", url).Msg("Restreamer: Getting stream config from URL")
	resp, err := http.Get(url)
	if err != nil {
		logger.Log.Fatal().Err(err)
		return StreamInfo{}, err
	}

	defer resp.Body.Close()
	var streamInfo StreamInfo
	err = json.NewDecoder(resp.Body).Decode(&streamInfo)
	if err != nil {
		logger.Log.Fatal().Err(err)
		return StreamInfo{}, err
	}
	logger.Log.Debug().Interface("Info", streamInfo).Msg("Received Stream Info")
	return streamInfo, nil
}
