package zerolog

import (
	"github.com/rs/zerolog"
)

func InitLogger(debug bool) {
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	InitDefaultLogger()
}
