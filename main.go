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

	httpRequest := "https://twhelp.app/api/v2/versions/pl/servers/pl206/tribes/309/ennoblements?limit=1&sort=createdAt%3ADESC"
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

	lastId := data.Data[0].Id
	for {
		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Panic(err)
		}
		if lastId != data.Data[0].Id {
			lastId = data.Data[0].Id
			msg := tgbotapi.NewMessage(281397467, fmt.Sprintf("Новый захват в %s\nДеревня: %s\nЗахватил: %s(%s)", data.Data[0].CreatedAt, data.Data[0].Village.FullName, data.Data[0].NewOwner.Name, data.Data[0].NewOwner.ProfileUrl))
			bot.Send(msg)
		}
		time.Sleep(1 * time.Minute)
	}
	//u := tgbotapi.NewUpdate(0)
	//u.Timeout = 60
	//
	//updates := bot.GetUpdatesChan(u)
	//
	//for update := range updates {
	//	if update.Message != nil {
	//		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	//
	//		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	//		bot.Send(msg)
	//	}
	//}
}
