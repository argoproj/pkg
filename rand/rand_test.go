package rand

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandString(t *testing.T) {
	ss := RandStringCharset(10, "A")
	if ss != "AAAAAAAAAA" {
		t.Errorf("Expected 10 As, but got %q", ss)
	}
	ss = RandStringCharset(5, "ABC123")
	if len(ss) != 5 {
		t.Errorf("Expected random string of length 10, but got %q", ss)
	}
}

func TestSecureRandString(t *testing.T) {
	str, err := SecureRandStringCharset(10, "A")
	assert.NoError(t, err)
	assert.Equal(t, "AAAAAAAAAA", str)

	str, err = SecureRandStringCharset(5, "ABC123")
	assert.NoError(t, err)
	assert.Regexp(t, `[ABC123]{5,5}`, str)

	str, err = SecureRandStringCharset(52, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	assert.NoError(t, err)
	assert.Regexp(t, `[abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ]{52,52}`, str)
}
