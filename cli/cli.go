package cli

import (
	"flag"
	"strconv"

	log "github.com/sirupsen/logrus"
	"k8s.io/klog/v2"

	"github.com/argoproj/pkg/errors"
)

// SetLogLevel parses and sets a logrus log level
func SetLogLevel(logLevel string) {
	level, err := log.ParseLevel(logLevel)
	errors.CheckError(err)
	log.SetLevel(level)
}

// SetGLogLevel set the glog level for the k8s go-client
func SetGLogLevel(glogLevel int) {
	klog.InitFlags(nil)
	_ = flag.Set("logtostderr", "true")
	_ = flag.Set("v", strconv.Itoa(glogLevel))
}
