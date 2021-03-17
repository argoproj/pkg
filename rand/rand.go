package rand

import (
	crand "crypto/rand"
	"math/rand"
	"sync"
	"time"
)

const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var srcMutex = sync.Mutex{}
var src = rand.NewSource(time.Now().UnixNano())

// RandString returns a cryptographically-secure pseudo-random alpha-numeric string of a given length
func RandString(n int) string {
	return RandStringCharset(n, letterBytes)
}

// RandStringCharset generates, from a given charset, a cryptographically-secure pseudo-random string of a given length
func RandStringCharset(n int, charset string) string {
	srcMutex.Lock()
	defer srcMutex.Unlock()

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(charset) {
			b[i] = charset[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

// SecureRandString returns a cryptographically-secure pseudo-random
// alpha-numeric string of a given length using crypto/rand
func SecureRandString(n int) (string, error) {
	str, err := SecureRandStringCharset(n, letterBytes)
	if err != nil {
		return "", err
	}
	return str, nil
}

func SecureRandStringCharset(n int, charset string) (string, error) {
	var err error
	b := make([]byte, n)
	cacheSize := int(float64(n) * 1.3)
	for i, j, cache := n-1, 0, []byte{}; i >= 0; j++ {
		if j%cacheSize == 0 {
			if cache, err = secureRandomBytes(cacheSize); err != nil {
				return "", err
			}

		}
		if idx := int(cache[j%cacheSize] & letterIdxMask); idx < len(charset) {
			b[i] = charset[idx]
			i--
		}
	}
	return string(b), nil
}

// secureRandomBytes returns the requested number of bytes using crypto/rand
func secureRandomBytes(length int) ([]byte, error) {
	var randomBytes = make([]byte, length)
	_, err := crand.Read(randomBytes)
	if err != nil {
		return randomBytes, err
	}
	return randomBytes, nil
}
