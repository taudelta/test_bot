package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/taudelta/test_bot/internal/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	var (
		dbFile      string
		cmdBotToken string
	)

	flag.StringVar(&dbFile, "db_file", "test.db", "путь до файла базы данных с вопросами")
	flag.StringVar(&cmdBotToken, "bot_token", "", "токен telegram бота")

	flag.Parse()

	botToken := os.Getenv("TEST_BOT_TOKEN")
	if cmdBotToken != "" {
		botToken = cmdBotToken
	}

	botAPI, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic("bot initialization error: ", err)
	}

	log.Println("Загружаем вопросы и ответы в память")
	bot.FillQuestions(dbFile)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Ожидание команд от пользователя")
	bot.Start(botAPI)
	sig := <-sigCh
	log.Println("Получен сигнал на завершение", sig)
	botAPI.StopReceivingUpdates()
}
