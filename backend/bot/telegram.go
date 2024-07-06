package bot

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"github.io/sammers21/owl-esports/backend/dotabuff"
)

type TelegramBot struct {
	Engine *dotabuff.Engine
	Token   string
}

func NewTelegramBot(engine *dotabuff.Engine, token string) *TelegramBot {
	return &TelegramBot{
		Engine: engine,
		Token:   token,
	}
}

func (b *TelegramBot) Start() error {
	log.Info().Msg("Starting telegram bot...")
	bot, err := tgbotapi.NewBotAPI(b.Token)
	if err != nil {
		log.Error().Err(err).Msgf("Error creating telegram bot with token %s", b.Token)
		panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil { // If we got a message
			text := update.Message.Text
			split := strings.Split(text, " ")
			if text == "/start" {
				log.Printf("[%s] %s", update.Message.From.UserName, text)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello! I'm a bot that can help you with Dota 2 hero counters. Just type the name of the hero you want to know the counters of and I'll provide you with the information.")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
				continue
			} else if len(split) == 10 {
				en
			}
			log.Printf("[%s] %s", update.Message.From.UserName, text)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}
	}
	panic("unreachable")
}
