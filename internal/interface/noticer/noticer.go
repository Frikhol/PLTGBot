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
	"strings"
	"tgBot/internal/infra/database"
	"time"
)

type Handler struct{}

type Noticer interface {
	Start(bot *tgbotapi.BotAPI, update tgbotapi.Update, psql *sql.DB, logger *zap.Logger) error
	SendEnnobleInfo(bot *tgbotapi.BotAPI, chatId int64, lastId int64) int64
	Stop(psql *sql.DB, logger *zap.Logger) error
	Register(bot *tgbotapi.BotAPI, update tgbotapi.Update, psql *sql.DB, logger *zap.Logger) error
	Check(bot *tgbotapi.BotAPI, update tgbotapi.Update, psql *sql.DB, logger *zap.Logger) error
	Reserve(bot *tgbotapi.BotAPI, update tgbotapi.Update, psql *sql.DB, logger *zap.Logger) error
	Clear(bot *tgbotapi.BotAPI, update tgbotapi.Update, psql *sql.DB, logger *zap.Logger) error
	Help(bot *tgbotapi.BotAPI, update tgbotapi.Update, psql *sql.DB, logger *zap.Logger) error
	Rename(bot *tgbotapi.BotAPI, update tgbotapi.Update, psql *sql.DB, logger *zap.Logger) error
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
			IsNoticing: false,
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

func (n *Handler) Register(bot *tgbotapi.BotAPI, update tgbotapi.Update, psql *sql.DB, logger *zap.Logger) error {
	db := database.NewDBHandler(psql, logger)
	words := strings.Fields(update.Message.Text)
	if len(words) < 2 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверно выполненная команда: /register [nickname]")
		bot.Send(msg)
	} else {
		user, err := db.GetUser(update.Message.From.ID)
		if err != nil {
			user = database.User{
				Id:            db.GetLastUserId() + 1,
				Nickname:      words[1],
				TgId:          update.Message.From.ID,
				ReservedCount: 0}
			err = db.InsertUser(user)
			if err == nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Успешно зарегистрирован: %s", words[1]))
				bot.Send(msg)
			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ранее зарегистрирован: %s", user.Nickname))
			bot.Send(msg)
		}
	}
	return nil
}

func (n *Handler) Rename(bot *tgbotapi.BotAPI, update tgbotapi.Update, psql *sql.DB, logger *zap.Logger) error {
	db := database.NewDBHandler(psql, logger)
	words := strings.Fields(update.Message.Text)
	if len(words) < 2 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверно выполненная команда: /rename [nickname]")
		bot.Send(msg)
	} else {
		user, err := db.GetUser(update.Message.From.ID)
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Для начала нужно зарегистрироваться: %s", words[1]))
			bot.Send(msg)
		} else {
			err = db.UpdateUserName(user.Id, words[1])
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка")
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Успешно переименован: %s", words[1]))
				bot.Send(msg)
			}
		}
	}
	return nil
}

func (n *Handler) Check(bot *tgbotapi.BotAPI, update tgbotapi.Update, psql *sql.DB, logger *zap.Logger) error {
	db := database.NewDBHandler(psql, logger)
	_, err := db.GetUser(update.Message.From.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я тебя не знаю, выполни регистрацию : /register [nickname]")
		bot.Send(msg)
	} else {
		cords := strings.Fields(update.Message.Text)
		if len(cords) < 2 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверно выполненная команда: /check [cords] ...")
			bot.Send(msg)
		} else {
			msgText := "Результат проверки координат:\n"
			for i, cord := range cords {
				if i == 0 {
					continue
				}
				reserverUser, err := db.GetUserByCoords(cord)
				var reserverNick string
				if err != nil {
					reserverNick = "пусто"
				} else {
					reserverNick = reserverUser.Nickname
				}
				msgText += fmt.Sprintf("%s - %s\n", cord, reserverNick)
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
			bot.Send(msg)
		}
	}
	return nil
}

func (n *Handler) Clear(bot *tgbotapi.BotAPI, update tgbotapi.Update, psql *sql.DB, logger *zap.Logger) error {
	db := database.NewDBHandler(psql, logger)
	user, err := db.GetUser(update.Message.From.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я тебя не знаю, выполни регистрацию : /register [nickname]")
		bot.Send(msg)
	} else {
		cords := strings.Fields(update.Message.Text)
		if len(cords) < 2 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверно выполненная команда: /clear [cords] ...")
			bot.Send(msg)
		} else {
			msgText := "Результат очистки координат:\n"
			for i, cord := range cords {
				if i == 0 {
					continue
				}
				reserverUser, err := db.GetUserByCoords(cord)
				var result string
				if err != nil {
					result = "было не забронировано"
				} else if reserverUser.Id == user.Id {
					if err = db.DeleteVillage(cord); err == nil {
						result = "разбронировано"
					} else {
						result = "Ошибка"
					}
				} else {
					result = "это чужая бронь"
				}
				msgText += fmt.Sprintf("%s - %s\n", cord, result)
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
			bot.Send(msg)
		}
	}
	return nil
}

func (n *Handler) Reserve(bot *tgbotapi.BotAPI, update tgbotapi.Update, psql *sql.DB, logger *zap.Logger) error {
	db := database.NewDBHandler(psql, logger)
	user, err := db.GetUser(update.Message.From.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я тебя не знаю, выполни регистрацию : /register [nickname]")
		bot.Send(msg)
	} else {
		cords := strings.Fields(update.Message.Text)
		if len(cords) < 2 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверно выполненная команда: /reserve [cords] ...")
			bot.Send(msg)
		} else {
			msgText := "Результат брони координат:\n"
			for i, cord := range cords {
				if i == 0 {
					continue
				}
				reserverUser, err := db.GetUserByCoords(cord)
				var reserverNick string
				if err != nil {
					reserverNick = "Успешно забронировано"
					db.InsertVillage(user.Id, cord)
				} else {
					reserverNick = "Уже забронено - " + reserverUser.Nickname
				}
				msgText += fmt.Sprintf("%s - %s\n", cord, reserverNick)
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
			bot.Send(msg)
		}
	}
	return nil
}

func (n *Handler) Help(bot *tgbotapi.BotAPI, update tgbotapi.Update, psql *sql.DB, logger *zap.Logger) error {
	msgText := "Команды: \n" +
		"/start - запускает отслеживание проебов и вывод уведомлений в чат\n" +
		"Далее везде где квадратные скобочки, их не пишем, меняем то что в них, например - /register freak\n" +
		"\n/register [игровойник] - регистрирует как юзера бота, ник без пробелов одним словом, можно условный\n" +
		"\n/check [координаты] - проверяет брони по введенным координатам, нет проверки на правильность ввода координат, главное без пробелов. Можно несколько координат подряд\n" +
		"\n/reserve [координаты] - резервирует за вами введенные координаты, те же правила что и с /check\n" +
		"\n/clear [координаты] - разбронь введенных координат, только если забронировано вами, иначе нахуй идете\n" +
		"\nох удачи не сломать эту хуету..."
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
	bot.Send(msg)
	return nil
}

func (n *Handler) Stop(psql *sql.DB, logger *zap.Logger) error {
	db := database.NewDBHandler(psql, logger)
	db.OffAllChats()
	return nil
}

func (n *Handler) SendEnnobleInfo(bot *tgbotapi.BotAPI, chatId int64, lastId int64) int64 {
	httpRequest := "https://twhelp.app/api/v2/versions/pl/servers/pl206/ennoblements?limit=500&sort=createdAt%3ADESC"
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
