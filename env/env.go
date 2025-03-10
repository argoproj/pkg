// Deprecated: this package is not used by any Argo project and will be removed in the next major version of this
// library.
package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Deprecated: this function is not used by any Argo project and will be removed in the next major version of this
// library.
func LookupEnvStringOr(key string, o string) string {
	v, found := os.LookupEnv(key)
	if found {
		return v
	}
	return o
}

// Deprecated: this function is not used by any Argo project and will be removed in the next major version of this
// library.
func LookupEnvDurationOr(key string, o time.Duration) time.Duration {
	v, found := os.LookupEnv(key)
	if found {
		d, err := time.ParseDuration(v)
		if err != nil {
			panic(fmt.Errorf("%s=%s: failed to convert to duration", key, v))
		} else {
			return d
		}
	}
	return o
}

// Deprecated: this function is not used by any Argo project and will be removed in the next major version of this
// library.
func LookupEnvIntOr(key string, o int) int {
	v, found := os.LookupEnv(key)
	if found {
		d, err := strconv.Atoi(v)
		if err != nil {
			panic(fmt.Errorf("%s=%s: failed to convert to int", key, v))
		} else {
			return d
		}
	}
	return o
}

// Deprecated: this function is not used by any Argo project and will be removed in the next major version of this
// library.
func LookupEnvFloatOr(key string, o float64) float64 {
	v, found := os.LookupEnv(key)
	if found {
		d, err := strconv.ParseFloat(v, 64)
		if err != nil {
			panic(fmt.Errorf("%s=%s: failed to convert to float", key, v))
		} else {
			return d
		}
	}
	return o
}
