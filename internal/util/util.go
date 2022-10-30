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

func ReplaceMarkdownCaption(message *string) {
	*message = strings.ReplaceAll(*message, "[", "[[")
	*message = strings.ReplaceAll(*message, "]", "]]")
}

func GetPhotoFromUrl(url string) ([]byte, error) {
	logger.Log.Debug().Msg("Getting image from URL")
	response, e := http.Get(url)
	if e != nil {
		logger.Log.Fatal().Err(e)
		return nil, errors.New("could not fetch image from URL")
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("could not fetch image from URL")
	}

	defer response.Body.Close()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Log.Fatal().Err(err)
		return nil, errors.New("could not read image bytes")
	}

	return bytes, nil
}

func FormatTwitchPhotoUrl(url *string) {
	*url = strings.Replace(*url, heightPlaceholder, "1080", 1)
	*url = strings.Replace(*url, widthPlaceholder, "1920", 1)
}
