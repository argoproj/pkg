package cli

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestSetLogTimestampFormat
func TestSetLogTimestampFormat(t *testing.T) {
	message, _ := log.NewEntry(log.StandardLogger()).String()
	assert.Contains(t, message, "0001-01-01T00:00:00Z")

	SetLogTimestampFormat("2006-01-02T15:04:05.000Z")

	message, _ = log.NewEntry(log.StandardLogger()).String()
	assert.Contains(t, message, "0001-01-01T00:00:00.000Z")

	SetLogTimestampFormat("01-02-2006T15:04Z")

	message, _ = log.NewEntry(log.StandardLogger()).String()
	assert.Contains(t, message, "01-01-0001T00:00Z")
}
