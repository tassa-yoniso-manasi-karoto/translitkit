
package common

import (
	"github.com/rs/zerolog"
)

// logger is the package-level logger of common
var logger zerolog.Logger

func init() {
	//logger = zerolog.Nop()
}

func SetLogger(l zerolog.Logger) {
	logger = l
}

func GetLogger() zerolog.Logger {
	return logger
}