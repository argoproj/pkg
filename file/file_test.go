package file

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/argoproj/pkg/rand"
)

// TestIsDirectory tests if a path is a directory. Errors if directory doesn't exist
func TestIsDirectory(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("could not determine test directory")
	}
	testDir := filepath.Dir(filename)

	isDir, err := IsDirectory(testDir)
	assert.Nil(t, err)
	assert.True(t, isDir)

	isDir, err = IsDirectory(filename)
	assert.Nil(t, err)
	assert.False(t, isDir)

	isDir, err = IsDirectory("doesnt-exist")
	assert.NotNil(t, err)
	assert.False(t, isDir)
}

func TestExists(t *testing.T) {
	assert.True(t, Exists("/"))
	randFilePath := fmt.Sprintf("/%s", rand.RandString(10))
	assert.False(t, Exists(randFilePath))
}
