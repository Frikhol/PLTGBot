package database

import "time"

type user struct {
	Id            uint64
	Nickname      string
	TgId          string
	Role          string
	ReservedCount uint64
}

type chat struct {
	Id         uint64
	ChatId     uint64
	IsNoticing bool
	LastLostId uint64
	LastGetId  uint64
}

type chatConfig struct {
	Id                uint64
	ChatId            uint64
	UpdateTimeout     uint64
	NoticingLimit     uint64
	ReserveTime       time.Time
	ReserveLimit      uint64
	IsInternalEnnoble bool
	IsReturnNoticing  bool
}

type village struct {
	Id         uint64
	Info       string
	IsReserved bool
	ReserverId uint64
}
