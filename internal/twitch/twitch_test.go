package twitch

import (
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	// make sure we're in project root for tests
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestReadConfig(t *testing.T) {
	readConfig()

	assert.NotNil(t, configuration)
	assert.NotEmpty(t, configuration.ClientID)
}

func TestAuthenticate(t *testing.T) {
	authenticate()

	assert.NotNil(t, token)
	assert.NotEmpty(t, token.AccessToken)
}

func TestAuthenticateWithValidToken(t *testing.T) {

	token = &authToken{
		tokenCreation: time.Now(),
		ExpiresIn:     1500,
		AccessToken:   "foo",
	}
	authenticate()

	assert.Equal(t, "foo", token.AccessToken)
}
