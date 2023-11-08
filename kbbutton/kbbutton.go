package kbbutton

import (
	"log"
	"tgappart/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// CreateInlineKeyboard создает инлайн-клавиатуру.
func CreateInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	button1 := tgbotapi.NewInlineKeyboardButtonData("ЖКХ", "zhkh")
	button2 := tgbotapi.NewInlineKeyboardButtonData("Общие расходы", "common")
	button3 := tgbotapi.NewInlineKeyboardButtonData("Добавить расходы", "add")
	button4 := tgbotapi.NewInlineKeyboardButtonData("Обнулить свои расходы", "delete")
	button5 := tgbotapi.NewInlineKeyboardButtonData("Мои долги", "mydlg")
	button6 := tgbotapi.NewInlineKeyboardButtonData("Просто картинка с котиком ^_^", "cat")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(button1),
		tgbotapi.NewInlineKeyboardRow(button2),
		tgbotapi.NewInlineKeyboardRow(button3),
		tgbotapi.NewInlineKeyboardRow(button4),
		tgbotapi.NewInlineKeyboardRow(button5),
		tgbotapi.NewInlineKeyboardRow(button6),
	)

	return keyboard
}

// HandleCommand обрабатывает команды от пользователя.
func HandleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	msg.ParseMode = "Markdown"
	switch message.Command() {
	case "start":
		msg.Text = "Привет! Нажми одну из кнопок:"
		msg.ReplyMarkup = CreateInlineKeyboard()
		config.HandleUnknownUser(bot, message)
	default:
		msg.Text = "Для начала нажми /start."
	}
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func Catkb() tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Еще котика)", "cat"),
			tgbotapi.NewInlineKeyboardButtonData("В главное меню", "back"),
		),
	)
	return keyboard
}

func AddExpensesKeyboard() tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить", "addcommon"),
			tgbotapi.NewInlineKeyboardButtonData("В главное меню", "back"),
		),
	)

	return keyboard
}

func Addzhkhboard() tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить", "addzhkh"),
			tgbotapi.NewInlineKeyboardButtonData("В главное меню", "back"),
		),
	)

	return keyboard
}

func DeleteZhkhBoard() tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Удалить, если уже оплатили. (отменить нельзя)", "delzhkh")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("В главное меню", "back")),
	)

	return keyboard
}

func MainMenu() tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("В главное меню", "back")),
	)
	return keyboard
}

func DeleteDebtKB() tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Закрыть долг", "deldebt")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("В главное меню", "back")),
	)

	return keyboard
}
