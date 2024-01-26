// Fedchenko Alexey (fedchenko.alexey@r7-office.ru) R7-Office. 2024

package main

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
)

// SystemConfig contains system config
var SystemConfig ConfigData

func _check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func main() {

	// load environment variables
	SystemConfig = loadConfigFromEnv()

	// create bot using token, client
	bot, err := tgbotapi.NewBotAPI(SystemConfig.botAPIKey)
	_check(err)

	// debug mode on
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// set update interval
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 1000

	updates, err := bot.GetUpdatesChan(u)
	_check(err)

	// get new updates
	for update := range updates {
		// if message from user
		if update.Message != nil {
			if update.Message.Document != nil {
				fmt.Println("Get document", update.Message.Document.FileName)
				fmt.Println("\t", update.Message.Document.FileID)
				fmt.Println("\t", update.Message.Document.FileSize)

				if fileIsLicense(update.Message.Document.FileSize, update.Message.Document.FileName) {
					go replyDump(update.Message.Chat.ID, update.Message.MessageID, update.Message.Document.FileID, bot)
				} else {
					go replyMessage(update.Message.Chat.ID, update.Message.MessageID, "Файл не похож на файл лицензии", bot)
				}
			}
		}
	}
}

func replyMessage(chatId int64, messageId int, text string, bot *tgbotapi.BotAPI) {
	replyMessage := tgbotapi.NewMessage(chatId, text)
	replyMessage.ParseMode = "HTML"
	replyMessage.ReplyToMessageID = messageId

	fmt.Println("reply to", messageId)

	replyedMessage, _ := bot.Send(replyMessage)

	startTimerToDeleteMessage(replyedMessage.Chat.ID, replyedMessage.MessageID, bot)
}

func replyDump(chatId int64, messageId int, fileId string, bot *tgbotapi.BotAPI) {
	text := getLicenseDump(fileId)
	replyMessage(chatId, messageId, text, bot)
}

func startTimerToDeleteMessage(chatId int64, messageId int, bot *tgbotapi.BotAPI) {
	removeReplyTimer := time.NewTimer(time.Duration(SystemConfig.botAutodeleteTiming) * time.Second)
	go func() {
		<-removeReplyTimer.C
		fmt.Println("Message", messageId, "deleted")
		removeReplyTimer.Stop()
		deleteConfig := tgbotapi.NewDeleteMessage(int64(chatId), messageId)
		bot.DeleteMessage(deleteConfig)
	}()
}
