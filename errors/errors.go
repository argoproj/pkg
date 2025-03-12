package errors

import (
	log "github.com/sirupsen/logrus"
)

// CheckError is a convenience function to fatally log an exit if the supplied error is non-nil
// Deprecated: this function is trivial and should be implemented by the caller instead.
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
