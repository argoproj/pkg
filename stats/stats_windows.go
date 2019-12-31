package stats

import (
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
)

// StartStatsTicker starts a goroutine which dumps stats at a specified interval
func StartStatsTicker(d time.Duration) {
	ticker := time.NewTicker(d)
	go func() {
		for {
			<-ticker.C
			LogStats()
		}
	}()
}

// RegisterStackDumper spawns a goroutine which dumps stack trace upon a SIGUSR1
func RegisterStackDumper() {
	log.Warn("RegisterStackDumper is not supported on windows - noop")
}

// RegisterHeapDumper spawns a goroutine which dumps heap profile upon a SIGUSR2
func RegisterHeapDumper(filePath string) {
	log.Warn("RegisterHeapDumper is not supported on windows - noop")
}

// LogStats logs runtime statistics
func LogStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Infof("Alloc=%v TotalAlloc=%v Sys=%v NumGC=%v Goroutines=%d",
		m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC, runtime.NumGoroutine())

}

// LogStack will log the current stack
func LogStack() {
	buf := make([]byte, 1<<20)
	stacklen := runtime.Stack(buf, true)
	log.Infof("*** goroutine dump...\n%s\n*** end\n", buf[:stacklen])
}
