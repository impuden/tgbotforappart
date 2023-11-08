package kbbutton

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"tgappart/data"
	"tgappart/usermap"
	"time"

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
			msg.Text = "ЖКХ уже оплачено, можете внести данные новой квитанции"
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
		// Отправляем сообщение с приглашением к вводу данных
		msg.Text = "Введи данные с квитанции просто числом, учти сразу все затраты, так как свет или иная платежка может быть отдельно"
		bot.Send(msg)
		// Переводим контекст пользователя на следующий шаг
		userContext.Step = 1
		// Ожидание ввода пользователя
		for update := range updates {

			if update.Message != nil && update.Message.Chat.ID == chatID {
				go func(update tgbotapi.Update) {
					// Обработка ввода пользователя
					input := update.Message.Text
					value, err := strconv.ParseFloat(input, 64)
					//comments :=
					if err != nil {
						msg.Text = "Неверный формат числа. Введите число (в формате 0.0):"
						bot.Send(msg)
					} else {
						// Шаг 2: Запись значения в базу данных
						err = data.Writezhkh(db, chatID, value)
						if err != nil {
							errMsg := "Произошла ошибка при сохранении данных. Пожалуйста, попробуйте позже."
							msg.Text = errMsg
							bot.Send(msg)
							log.Fatal(err)
						} else {
							response := fmt.Sprintf("Значение %f было записано в базу данных.", value)
							msg.Text = response
							data.Writedebt(db)
							//bot.Send(msg)

							// Сброс контекста пользователя
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

var isProcessing = false

func Delzhkhbutton(bot *tgbotapi.BotAPI, chatID int64, db *sql.DB, userID int, updates <-chan tgbotapi.Update) {
	if isProcessing {
		msg := tgbotapi.NewMessage(chatID, "Кто-то уже удаляет данные ЖКХ")
		bot.Send(msg)
		return
	}

	isProcessing = true
	defer func() {
		isProcessing = false
	}()

	_, exist := usermap.TelegramID[userID]

	if !exist {
		msg := tgbotapi.NewMessage(chatID, "Ваше имя не найдено в базе для этой квартиры.")
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Ты уверен что хочешь удалить данные о последней квитанции ЖКХ?\nРекомендую делать это только после оплаты.\nНапишите 'Да' для подтверждения.\nУчти что мой создатель еще учится и для выхода с этого шага надо что-то написать боту :)")
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
		return
	}

	select {
	case <-time.After(15 * time.Second):
		msg := tgbotapi.NewMessage(chatID, "Время вышло, теперь все заново")
		bot.Send(msg)
		return
	case update := <-updates:
		if update.Message != nil && update.Message.Chat.ID == chatID {
			if update.Message.Text == "Да" || update.Message.Text == "да" {
				go func() {
					// Выполнение удаления данных из БД
					error := data.Delzhkh(db, chatID)
					if error != nil {
						errMsg := "Неладное вышло, не удалилась запись ЖКХ. Стукани кодеру, пусть посмотрит логи"
						msg := tgbotapi.NewMessage(chatID, errMsg)
						bot.Send(msg)
						msg.ReplyMarkup = MainMenu()
						log.Printf("Ошибка при записи в БД: %v", error)
						return
					}
					// Уведомление об удалении
					msg := tgbotapi.NewMessage(chatID, "Данные удалены.")
					msg.ReplyMarkup = MainMenu()
					_, err := bot.Send(msg)
					if err != nil {
						log.Println("Ошибка при отправке сообщения:", err)
					}
				}()
				break
			} else {
				// В случае, если ответ не 'Да'
				msg := tgbotapi.NewMessage(chatID, "Я тебя не понял, повторить можно только через главное меню.")
				msg.ReplyMarkup = MainMenu()
				_, err := bot.Send(msg)
				if err != nil {
					log.Println("Ошибка при отправке сообщения:", err)
				}
				break
			}
		}

	}

}
