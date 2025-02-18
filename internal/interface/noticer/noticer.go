package noticer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	database "tgBot/internal/infra/database"
	"time"
)

type Handler struct{}

type Noticer interface {
	Start(bot *tgbotapi.BotAPI, update tgbotapi.Update, psql *sql.DB, logger *zap.Logger) error
	SendEnnobleInfo(bot *tgbotapi.BotAPI, chatId int64, lastId int64) int64
}

func NewNoticer() *Handler {
	return &Handler{}
}

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

func (n *Handler) Start(bot *tgbotapi.BotAPI, update tgbotapi.Update, psql *sql.DB, logger *zap.Logger) error {
	db := database.NewDBHandler(psql, logger)
	chat, err := db.GetChat(update.Message.Chat.ID)
	if err != nil {
		chat = database.Chat{
			Id:         db.GetLastChatId() + 1,
			ChatId:     update.Message.Chat.ID,
			IsNoticing: true,
			LastLostId: 0,
			LastGetId:  0}
		db.InsertChat(chat)
	}

	if chat.IsNoticing {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Бот уже уведомляет этот чат")
		bot.Send(msg)
	}
	if !chat.IsNoticing {
		chat.IsNoticing = true
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Последние проёбы:")
		bot.Send(msg)
		chat.LastLostId = n.SendEnnobleInfo(bot, update.Message.Chat.ID, chat.LastLostId)
		go func() {
			for {
				time.Sleep(1 * time.Minute)
				chat.LastLostId = n.SendEnnobleInfo(bot, update.Message.Chat.ID, chat.LastLostId)
			}
		}()
	}
	db.UpdateChat(chat)
	return nil
}

func (n *Handler) SendEnnobleInfo(bot *tgbotapi.BotAPI, chatId int64, lastId int64) int64 {
	httpRequest := "https://twhelp.app/api/v2/versions/pl/servers/pl206/ennoblements?limit=100&sort=createdAt%3ADESC"
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

	index := 0
	for i, v := range data.Data {
		if v.Village.Player != nil && v.Village.Player.Tribe != nil && lastId != v.Id && v.Village.Player.Tribe.Id == 309 {
			continue
		}
		if lastId == v.Id {
			index = i - 1
			break
		}
		if i == 99 {
			index = 99
		}
	}
	for ; index >= 0; index-- {
		if data.Data[index].Village.Player != nil && data.Data[index].Village.Player.Tribe != nil && data.Data[index].Village.Player.Tribe.Id == 309 && lastId != data.Data[index].Id {
			lastId = data.Data[index].Id
			//if data.Data[index].NewOwner.Tribe.Id == 309 {
			//	continue
			//}
			ennobleTime, _ := time.Parse(time.RFC3339, data.Data[index].CreatedAt)
			ennobleTime = ennobleTime.Add(time.Hour)
			formatedTime := ennobleTime.Format("15:04:05 02.01.2006")
			villageInfo := fmt.Sprintf("<a href='%s'>%s</a>", data.Data[index].Village.ProfileUrl, data.Data[index].Village.FullName)
			oldOwnerInfo := fmt.Sprintf("<a href='%s'>%s</a>", data.Data[index].Village.Player.ProfileUrl, data.Data[index].Village.Player.Name)
			newOwnerInfo := fmt.Sprintf("<a href='%s'>%s</a>", data.Data[index].NewOwner.ProfileUrl, data.Data[index].NewOwner.Name)
			ownerTribeInfo := fmt.Sprintf("<a href='%s'>%s</a>", data.Data[index].NewOwner.Tribe.ProfileUrl, data.Data[index].NewOwner.Tribe.Name)
			msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("%s проебал хату в %s(PL)\nДеревня: %s\nПидарасина: %s\nПлемя: %s\n", oldOwnerInfo, formatedTime, villageInfo, newOwnerInfo, ownerTribeInfo))
			msg.ParseMode = tgbotapi.ModeHTML
			msg.DisableWebPagePreview = true
			bot.Send(msg)
		}
	}
	return lastId
}
