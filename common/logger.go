
package common

import (
	"os"
	"time"
	"fmt"
	
	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func init() {
	//Log = zerolog.Nop()
	w := zerolog.ConsoleWriter{
		Out: os.Stdout,
		TimeFormat: time.TimeOnly,
	}
	SetLoggerWriter(w)
}

func SetLoggerWriter(w zerolog.ConsoleWriter) {
	w.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("[translitkit] %s", i)
	}
	Log = zerolog.New(w).With().Timestamp().Logger()
}

func DisableLogger(l zerolog.Logger) {
       Log = zerolog.Nop()
}
