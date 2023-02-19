package util

import (
	"errors"
	"io"
	"net/http"
	"streamobserver/internal/logger"
	"strings"
)

const (
	widthPlaceholder  = "{width}"
	heightPlaceholder = "{height}"
)

func GetPhotoFromUrl(url string) ([]byte, error) {
	logger.Log.Debug().Msg("Getting image from URL")
	response, e := http.Get(url)
	if e != nil {
		logger.Log.Error().Err(e)
		return nil, errors.New("could not fetch image from URL")
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("could not fetch image from URL")
	}

	if !strings.Contains(response.Header.Get("content-type"), "image") {
		return nil, errors.New("invalid image response")
	}

	defer response.Body.Close()

	imageBytes, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Log.Error().Err(err)
		return nil, errors.New("could not read image bytes")
	}

	return imageBytes, nil
}

func FormatTwitchPhotoUrl(url *string) {
	*url = strings.Replace(*url, heightPlaceholder, "1080", 1)
	*url = strings.Replace(*url, widthPlaceholder, "1920", 1)
}
