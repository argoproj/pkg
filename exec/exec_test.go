package exec

import (
	"os/exec"
	"testing"
	"time"

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
	assert.Equal(t, "hello world", message)

	assert.Len(t, hook.Entries, 2)

	entry := hook.Entries[0]
	assert.Equal(t, log.InfoLevel, entry.Level)
	assert.Equal(t, "echo hello world", entry.Message)
	assert.Contains(t, entry.Data, "dir")
	assert.Contains(t, entry.Data, "execID")

	entry = hook.Entries[1]
	assert.Equal(t, log.DebugLevel, entry.Level)
	assert.Equal(t, "hello world\n", entry.Message)
	assert.Contains(t, entry.Data, "duration")
	assert.Contains(t, entry.Data, "execID")
}

func TestRunCommandTimeout(t *testing.T) {
	hook := test.NewGlobal()
	log.SetLevel(log.DebugLevel)
	defer log.SetLevel(log.InfoLevel)

	output, err := RunCommand("sh", CmdOpts{timeout: 500 * time.Millisecond}, "-c", "echo my-output && echo my-error >&2 && sleep 2")
	assert.Equal(t, output, "my-output")
	assert.EqualError(t, err, "`sh -c echo my-output && echo my-error >&2 && sleep 2` timeout after 500ms")

	assert.Len(t, hook.Entries, 3)

	entry := hook.Entries[0]
	assert.Equal(t, log.InfoLevel, entry.Level)
	assert.Equal(t, "sh -c echo my-output && echo my-error >&2 && sleep 2", entry.Message)
	assert.Contains(t, entry.Data, "dir")
	assert.Contains(t, entry.Data, "execID")

	entry = hook.Entries[1]
	assert.Equal(t, log.DebugLevel, entry.Level)
	assert.Equal(t, "my-output\n", entry.Message)
	assert.Contains(t, entry.Data, "duration")
	assert.Contains(t, entry.Data, "execID")

	entry = hook.Entries[2]
	assert.Equal(t, log.ErrorLevel, entry.Level)
	assert.Equal(t, "`sh -c echo my-output && echo my-error >&2 && sleep 2` timeout after 500ms", entry.Message)
	assert.Contains(t, entry.Data, "execID")
}

func TestTrimmedOutput(t *testing.T) {
	message, err := RunCommand("printf", CmdOpts{}, "hello world")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", message)
}
func TestRunCommandExitErr(t *testing.T) {
	hook := test.NewGlobal()
	log.SetLevel(log.DebugLevel)
	defer log.SetLevel(log.InfoLevel)

	output, err := RunCommand("sh", CmdOpts{}, "-c", "echo my-output && echo my-error >&2 && exit 1")
	assert.Equal(t, "my-output", output)
	assert.EqualError(t, err, "`sh -c echo my-output && echo my-error >&2 && exit 1` failed: my-error")

	assert.Len(t, hook.Entries, 3)

	entry := hook.Entries[0]
	assert.Equal(t, log.InfoLevel, entry.Level)
	assert.Equal(t, "sh -c echo my-output && echo my-error >&2 && exit 1", entry.Message)
	assert.Contains(t, entry.Data, "dir")
	assert.Contains(t, entry.Data, "execID")

	entry = hook.Entries[1]
	assert.Equal(t, log.DebugLevel, entry.Level)
	assert.Equal(t, "my-output\n", entry.Message)
	assert.Contains(t, entry.Data, "duration")
	assert.Contains(t, entry.Data, "execID")

	entry = hook.Entries[2]
	assert.Equal(t, log.ErrorLevel, entry.Level)
	assert.Equal(t, "`sh -c echo my-output && echo my-error >&2 && exit 1` failed: my-error", entry.Message)
	assert.Contains(t, entry.Data, "execID")
}

func TestRunCommandErr(t *testing.T) {
	//_ := test.NewGlobal()
	log.SetLevel(log.DebugLevel)
	defer log.SetLevel(log.InfoLevel)

	// this would create a panic
	output, err := RunCommand("", CmdOpts{})
	assert.Empty(t, output)
	assert.EqualError(t, err, "fork/exec : no such file or directory")
}

func TestRunInDir(t *testing.T) {
	cmd := exec.Command("pwd")
	cmd.Dir = "/"
	message, err := RunCommandExt(cmd, CmdOpts{})
	assert.Nil(t, err)
	assert.Equal(t, "/", message)
}
