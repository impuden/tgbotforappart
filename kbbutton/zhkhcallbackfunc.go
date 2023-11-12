package kbbutton

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"tgappart/data"
	"tgappart/usermap"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type UserContext struct {
	ChatID int64
	Step   int    // Шаг взаимодействия с пользователем
	Data   string // Временное хранение введенных данных
}

var userContextMap = make(map[int64]*UserContext) // Карта для хранения контекста пользователей

func HandleZhkhCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, db *sql.DB) {
	vata := callback.Data
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "")

	if vata == "ЖКХ" {
		zhkhAmount, err := data.GetZhkhAmount(db)
		if err != nil {
			log.Panic(err)
			fmt.Println("fail in handle")
			return
		}

		quaterzhkh := zhkhAmount / 4

		if zhkhAmount == 0 {
			msg.Text = "ЖКХ было обнулено, можете внести данные новой квитанции"
			msg.ReplyMarkup = Addzhkhboard()
		} else {
			msg.Text = fmt.Sprintf("ЖКХ не оплачено, сумма: %d. Для изменения можете удалить.\nПо умолчанию каждый должен %d", zhkhAmount, quaterzhkh)
			msg.ReplyMarkup = DeleteZhkhBoard()
		}
	}

	_, err := bot.Send(msg)
	if err != nil {
		log.Panic(err)
		fmt.Println("fail in handle")
	}
}

func Addzhkh(bot *tgbotapi.BotAPI, db *sql.DB, chatID int64, updates <-chan tgbotapi.Update, callback *tgbotapi.CallbackQuery) {
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "")
	userContext, exists := userContextMap[chatID]

	if !exists {
		userContext = &UserContext{ChatID: chatID}
		userContextMap[chatID] = userContext
	}

	if userContext.Step == 0 {
		msg.Text = "Введи данные с квитанции просто числом, учти сразу все затраты, так как свет или иная платежка может быть отдельно"
		bot.Send(msg)
		userContext.Step = 1
		for update := range updates {
			if update.Message != nil && update.Message.Chat.ID == chatID {
				go func(update tgbotapi.Update) {
					input := update.Message.Text
					value, err := strconv.ParseFloat(input, 64)
					if err != nil {
						msg.Text = "Неверный формат числа. Введите число (в формате 0.0):"
						bot.Send(msg)
					} else {
						err = data.Writezhkh(db, chatID, value)
						if err != nil {
							errMsg := "Произошла ошибка при сохранении данных. Пожалуйста, попробуйте позже."
							msg.Text = errMsg
							bot.Send(msg)
							log.Fatal(err)
						} else {
							response := fmt.Sprintf("Значение %.2f было записано в базу данных.", value)
							msg.Text = response
							userContext.Step = 0
							userContext.Data = ""
							msg.ReplyMarkup = MainMenu()
							bot.Send(msg)

						}
					}
				}(update)
				break
			}

		}

	}
}

func Delzhkhbutton(bot *tgbotapi.BotAPI, chatID int64, db *sql.DB, userID int) {
	_, exist := usermap.TelegramID[userID]

	if !exist {
		msg := tgbotapi.NewMessage(chatID, "Ваше имя не найдено в базе для этой квартиры.")
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Ты уверен что хочешь удалить данные о последней квитанции ЖКХ?\nРекомендую делать это только после оплаты.")
	msg.ReplyMarkup = AproveDelZHKH()
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
		return
	}
}

func AccDelZHKH(bot *tgbotapi.BotAPI, chatID int64, db *sql.DB) {
	error := data.Delzhkh(db, chatID)
	if error != nil {
		errMsg := "Неладное вышло, не удалилась запись ЖКХ. Стукани кодеру, пусть посмотрит логи"
		msg := tgbotapi.NewMessage(chatID, errMsg)
		bot.Send(msg)
		msg.ReplyMarkup = MainMenu()
		log.Printf("Ошибка при записи в БД: %v", error)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Данные удалены.")
	msg.ReplyMarkup = MainMenu()
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
	}
}
