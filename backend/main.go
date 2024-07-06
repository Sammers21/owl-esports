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
	storage := dotabuff.NewStorage() // Use the correct package name for NewStorage
	err := storage.LoadHeroes()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading heroes")
	}
	err = storage.LoadCounters()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading counters")
	}

	heroes, err := storage.FindHeroes([]string{"brew", "sniper", "et", "pang", "hood", 
								"phoenix", "mk", "ck", "marci", "enigma"})
	if err != nil {
		log.Fatal().Err(err).Msg("Error finding heroes")
	}
	radiant := heroes[:5]
	dire := heroes[5:]
	rw, dw := storage.PickWinRate(radiant, dire)
	log.Info().Msgf("Radiant winrate: %.2f%%", rw)
	log.Info().Msgf("Dire winrate: %.2f%%", dw)
}
