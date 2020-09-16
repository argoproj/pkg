package sync

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyLock(t *testing.T) {
	l := NewKeyLock()

	l.Lock("my-key")

	unlocked := false

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		l.Lock("my-key")
		unlocked = true
		wg.Done()
	}()

	assert.False(t, unlocked)

	l.Unlock("my-key")

	wg.Wait()

	assert.True(t, unlocked)

	l.Unlock("my-key")
}
