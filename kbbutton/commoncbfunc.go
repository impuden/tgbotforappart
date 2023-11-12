package kbbutton

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"tgappart/data"
	"tgappart/usermap"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type UserContext1 struct {
	ChatID      int64
	Step        int
	UserName    string
	Value       float64
	LastMessage int
}

var userContextMap1 = make(map[int64]*UserContext1)

func AddCommon(db *sql.DB, bot *tgbotapi.BotAPI, chatID int64, updates <-chan tgbotapi.Update, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	username, exist := usermap.TelegramID[userID]

	if !exist {
		msg := tgbotapi.NewMessage(chatID, "Ваше имя не найдено в базе для этой квартиры.")
		bot.Send(msg)
		return
	}

	ctx, exist := userContextMap1[chatID]
	if !exist {
		ctx = &UserContext1{
			ChatID:   chatID,
			UserName: username,
			Step:     0,
		}
		userContextMap1[chatID] = ctx
		msg := tgbotapi.NewMessage(chatID, "Напиши общую сумму покупки и после пробела перечисли наименования.")
		sentMsg, err := bot.Send(msg)
		if err != nil {
			log.Printf("Ошибка при отправке: %v", err)
			return
		}

		ctx.LastMessage = sentMsg.MessageID

	}

	for update := range updates {
		if ctx.Step == 0 {
			if update.Message != nil && update.Message.Chat.ID == chatID {
				text := update.Message.Text
				parts := strings.Fields(text)
				if len(parts) < 2 {
					response := "слишком коротко, повтори"
					msg := tgbotapi.NewMessage(chatID, response)
					bot.Send(msg)
					continue
				}

				value, err := strconv.ParseFloat(parts[0], 64)
				if err != nil {
					response := "странное число, повтори"
					msg := tgbotapi.NewMessage(chatID, response)
					bot.Send(msg)
					log.Println("ошибка странного числа", value)
					continue
				}

				ctx.Value = value
				comment := strings.Join(parts[1:], " ")
				ctx.Step = 1
				log.Printf("Данные для БД: username=%s, value=%f, comment=%s", username, value, comment)

				err = data.Writeexp(db, username, value, comment)
				if err != nil {
					errMsg := "Неладное вышло, не записались твои копеечки. Стукани кодеру, пусть посмотрит логи"
					msg := tgbotapi.NewMessage(chatID, errMsg)
					bot.Send(msg)
					msg.ReplyMarkup = MainMenu()
					log.Printf("Ошибка при записи в БД: %v", err)
					return
				}

				log.Printf("Данные записаны в БД: username=%s, value=%f, comment=%s", username, value, comment)

				msg := tgbotapi.NewMessage(chatID, "Твои потраченные кровные учтены, будь здоров.")
				msg.ReplyMarkup = MainMenu()
				bot.Send(msg)
				data.Writedebt(db)
				return
			}
		}
	}
}

func Printcommon(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, db *sql.DB) {
	var messageText string

	for _, name := range usermap.TelegramID {
		value, comment, err := data.Getexp(db, name)
		if err != nil {
			log.Printf("Ошибка при получении значения для %s: %v", name, err)
			continue
		}
		messageText += fmt.Sprintf("%s: %.0f, %s\n", name, value, comment)
	}

	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, messageText)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func Resetexp(bot *tgbotapi.BotAPI, chatID int64, userID int) {
	_, exist := usermap.TelegramID[userID]

	if !exist {
		msg := tgbotapi.NewMessage(chatID, "Ваше имя не найдено в базе для этой квартиры.")
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Ты уверен что хочешь удалить данные о ВСЕХ своих затратах, без возможности восстановить данные?")
	msg.ReplyMarkup = DelExpKB()
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
		return
	}
}

func DeleteExp(bot *tgbotapi.BotAPI, chatID int64, db *sql.DB, userID int) {
	username, exist := usermap.TelegramID[userID]
	if !exist {
		msg := tgbotapi.NewMessage(chatID, "Вас нет в этой квартире.")
		bot.Send(msg)
		return
	}

	// Выполнение удаления данных из БД
	data.Delexp(db, username)
	data.Writedebt(db)
	log.Println("Данные удалены пользователем", username)
	// Уведомление об удалении
	msg := tgbotapi.NewMessage(chatID, "Данные удалены.")
	msg.ReplyMarkup = MainMenu()
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
	}
}

func DeclineDelExp(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Ясно, данные не удалены, возвращаю в главное меню.")
	msg.ReplyMarkup = CreateInlineKeyboard()
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Ошибка при отправке сообщения:", err)
	}

}
