package exec

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var ErrWaitPIDTimeout = fmt.Errorf("Timed out waiting for PID to complete")

// RunCommandExt is a convenience function to run/log a command and return/log stderr in an error upon
// failure. Chops off a trailing newline (if present)
func RunCommandExt(cmd *exec.Cmd) (string, error) {
	cmdStr := strings.Join(cmd.Args, " ")
	log.Info(cmdStr)
	outBytes, err := cmd.Output()
	if err != nil {
		exErr, ok := err.(*exec.ExitError)
		if !ok {
			return "", errors.WithStack(err)
		}
		errOutput := string(exErr.Stderr)
		log.Errorf("`%s` failed: %s", cmdStr, errOutput)
		return "", errors.New(strings.TrimSpace(errOutput))
	}
	// Trims off a single newline for user convenience
	output := string(outBytes)
	outputLen := len(output)
	if outputLen > 0 && output[outputLen-1] == '\n' {
		output = output[0 : outputLen-1]
	}
	log.Debug(output)
	return output, nil
}

func RunCommand(name string, arg ...string) (string, error) {
	return RunCommandExt(exec.Command(name, arg...))
}

// WaitPIDOpts are options to WaitPID
type WaitPIDOpts struct {
	PollInterval time.Duration
	Timeout      time.Duration
}

// WaitPID waits for a non-child process id to exit
func WaitPID(pid int, opts ...WaitPIDOpts) error {
	if runtime.GOOS != "linux" {
		return errors.Errorf("Platform '%s' unsupported", runtime.GOOS)
	}
	var timeout time.Duration
	var pollInterval = time.Second
	if len(opts) > 0 {
		if opts[0].PollInterval != 0 {
			pollInterval = opts[0].PollInterval
		}
		if opts[0].Timeout != 0 {
			timeout = opts[0].Timeout
		}
	}
	path := fmt.Sprintf("/proc/%d", pid)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	var timoutCh <-chan time.Time
	if timeout != 0 {
		timoutCh = time.NewTimer(timeout).C
	}
	for {
		select {
		case <-ticker.C:
			_, err := os.Stat(path)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return errors.WithStack(err)
			}
		case <-timoutCh:
			return ErrWaitPIDTimeout
		}
	}
}
