package main

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"net/http"
	"time"
)

type Tribe struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	ProfileUrl string `json:"profileUrl"`
	Tag        string `json:"tag"`
}

type Player struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	ProfileUrl string `json:"profileUrl"`
	Tribe      *Tribe `json:"tribe"`
}

type Village struct {
	Continent  string  `json:"continent"`
	FullName   string  `json:"fullName"`
	Id         int     `json:"id"`
	Player     *Player `json:"player"`
	ProfileUrl string  `json:"profileUrl"`
	X          int     `json:"x"`
	Y          int     `json:"y"`
}

type Ennoblement struct {
	CreatedAt string   `json:"createdAt"`
	Id        int64    `json:"id"`
	NewOwner  *Player  `json:"newOwner"`
	Points    int      `json:"points"`
	Village   *Village `json:"village"`
}

type Cursor struct {
	Next string `json:"next"`
	Self string `json:"self"`
}

type EnnobleData struct {
	Cursor Cursor        `json:"cursor"`
	Data   []Ennoblement `json:"data"`
}

func main() {
	bot, err := tgbotapi.NewBotAPI("7298682015:AAGv5nB23Ym_Dn-s760aQRMTDzbINJXJBCA")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	lastId := int64(0)

	go func() {
		for {
			lastId = sendEnnobleInfo(bot, 1, lastId)
			time.Sleep(1 * time.Minute)
		}
	}()

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			bot.Send(msg)
		}
	}
}

func sendEnnobleInfo(bot *tgbotapi.BotAPI, n int, last int64) int64 {
	httpRequest := fmt.Sprintf("https://twhelp.app/api/v2/versions/pl/servers/pl206/tribes/309/ennoblements?limit=%d&sort=createdAt%%3ADESC", n)
	resp, err := http.Get(httpRequest)
	if err != nil {
		log.Panic(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}

	var data EnnobleData
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Panic(err)
	}

	lastId := last
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Panic(err)
	}
	if lastId != data.Data[0].Id {
		lastId = data.Data[0].Id
		ennobleTime, _ := time.Parse(time.RFC3339, data.Data[0].CreatedAt)
		ennobleTime = ennobleTime.Add(time.Hour)
		formatedTime := ennobleTime.Format("15:04:05 02.01.2006")
		villageInfo := fmt.Sprintf("<a href='%s'>%s</a>", data.Data[0].Village.ProfileUrl, data.Data[0].Village.FullName)
		newOwnerInfo := fmt.Sprintf("<a href='%s'>%s</a>", data.Data[0].NewOwner.ProfileUrl, data.Data[0].NewOwner.Name)
		msg := tgbotapi.NewMessage(281397467, fmt.Sprintf("Новый захват в %s(PL)\nДеревня: %s\nЗахватил: %s", formatedTime, villageInfo, newOwnerInfo))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.DisableWebPagePreview = true
		bot.Send(msg)
	}
	return lastId
}
