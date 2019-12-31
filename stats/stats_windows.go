//+build windows

package stats

import log "github.com/sirupsen/logrus"

func RegisterStackDumper() {
	log.Warn("RegisterStackDumper is not supported on windows - noop")
}

func RegisterHeapDumper(filePath string) {
	log.Warn("RegisterHeapDumper is not supported on windows - noop")
}
