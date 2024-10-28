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
	ThumbnailURL: "https://via.placeholder.com/960.jpg",
}

var testTwitchStream = &twitch.Stream{
	UserName:     "testchannel",
	GameName:     "testgame",
	ThumbnailURL: "https://via.placeholder.com/{height}.jpg",
	Title:        "[test] (streamobserver) title",
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
	err = InitBot(true)
	if err != nil {
		panic(err)
	}

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

func TestSendTwitchStreamInfo(t *testing.T) {

	result, err := SendTwitchStreamInfo(*testchatid, *testTwitchStream)

	if assert.NoError(t, err, "correctly formatted send request should not throw error") {
		assert.Equal(t, *testchatid, result.Chat.ID, "response from server should match chatID")
		assert.Contains(t, result.Caption, testTwitchStream.Title, "message should contain title")
		assert.Contains(t, result.Caption, testTwitchStream.UserName, "message should contain title")
		assert.Contains(t, result.Caption, testTwitchStream.GameName, "message should contain game name")
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

	result, err = SendUpdateStreamOffline(result, *testchatid)

	expectedString := "❌ OFFLINE"
	if assert.NoError(t, err, "updating message with valid request should not return error") {
		assert.Equal(t, *testchatid, result.Chat.ID, "updated response should contain equal chatID")
		assert.Contains(t, result.Caption, expectedString, "response should contain updated substring")
	}

}

func TestSendUpdateTwitchStreamInfo(t *testing.T) {
	stream := *testTwitchStream
	result, err := SendTwitchStreamInfo(*testchatid, stream)

	if assert.NoError(t, err, "sending message with valid request should not return error") {
		assert.Equal(t, *testchatid, result.Chat.ID, "response should contain equal chatID")
	}

	stream.GameName = "New Game"
	stream.Title = "New Title"

	result, err = SendUpdateTwitchStreamInfo(*testchatid, result, stream)

	expectedStringGame := "New Game"
	expectedStringTitle := "New Title"

	if assert.NoError(t, err, "updating message with valid request should not return error") {
		assert.Equal(t, *testchatid, result.Chat.ID, "updated response should contain equal chatID")
		assert.Contains(t, result.Caption, expectedStringGame, "response should contain updated title string")
		assert.Contains(t, result.Caption, expectedStringTitle, "response should contain updated game string")
	}
}
