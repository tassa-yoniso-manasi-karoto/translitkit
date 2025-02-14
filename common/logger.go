
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
	SetLogger(zerolog.New(w).With().Timestamp().Logger())
}

func SetLogger(baseLogger zerolog.Logger) {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.TimeOnly,
	}
	writer.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("[translitkit] %s", i)
	}
	Log =  baseLogger.Output(writer)
}


func DisableLogger() {
       Log = zerolog.Nop()
}
