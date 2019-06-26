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

type CmdOpts struct {
}

// RunCommandExt is a convenience function to run/log a command and return/log stderr in an error upon
// failure.
func RunCommandExt(cmd *exec.Cmd, opts CmdOpts) (string, error) {

	// log in a way we can copy-and-paste into a terminal
	args := strings.Join(cmd.Args, " ")
	log.WithFields(log.Fields{"dir": cmd.Dir}).Info(args)

	start := time.Now()
	outBytes, err := cmd.Output()
	output := string(outBytes)
	log.WithFields(log.Fields{"err": err, "duration": time.Since(start)}).Debug(output)

	if err != nil {
		exErr, ok := err.(*exec.ExitError)
		if ok {
			// clearly propagate diagnostics up
			err := fmt.Errorf("`%v` failed: %v", args, string(exErr.Stderr))
			log.Error(err)
			return output, err
		}
	}

	return output, err
}

func RunCommand(name string, opts CmdOpts, arg ...string) (string, error) {
	return RunCommandExt(exec.Command(name, arg...), opts)
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
