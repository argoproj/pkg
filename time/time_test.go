package time

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestParseDuration tests TestParseDuration
func TestParseDuration(t *testing.T) {
	type testData struct {
		duration string
		xVal     time.Duration
	}
	testdata := []testData{
		{"1s", time.Second},
		{"10s", 10 * time.Second},
		{"60s", time.Minute},
		{"1m", time.Minute},
		{"1h", time.Hour},
		{"1d", 24 * time.Hour},
		{"2d", 48 * time.Hour},
	}
	for _, data := range testdata {
		dur, err := ParseDuration(data.duration)
		assert.Nil(t, err)
		assert.Equal(t, dur.Nanoseconds(), data.xVal.Nanoseconds())
	}
	_, err := ParseDuration("1z")
	assert.NotNil(t, err)
}

// TestParseSince tests parsing of since strings
func TestParseSince(t *testing.T) {
	oneDayAgo, err := ParseSince("1d")
	assert.Nil(t, err)
	yesterday := time.Now().UTC().Add(-24 * time.Hour)
	assert.Equal(t, yesterday.Minute(), oneDayAgo.Minute())
}
