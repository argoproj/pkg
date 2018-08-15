package exec

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// RunCommand is a convenience function to run/log a command and return/log stderr in an error upon
// failure. Chops off a trailing newline (if present)
func RunCommand(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
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
