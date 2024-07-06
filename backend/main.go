package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.io/sammers21/owl-esports/backend/dotabuff"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	heroes, err := dotabuff.Heroes()
	if err != nil {
		log.Printf("Error fetching heroes: %v", err)
		return
	}
	log.Info().Msg(fmt.Sprintf("Heroes: %v", heroes))
}
