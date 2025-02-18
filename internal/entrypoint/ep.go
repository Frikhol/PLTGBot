package entrypoint

import (
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"log"
	"tgBot/internal/config"
	nt "tgBot/internal/interface/noticer"
)

func Run(cfg *config.Config, logger *zap.Logger) error {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		logger.Info("Failed to initialize bot", zap.Error(err))
	}

	connStr := "postgres://test:test@localhost:5432/plnoticer?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Info("failed to connect to db", zap.Error(err))
	}
	defer db.Close()

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	var noticer nt.Noticer = nt.NewNoticer()

	for update := range updates {
		if update.Message != nil {
			//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			if update.Message.IsCommand() {
				if update.Message.Command() == "start" {
					noticer.Start(bot, update, db, logger)
					continue
				}
				if update.Message.Command() == "register" {
					//TODO: must impl
					continue
				}
				if update.Message.Command() == "reserve" {
					//TODO: must impl
					continue
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда")
				if _, err := bot.Send(msg); err != nil {
					logger.Info("Failed to send message", zap.Error(err))
				}
			}
		}
	}
	return nil
}
