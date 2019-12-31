//+build windows

package stats

// RegisterStackDumper spawns a goroutine which dumps stack trace upon a SIGUSR1
func RegisterStackDumper() {
	// NOOP
}

// RegisterHeapDumper spawns a goroutine which dumps heap profile upon a SIGUSR2
func RegisterHeapDumper(filePath string) {
	// NOOP
}
