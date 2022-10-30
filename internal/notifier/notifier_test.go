package notifier

import (
	"os"
	"path"
	"runtime"
	"testing"

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

func TestPopulateObservers(t *testing.T) {
	PopulateObservers()
	assert.NotEmpty(t, chats, "chat slice should not be empty after reading config")
}
