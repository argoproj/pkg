package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRunCommand tests the RunCommand function
func TestRunCommand(t *testing.T) {
	message, err := RunCommand("echo", "hello world")
	assert.Nil(t, err)
	assert.Equal(t, "hello world", message)

	message, err = RunCommand("printf", "hello world")
	assert.Nil(t, err)
	assert.Equal(t, "hello world", message)

	message, err = RunCommand("ls", "non-existant-dir")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No such file or directory")
}
