package telegram

import (
	"errors"
	"os"
	"path"
	"runtime"
	"streamobserver/internal/config"
	"streamobserver/internal/logger"
	"streamobserver/internal/restreamer"
	"streamobserver/internal/twitch"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testchatid *int64

var testRestreamerInfo = &restreamer.StreamInfo{
	Description:  "testdesc",
	UserName:     "testchannel",
	ThumbnailURL: "https://via.placeholder.com/300.jpg",
}

var testTwitchStream = &twitch.Stream{
	UserName:     "testchannel",
	GameName:     "testgame",
	ThumbnailURL: "https://via.placeholder.com/300.jpg",
	Title:        "testtitle",
}

var testRestreamerStream = &restreamer.Stream{
	BaseURL:   "https://teststream.tld",
	ID:        "testid",
	CustomURL: "testcustomurl",
}

func init() {
	// make sure we're in project root for tests
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
	InitBot(true)

	config, err := config.GetConfig()
	if err != nil {
		logger.Log.Panic().Err(err)
		return
	}

	if config == nil {
		logger.Log.Panic().Err(errors.New("got empty config"))
		return
	}
	testchatid = &config.General.TestChatID
}

func TestFormatTwitchPhotoUrl(t *testing.T) {
	result := formatTwitchPhotoUrl("test-{width}x{height}.jpg")
	expected := "test-1920x1080.jpg"

	assert.Equal(t, expected, result, "dimensions should be present")
}

func TestCreatePhotoMessageNotImage(t *testing.T) {
	// Testing broken URL
	_, err := createPhotoMessage("test", 42, "https://via.placeholder.com/notanimage")
	expectedError := "could not retrieve photo"

	if assert.Error(t, err, "should report error on broken image URL") {
		assert.Equal(t, expectedError, err.Error(), "error message should reflect image retrieval issue")
	}
}

func TestCreatePhotoMessageBrokenUrl(t *testing.T) {
	// Testing 404 URL
	_, err := createPhotoMessage("test", 42, "http://notfound.tld/image.jpg")
	expectedError := "could not retrieve photo"

	if assert.Error(t, err, "should report error on non-image URL") {
		assert.Equal(t, expectedError, err.Error(), "error message should reflect image retrieval issue")
	}
}

func TestCreatePhotoMessageValid(t *testing.T) {
	// Testing created Photo Config
	testcaption := "testcaption"
	testid := int64(42)
	result, err := createPhotoMessage(testcaption, testid, "https://via.placeholder.com/300.jpg")

	if assert.NoError(t, err) {
		assert.Equal(t, testcaption, result.Caption, "config should return expected caption")
		assert.Equal(t, testid, result.ChatID, "config should return expected chatID")
		assert.True(t, result.File.NeedsUpload(), "not yet send message should indicate file upload bool")
	}
}

func TestSendTwitchStreamInfo(t *testing.T) {

	result, err := SendTwitchStreamInfo(*testchatid, *testTwitchStream)

	if assert.NoError(t, err, "correctly formatted send request should not throw error") {
		assert.Equal(t, *testchatid, result.Chat.ID, "response from server should match chatID")
	}
}

func TestSendTwitchStreamWithBadRequest(t *testing.T) {
	teststream := twitch.Stream{}
	_, err := SendTwitchStreamInfo(*testchatid, teststream)

	expectedError := "could not send message"

	if assert.Error(t, err, "empty stream info should return error on send request") {
		assert.Equal(t, expectedError, err.Error(), "request should report back correct error")
	}
}

func TestSendTwitchStreamWithBadChatId(t *testing.T) {
	testchatid := int64(42)

	res, _ := SendTwitchStreamInfo(testchatid, *testTwitchStream)

	assert.Nil(t, res.Chat, "bad chatID should return no chat in response")
}

func TestSendRestreamerStreamInfo(t *testing.T) {

	result, err := SendRestreamerStreamInfo(*testchatid, *testRestreamerInfo, *testRestreamerStream)

	if assert.NoError(t, err, "correctly formatted send request should not throw error") {
		assert.Equal(t, *testchatid, result.Chat.ID, "response from server should match chatID")
	}
}

func TestSendRestreamerStreamWithBadRequest(t *testing.T) {
	testchatid := int64(0)
	teststream := restreamer.Stream{}
	teststreaminfo := restreamer.StreamInfo{}

	_, err := SendRestreamerStreamInfo(testchatid, teststreaminfo, teststream)

	expectedError := "could not send message"

	if assert.Error(t, err, "empty stream info should return error on send request") {
		assert.Equal(t, expectedError, err.Error(), "request should report back correct error")
	}
}

func TestSendRestreamerStreamWithBadChatId(t *testing.T) {
	testchatid := int64(42)

	res, _ := SendRestreamerStreamInfo(testchatid, *testRestreamerInfo, *testRestreamerStream)

	assert.Nil(t, res.Chat, "bad chatID should return no chat in response")
}

func TestUpdateMessageStreamOffline(t *testing.T) {
	result, err := SendRestreamerStreamInfo(*testchatid, *testRestreamerInfo, *testRestreamerStream)

	if assert.NoError(t, err, "sending message with valid request should not return error") {
		assert.Equal(t, *testchatid, result.Chat.ID, "response should contain equal chatID")
	}

	result, err = UpdateMessageStreamOffline(result, *testchatid)

	expectedString := "❌ OFFLINE"
	if assert.NoError(t, err, "updating message with valid request should not return error") {
		assert.Equal(t, *testchatid, result.Chat.ID, "updated response should contain equal chatID")
		assert.Contains(t, result.Caption, expectedString, "response should contain updated substring")
	}

}
