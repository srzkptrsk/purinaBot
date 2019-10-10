package main

import (
	"github.com/jinzhu/gorm"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"strconv"
)

type Quiz struct {
	gorm.Model
	Day             int
	QuizQuestions   []QuizQuestion `gorm:"foreignkey:QuizID"`
	FirstQuestionId int
}

type QuizQuestion struct {
	gorm.Model
	QuizID      uint
	Question    string
	QuizAnswers []QuizAnswer `gorm:"foreignkey:QuizQuestionID"`
}

type QuizAnswer struct {
	gorm.Model
	QuizQuestionID uint
	Answer         string
	Text           string
	Photo          string
	NextQuestion   uint
}

func SendQuiz(quizId int, db *gorm.DB, bot *tgbotapi.BotAPI, userId int64) {
	var quiz Quiz
	db.Where("ID = ?", quizId).First(&quiz)

	if quiz.ID != 0 {
		SendQuestion(quiz.FirstQuestionId, db, bot, userId)
	}
}

func SendQuestion(questionId int, db *gorm.DB, bot *tgbotapi.BotAPI, userId int64) {
	var quizQuestion QuizQuestion
	db.Where("ID = ?", questionId).First(&quizQuestion)

	if quizQuestion.ID != 0 {
		msg := tgbotapi.NewMessage(userId, quizQuestion.Question)

		keyboard := tgbotapi.InlineKeyboardMarkup{}
		var answers []QuizAnswer
		db.Model(quizQuestion).Related(&answers)

		for _, answer := range answers {
			var row []tgbotapi.InlineKeyboardButton
			answerId := strconv.Itoa(int(answer.ID))
			btn := tgbotapi.NewInlineKeyboardButtonData(answer.Answer, answerId)
			row = append(row, btn)
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
		}

		msg.ReplyMarkup = keyboard

		_, err := bot.Send(msg)
		if err != nil {
			ProcessError(err.Error())
		}
	}
}

func ProcessAnswer(answerId int, db *gorm.DB, bot *tgbotapi.BotAPI, userId int64) {
	var answer QuizAnswer
	db.Where("ID = ?", answerId).First(&answer)

	if answer.ID != 0 {
		if answer.NextQuestion > 0 {
			SendQuestion(int(answer.NextQuestion), db, bot, userId)
		} else {
			if answer.Text != "0" {
				msg := tgbotapi.NewMessage(userId, answer.Text)
				msg.ParseMode = "markdown"

				_, err := bot.Send(msg)
				if err != nil {
					ProcessError(err.Error())
				}
			}

			if answer.Photo != "0" {
				msg := tgbotapi.NewPhotoUpload(userId, "media/"+answer.Photo)

				_, err := bot.Send(msg)
				if err != nil {
					ProcessError(err.Error())
				}
			}
		}
	}
}
