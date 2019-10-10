/**
 * @copyright Piatrouski Software
 * @author Siaržuk Piatroŭski (siarzuk@piatrouski.com)
 */

package main

import (
	"github.com/jinzhu/gorm"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"time"
)

type User struct {
	gorm.Model
	UserId int
	Day    int
}

func SendMessages(db *gorm.DB, bot *tgbotapi.BotAPI) {
	var users []User
	db.Find(&users)

	for _, user := range users {
		if user.Day < 21 {
			nowTime, _ := TimeIn(time.Now(), "Europe/Minsk")
			userTime, _ := TimeIn(user.CreatedAt, "Europe/Minsk")

			if nowTime.Format("2019-01-01") != userTime.Format("2019-01-01") {
				user.Day = user.Day + 1
				db.Save(&user)
				ProcessDay(user.Day, db, bot, int64(user.UserId))

				time.Sleep(1 * time.Second)
			}
		}
	}
}
