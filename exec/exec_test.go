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

	message, err := RunCommand("echo", CmdOpts{}, "hello world")
	assert.NoError(t, err)
	assert.Equal(t, "hello world\n", message)

	assert.Len(t, hook.Entries, 2)

	assert.Equal(t, log.InfoLevel, hook.Entries[0].Level)
	assert.Equal(t, "echo hello world", hook.Entries[0].Message)
	assert.Equal(t, log.Fields{"dir": ""}, hook.Entries[0].Data)

	assert.Equal(t, log.DebugLevel, hook.Entries[1].Level)
	assert.Equal(t, "hello world\n", hook.Entries[1].Message)
	assert.Nil(t, hook.Entries[1].Data["err"])
	assert.NotNil(t, hook.Entries[1].Data["duration"])
}

func TestTrimmedOutput(t *testing.T) {
	message, err := RunCommand("printf", CmdOpts{}, "hello world")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", message)
}

func TestRunCommandErr(t *testing.T) {
	hook := test.NewGlobal()
	log.SetLevel(log.DebugLevel)
	defer log.SetLevel(log.InfoLevel)

	_, err := RunCommand("ls", CmdOpts{}, "non-existent")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No such file or directory")

	assert.Len(t, hook.Entries, 3)

	assert.Equal(t, log.InfoLevel, hook.Entries[0].Level)
	assert.Equal(t, "ls non-existent", hook.Entries[0].Message)
	assert.Equal(t, log.Fields{"dir": ""}, hook.Entries[0].Data)

	assert.Equal(t, log.DebugLevel, hook.Entries[1].Level)
	assert.Equal(t, "", hook.Entries[1].Message)
	assert.Error(t, hook.Entries[1].Data["err"].(error))
	assert.NotNil(t, hook.Entries[1].Data["duration"])

	assert.Equal(t, log.ErrorLevel, hook.Entries[2].Level)
	assert.Equal(t, "`ls non-existent` failed: ls: non-existent: No such file or directory\n", hook.Entries[2].Message)
}

func TestRunInDir(t *testing.T) {
	cmd := exec.Command("pwd")
	cmd.Dir = "/"
	message, err := RunCommandExt(cmd, CmdOpts{})
	assert.Nil(t, err)
	assert.Equal(t, "/", message)
}
