package exec

import (
	"os/exec"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestRunCommand(t *testing.T) {
	hook := test.NewGlobal()
	log.SetLevel(log.DebugLevel)
	defer log.SetLevel(log.InfoLevel)

	message, err := RunCommand("echo", "hello world")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", message)

	assert.Len(t, hook.Entries, 2)
	assert.Equal(t, "echo hello world", hook.Entries[0].Message)
	assert.Equal(t, log.InfoLevel, hook.Entries[0].Level)
	assert.Equal(t, "hello world\n", hook.Entries[1].Message)
	assert.Equal(t, log.DebugLevel, hook.Entries[1].Level)
}

func TestTrimmedOutput(t *testing.T) {
	message, err := RunCommand("printf", "hello world")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", message)
}

func TestRunCommandErr(t *testing.T) {
	hook := test.NewGlobal()
	log.SetLevel(log.DebugLevel)
	defer log.SetLevel(log.InfoLevel)

	_, err := RunCommand("ls", "non-existent")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No such file or directory")

	assert.Len(t, hook.Entries, 2)
	assert.Equal(t, "ls non-existent", hook.Entries[0].Message)
	assert.Equal(t, log.InfoLevel, hook.Entries[0].Level)
	assert.Equal(t, "`ls non-existent` failed: ls: non-existent: No such file or directory\n", hook.Entries[1].Message)
	assert.Equal(t, log.ErrorLevel, hook.Entries[1].Level)
}

func TestRunInDir(t *testing.T) {
	cmd := exec.Command("pwd")
	cmd.Dir = "/"
	message, err := RunCommandExt(cmd)
	assert.Nil(t, err)
	assert.Equal(t, "/", message)
}
