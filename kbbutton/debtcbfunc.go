package kbbutton

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"tgappart/data"
	"tgappart/usermap"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type ConfirmationRequest struct {
	Payer   string
	Payee   string
	ChatID  int64
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

type Nickel struct {
	Message tgbotapi.Message
}

var nickelCh = make(chan Nickel)

func Deldebt(userID int, bot *tgbotapi.BotAPI, updates <-chan tgbotapi.Update, chatID int64, db *sql.DB) {
	username, exist := usermap.TelegramID[userID]
	if !exist {
		log.Println("user not exist", userID)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "С кем ты хочешь расчитаться?")
	bot.Send(msg)

	var confirmationRequest *ConfirmationRequest

	go func() {
		for {
			select {
			case update := <-updates:
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
				confirmationRequest = &ConfirmationRequest{
					Payer:   username,
					Payee:   payee,
					ChatID:  payeeID,
					ReplyTo: update.Message.MessageID,
				}
				confirmationRequests[payee] = confirmationRequest

				confirmMsg := tgbotapi.NewMessage(payeeID, fmt.Sprintf("Подтверди что %s вернул твои кровные, напиши 'да'", username))
				bot.Send(confirmMsg)
				confirmMsg.ReplyToMessageID = update.Message.MessageID

				nickelCh <- Nickel{Message: *update.Message}
			}
		}
	}()

	for {
		select {
		case <-time.After(30 * time.Minute):
			log.Println("30 minutes timeout reached")
			msg := tgbotapi.NewMessage(confirmationRequest.ChatID, "Время вышло, теперь все по новой")
			bot.Send(msg)
			delete(confirmationRequests, confirmationRequest.Payer)
			return
		case nickel := <-nickelCh:
			log.Println("Received nickel message")
			text := strings.ToLower(strings.TrimSpace(nickel.Message.Text))
			if text == "да" {
				log.Println("have da")
				if nickel.Message.Chat.ID != confirmationRequest.ChatID {
					continue
				}

				err := data.Deletedebt(db, confirmationRequest.Payer, confirmationRequest.Payee)
				if err != nil {
					log.Println("fail delete debt", err)
					msg := tgbotapi.NewMessage(confirmationRequest.ChatID, "Что-то не так при удалении долга, напиши создателю")
					bot.Send(msg)
					return
				}

				msg := tgbotapi.NewMessage(confirmationRequest.ChatID, "Расчитались, молодцы!")
				msg.ReplyMarkup = MainMenu() // Ваша настраиваемая разметка
				bot.Send(msg)
				delete(confirmationRequests, confirmationRequest.Payer)
				return
			} else {
				msg := tgbotapi.NewMessage(confirmationRequest.ChatID, "Расчет не подтвердился, не фортануло..")
				msg.ReplyMarkup = MainMenu() // Ваша настраиваемая разметка
				bot.Send(msg)
				delete(confirmationRequests, confirmationRequest.Payer)
				return
			}
		}
	}
}
