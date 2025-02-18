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
	GetLastChatId() uint64
	InsertChat(chat Chat) error
	UpdateChat(chat Chat) error
}

type User struct {
	Id            uint64
	Nickname      string
	TgId          string
	Role          string
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
	Info       string
	IsReserved bool
	ReserverId uint64
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
	var res uint64
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
