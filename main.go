package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log/syslog"
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"os"
	"strconv"
)

type Db struct {
	Dialect string
	Dsn     string
}

type Bot struct {
	Token string
}

func main() {
	conf := struct {
		Db  Db
		Bot Bot
	}{}

	file, _ := os.Open("configuration.json")
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&conf)
	if err != nil {
		log.Println("error:", err)
	}

	db, _ := gorm.Open(conf.Db.Dialect, conf.Db.Dsn)
	defer db.Close()
	db.AutoMigrate(&User{}, &Message{}, &Quiz{}, &QuizQuestion{}, &QuizAnswer{})

	bot, err := tgbotapi.NewBotAPI(conf.Bot.Token)
	if err != nil {
		log.Panic(err)
	}

	if len(os.Args) > 0 {
		for _, n := range os.Args[1:] {
			if n == "cron" {
				fmt.Println(n)
				SendMessages(db, bot)
				os.Exit(1)
			}
		}
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil && update.InlineQuery != nil {
			//query := update.InlineQuery.Query
			//fmt.Println(query)
			continue
		} else {
			if update.Message != nil {
				var user User
				db.Where("user_id = ?", update.Message.From.ID).First(&user)

				if user.UserId == 0 {
					user := User{UserId: update.Message.From.ID, Day: 1}
					db.NewRecord(user)
					db.Create(&user)
					ProcessDay(1, db, bot, int64(update.Message.From.ID))

					continue
				}

				day, _ := strconv.ParseInt(update.Message.Text, 10, 32)
				ProcessDay(int(day), db, bot, int64(update.Message.From.ID))
			} else {
				if update.CallbackQuery != nil {
					answerId, _ := strconv.ParseInt(update.CallbackQuery.Data, 10, 32)
					ProcessAnswer(int(answerId), db, bot, int64(update.CallbackQuery.From.ID))
					_, err := bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{update.CallbackQuery.ID, "🐈", false, "", 0})
					if err != nil {
						ProcessError(err.Error())
					}

					continue
				}
			}
		}
	}
}

func TimeIn(t time.Time, name string) (time.Time, error) {
	loc, err := time.LoadLocation(name)
	if err == nil {
		t = t.In(loc)
	}

	return t, err
}

func ProcessError(error string) {
	_, err := syslog.Dial("udp", "logs6.papertrailapp.com:48890", syslog.LOG_EMERG|syslog.LOG_KERN, "purinaBot")
	if err != nil {
		log.Fatal("failed to dial syslog")
	}
}
