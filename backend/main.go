package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.io/sammers21/owl-esports/backend/dotabuff"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msg("Loading heroes...")
	heroes, err := dotabuff.Heroes()
	if err != nil {
		log.Printf("Error fetching heroes: %v", err)
		return
	}
	log.Info().
		Int("count", len(heroes)).
		Msg("Heroes has been loaded")


	for _, hero := range heroes {
		hero.Counters()
		break
		// counters, err := hero.Counters()
		// if err != nil {
		// 	log.Printf("Error fetching counters: %v", err)
		// 	return
		// }
		// log.Info().
		// 	Str("hero", hero.Name).
		// 	Int("counters-cnt", len(counters)).
		// 	Msg("Counters")
	}
}
