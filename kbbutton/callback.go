package kbbutton

import (
	"database/sql"
	"fmt"
	"tgappart/usermap"

	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Callback обрабатывает callback-запросы от кнопок.
func Callback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, db *sql.DB, updates <-chan tgbotapi.Update) {
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "")
	chatID := callback.Message.Chat.ID
	userID := callback.From.ID

	data := callback.Data

	switch data {
	case "zhkh":
		callback.Data = "ЖКХ"
		HandleZhkhCallback(bot, callback, db)
	case "common":
		Printcommon(bot, callback, db)
		msg.ReplyMarkup = MainMenu()
	case "add":
		msg.Text = "Потратился? Запиши!\nТыкни 'Добавить'"
		msg.ReplyMarkup = AddExpensesKeyboard()
	case "delete":
		Resetexp(bot, chatID, userID)
	case "mydlg":
		msg.ReplyMarkup = DeleteDebtKB()
		ShowDebt(db, bot, callback, chatID, userID)
		msg.ReplyMarkup = DeleteDebtKB()
	case "back":
		msg.Text = "Вернулись назад"
		msg.ReplyMarkup = CreateInlineKeyboard()
	case "addzhkh":
		Addzhkh(bot, db, chatID, updates, callback)
	case "delzhkh":
		Delzhkhbutton(bot, chatID, db, userID)
	case "addcommon":
		AddCommon(db, bot, callback.Message.Chat.ID, updates, callback)
	case "deldebt":
		Deldebt(userID, bot, updates, chatID)
	case "accept":
		payee := usermap.TelegramID[userID]
		log.Println("payee", payee)
		Confirm(db, bot, payee)
	case "decline":
		payee := usermap.TelegramID[userID]
		Decline(bot, chatID, payee)
	case "accdelexp":
		DeleteExp(bot, chatID, db, userID)
	case "decldelexp":
		DeclineDelExp(bot, chatID)
	case "accdelzhkh":
		AccDelZHKH(bot, chatID, db)
	case "cat":
		err := Cat(bot, chatID)
		if err != nil {
			log.Fatal("Error sending cat image to Telegram:", err)
		}

		// Отправка клавиатуры после отправки изображения
		msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
		msg.ReplyMarkup = Catkb()
		_, err = bot.Send(msg)
		if err != nil {
			log.Fatal("Error sending keyboard to Telegram:", err)
		}
	}

	if msg.Text != "" {
		_, err := bot.Send(msg)
		if err != nil {
			log.Panic(err)
			fmt.Println("fail in callback")
		}
	}
}
