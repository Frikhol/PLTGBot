package database

import (
	"database/sql"
	"go.uber.org/zap"
	"time"
)

type Handler struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewDBHandler(db *sql.DB, logger *zap.Logger) *Handler {
	return &Handler{db, logger}
}

type DataBase interface {
	GetChat(chatId int64) (Chat, error)
	GetUser(tgId int64) (User, error)
	GetUserByCoords(cord string) (User, error)
	GetLastChatId() uint64
	GetLastUserId() uint64
	InsertChat(chat Chat) error
	InsertUser(user User) error
	InsertVillage(userId uint64, cords string) error
	UpdateChat(chat Chat) error
	OffAllChats()
	DeleteVillage(cord string) error
	UpdateUserName(userId uint64, newName string) error
}

type User struct {
	Id            uint64
	Nickname      string
	TgId          int64
	ReservedCount int64
}

type Chat struct {
	Id         uint64
	ChatId     int64
	IsNoticing bool
	LastLostId int64
	LastGetId  int64
}

type ChatConfig struct {
	Id                uint64
	ChatId            uint64
	UpdateTimeout     time.Time
	NoticingLimit     int64
	ReserveTime       time.Time
	ReserveLimit      int64
	IsInternalEnnoble bool
	IsReturnNoticing  bool
}

type Village struct {
	Id         uint64
	X          int
	Y          int
	IsReserved bool
	ReserverId uint64
}

func (h *Handler) OffAllChats() {
	_, err := h.db.Exec("update chats set is_noticing = false")
	if err != nil {
		h.logger.Info("Failed to update chats", zap.Error(err))
	}
}

func (h *Handler) GetUser(tgId int64) (User, error) {
	row := h.db.QueryRow("select * from users where tg_id=$1", tgId)
	user := User{}
	err := row.Scan(&user.Id, &user.Nickname, &user.TgId, &user.ReservedCount)
	if err != nil {
		h.logger.Info("Failed to load from users table", zap.Error(err))
	}
	return user, err
}

func (h *Handler) GetUserByCoords(cord string) (User, error) {
	row := h.db.QueryRow("select users.id,users.nickname,users.tg_id,users.reserved_count from users join villages on users.id = villages.reserver_id where cords = $1", cord)
	user := User{}
	err := row.Scan(&user.Id, &user.Nickname, &user.TgId, &user.ReservedCount)
	if err != nil {
		h.logger.Info("Failed to load from users table", zap.Error(err))
	}
	return user, err
}

func (h *Handler) GetLastUserId() uint64 {
	row := h.db.QueryRow("select max(id) from users")
	var res uint64 = 0
	err := row.Scan(&res)
	if err != nil {
		h.logger.Info("Failed to select id's from users", zap.Error(err))
	}
	return res
}

func (h *Handler) InsertUser(user User) error {
	_, err := h.db.Exec("insert into users (id,nickname,tg_id,reserved_count) values ($1,$2,$3,$4)", user.Id, user.Nickname, user.TgId, user.ReservedCount)
	if err != nil {
		h.logger.Info("Failed to insert new user", zap.Error(err))
	}
	return err
}

func (h *Handler) InsertVillage(userId uint64, cords string) error {
	_, err := h.db.Exec("insert into villages (cords,is_reserved,reserver_id) values ($1,$2,$3)", cords, true, userId)
	if err != nil {
		h.logger.Info("Failed to insert new village", zap.Error(err))
	}
	return err
}

func (h *Handler) GetChat(chatId int64) (Chat, error) {
	row := h.db.QueryRow("select * from chats where chat_id=$1", chatId)
	chat := Chat{}
	err := row.Scan(&chat.Id, &chat.ChatId, &chat.IsNoticing, &chat.LastLostId, &chat.LastGetId)
	if err != nil {
		h.logger.Info("Failed to load from chats table", zap.Error(err))
	}
	return chat, err
}

func (h *Handler) GetLastChatId() uint64 {
	row := h.db.QueryRow("select max(id) from chats")
	var res uint64 = 0
	err := row.Scan(&res)
	if err != nil {
		h.logger.Info("Failed to select id's from chats", zap.Error(err))
	}
	return res
}

func (h *Handler) InsertChat(chat Chat) error {
	_, err := h.db.Exec("insert into chats (id,chat_id,is_noticing,last_lost_id,last_get_id) values ($1,$2,$3,$4,$5)", chat.Id, chat.ChatId, chat.IsNoticing, chat.LastLostId, chat.LastGetId)
	if err != nil {
		h.logger.Info("Failed to insert new chat", zap.Error(err))
	}
	return err
}

func (h *Handler) UpdateChat(chat Chat) error {
	_, err := h.db.Exec("delete from chats where id =$1", chat.Id)
	if err != nil {
		h.logger.Info("Failed to delete from chats", zap.Error(err))
	}
	h.InsertChat(chat)
	return err
}

func (h *Handler) DeleteVillage(cord string) error {
	_, err := h.db.Exec("delete from villages where cords = $1", cord)
	if err != nil {
		h.logger.Info("Failed to delete from villages", zap.Error(err))
	}
	return err
}

func (h *Handler) UpdateUserName(userId uint64, newName string) error {
	_, err := h.db.Exec("update users set nickname = $1 where id = $2", newName, userId)
	if err != nil {
		h.logger.Info("Failed to update name", zap.Error(err))
	}
	return err
}
