package main

import (
	"go.uber.org/zap"
	"log"
	"tgBot/internal/config"
	"tgBot/internal/entrypoint"
	"tgBot/internal/logger"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.GetConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %s\n", err.Error())
	}

	// Инициализация логгера
	zapLogger := logger.NewClientZapLogger(cfg.LogLevel, cfg.Name)

	// Запуск сервера
	if err = entrypoint.Run(cfg, zapLogger); err != nil {
		zapLogger.Fatal("Run failed: %s\n", zap.Error(err))
	}
}
