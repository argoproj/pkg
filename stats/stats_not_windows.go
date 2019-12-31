//+build !windows

package stats

import (
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// RegisterStackDumper spawns a goroutine which dumps stack trace upon a SIGUSR1
func RegisterStackDumper() {
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGUSR1)
		for {
			<-sigs
			LogStats()
			LogStack()
		}
	}()
}

// RegisterHeapDumper spawns a goroutine which dumps heap profile upon a SIGUSR2
func RegisterHeapDumper(filePath string) {
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGUSR2)
		for {
			<-sigs
			runtime.GC()
			if _, err := os.Stat(filePath); err == nil {
				err = os.Remove(filePath)
				if err != nil {
					log.Warnf("could not delete heap profile file: %v", err)
					return
				}
			}
			f, err := os.Create(filePath)
			if err != nil {
				log.Warnf("could not create heap profile file: %v", err)
				return
			}

			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Warnf("could not write heap profile: %v", err)
				return
			} else {
				log.Infof("dumped heap profile to %s", filePath)
			}
		}
	}()
}
