package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatTwitchPhotoUrl(t *testing.T) {

	result := "test-{width}x{height}.jpg"
	FormatTwitchPhotoUrl(&result)
	expected := "test-1920x1080.jpg"

	assert.Equal(t, expected, result, "dimensions should be present")
}
