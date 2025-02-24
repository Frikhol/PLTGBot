package entrypoint

import (
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"tgBot/internal/config"
	nt "tgBot/internal/interface/noticer"
)

func Run(cfg *config.Config, logger *zap.Logger) error {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		logger.Info("Failed to initialize bot", zap.Error(err))
	}

	connStr := fmt.Sprintf("postgres://%s:%s@localhost:5432/plnoticer?sslmode=disable", cfg.DbUser, cfg.DbPass)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Info("failed to connect to db", zap.Error(err))
	}
	defer db.Close()

	//bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	var noticer nt.Noticer = nt.NewNoticer()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		for update := range updates {
			if update.Message != nil {
				//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
				if update.Message.IsCommand() {
					if update.Message.Command() == "notice" {
						noticer.Start(bot, update, db, logger)
						continue
					}
					if update.Message.Command() == "register" {
						noticer.Register(bot, update, db, logger)
						continue
					}
					if update.Message.Command() == "reserve" {
						noticer.Reserve(bot, update, db, logger)
						continue
					}
					if update.Message.Command() == "rename" {
						noticer.Rename(bot, update, db, logger)
						continue
					}
					if update.Message.Command() == "check" {
						noticer.Check(bot, update, db, logger)
						continue
					}
					if update.Message.Command() == "clear" {
						noticer.Clear(bot, update, db, logger)
						continue
					}
					if update.Message.Command() == "help" {
						noticer.Help(bot, update, db, logger)
						continue
					}
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда, используй /help")
					if _, err := bot.Send(msg); err != nil {
						logger.Info("Failed to send message", zap.Error(err))
					}
				}
			}
		}
	}()

	<-stop
	noticer.Stop(db, logger)
	return nil
}
