package kbbutton

import (
	"database/sql"
	"fmt"
	"log"
	"tgappart/data"
	"tgappart/usermap"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type ConfirmationRequest struct {
	Payer   string
	Payee   string
	ChatID  int64
	PayerID int64
	ReplyTo int
}

var confirmationRequests = make(map[string]*ConfirmationRequest)

func ShowDebt(db *sql.DB, bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, chatID int64, userID int) {
	username, exist := usermap.TelegramID[userID]
	if !exist {
		log.Println("this user didnt exist", userID)
		return
	}

	ars, max, egor, vova, err := data.Readdebt(db, username)
	if err != nil {
		log.Println("fail reading debt", err)
		bot.Send(tgbotapi.NewMessage(chatID, "проблемка вышла с чтением данных, стукани создателю"))
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "")
		msg.ReplyMarkup = MainMenu()
		return
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ты должен:\nАрсению: %.2f\nМаксиму: %.2f\nЕгору: %.2f\nВладимиру: %.2f", ars, max, egor, vova))
	msg.ReplyMarkup = DeleteDebtKB()
	_, err = bot.Send(msg)
	if err != nil {
		log.Println("fail when show debt", err)
		bot.Send(tgbotapi.NewMessage(chatID, "проблемка вышла с выводом данных, стукани создателю"))
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "")
		msg.ReplyMarkup = MainMenu()
	}

}

func Deldebt(userID int, bot *tgbotapi.BotAPI, updates <-chan tgbotapi.Update, chatID int64) {
	username, exist := usermap.TelegramID[userID]
	if !exist {
		log.Println("user not exist", userID)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "С кем ты хочешь расчитаться?")
	bot.Send(msg)

	for update := range updates {
		log.Println("Received update")
		if update.Message.Text == "" {
			log.Println("Empty message received")
			continue
		}

		payee := update.Message.Text

		var payeeID int64

		for id, name := range usermap.TelegramID {
			if name == payee {
				payeeID = int64(id)
				break
			}
		}

		if payeeID == 0 {
			msg := tgbotapi.NewMessage(chatID, "Нет у нас в квартире таких, повтори ввод.")
			bot.Send(msg)
			continue
		}

		log.Println("Creating confirmation request")
		confirmationRequest := &ConfirmationRequest{
			Payer:   username,
			Payee:   payee,
			ChatID:  payeeID,
			PayerID: chatID,
			ReplyTo: update.Message.MessageID,
		}
		confirmationRequests[payee] = confirmationRequest

		log.Printf("Confirmation Request: %+v", confirmationRequest)

		confirmMsg := tgbotapi.NewMessage(payeeID, fmt.Sprintf("Подтверди что %s вернул твои кровные", username))
		confirmMsg.ReplyMarkup = Accept()
		bot.Send(confirmMsg)
		log.Println("conf sent")

		return
	}
}

func Confirm(db *sql.DB, bot *tgbotapi.BotAPI, payee string) {
	confirmationRequest, exists := confirmationRequests[payee]
	if !exists {
		log.Println("no confirm request exist for this user")
		return
	}

	err := data.Deletedebt(db, confirmationRequest.Payer, confirmationRequest.Payee)
	if err != nil {
		log.Println("fail delete debt", err)
		msg := tgbotapi.NewMessage(confirmationRequest.ChatID, "Что-то не так при удалении долга, напиши создателю")
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(confirmationRequest.ChatID, "Расчитались, молодцы!")
	nsg := tgbotapi.NewMessage(confirmationRequest.PayerID, "Расчитались, молодцы!")
	msg.ReplyMarkup = MainMenu()
	nsg.ReplyMarkup = MainMenu()
	bot.Send(nsg)
	bot.Send(msg)
	delete(confirmationRequests, confirmationRequest.Payer)

}

func Decline(bot *tgbotapi.BotAPI, chatID int64, payee string) {
	confirmationRequest, exists := confirmationRequests[payee]
	if !exists {
		log.Println("no confirm request exist for this user, on decline")
		return
	}

	msg := tgbotapi.NewMessage(confirmationRequest.ChatID, "Отклонено.")
	nsg := tgbotapi.NewMessage(confirmationRequest.PayerID, "Отклонено другой стороной.")
	nsg.ReplyMarkup = MainMenu()
	msg.ReplyMarkup = MainMenu()
	bot.Send(msg)
	bot.Send(nsg)
	delete(confirmationRequests, confirmationRequest.Payer)

}
