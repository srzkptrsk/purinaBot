/**
 * @copyright Piatrouski Software
 * @author Siaržuk Piatroŭski (siarzuk@piatrouski.com)
 */

package main

import (
	"github.com/jinzhu/gorm"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type Message struct {
	gorm.Model
	Text   string
	Photo  string
	QuizId uint
	Day    int
}

func ProcessDay(day int, db *gorm.DB, bot *tgbotapi.BotAPI, userId int64) {
	var message Message
	db.Where("day = ?", day).First(&message)

	if message.ID != 0 {
		if message.Text != "0" {
			msg := tgbotapi.NewMessage(userId, message.Text)
			msg.ParseMode = "markdown"

			_, err := bot.Send(msg)
			if err != nil {
				ProcessError(err.Error())
			}
		}

		if message.Photo != "0" {
			msg := tgbotapi.NewPhotoUpload(userId, "media/"+message.Photo)

			_, err := bot.Send(msg)
			if err != nil {
				ProcessError(err.Error())
			}
		}

		if message.QuizId != 0 {
			SendQuiz(int(message.QuizId), db, bot, userId)
		}
	}
}
