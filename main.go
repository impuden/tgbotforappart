package main

import (
	"fmt"
	"log"
	"os"
	"tgappart/config"
	"tgappart/data"

	"tgappart/kbbutton"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	// Инициируем логгирование
	logFile, err := os.OpenFile("bot.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Не удалось открыть файл логов: %v", err)
	}
	defer logFile.Close()

	// Вызываем подключение к базе данных из файла
	db, err := data.ConnectDB()
	if err != nil {
		log.Fatal(err)
		fmt.Println("data connection fail")
	}
	defer db.Close()

	log.SetOutput(logFile)

	// Инициируем самого бота
	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	// Открываем канал для обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}
		if update.Message != nil {
			if update.Message.IsCommand() {
				kbbutton.HandleCommand(bot, update.Message)
			}
		} else if update.CallbackQuery != nil {
			kbbutton.Callback(bot, update.CallbackQuery, db, updates)
		}
	}

}
