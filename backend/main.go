package main

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.io/sammers21/owl-esports/backend/bot"
	"github.io/sammers21/owl-esports/backend/dotabuff"
	"github.io/sammers21/owl-esports/backend/http"
)

func main() {
	telegramTokenCli := flag.String("t", "", "Telegram bot token")
	flag.Parse()
	telegramToken := *telegramTokenCli
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	engine := dotabuff.NewEngine()
	if telegramToken != "" {
		log.Info().Str("token", telegramToken).Msg("Starting telegram bot")
		telegramBot := bot.NewTelegramBot(engine, telegramToken)
		go telegramBot.Start()
	} else {
		log.Info().Msg("Telegram bot token not provided")
	}
	server := http.NewServer(engine)
	go server.Start(8080)
	err := engine.LoadHeroes()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading heroes")
	}
	err = engine.LoadCounters()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading counters")
	}
	select {}
}
