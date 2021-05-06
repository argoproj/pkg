package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func LookupEnvStringOr(key string, o string) string {
	v, found := os.LookupEnv(key)
	if found {
		return v
	}
	return o
}

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
