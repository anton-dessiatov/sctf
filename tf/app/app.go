package app

import (
	"github.com/anton-dessiatov/sctf/tf/dal"
	"github.com/anton-dessiatov/sctf/tf/terra"
	"github.com/jinzhu/gorm"
)

type App struct {
	DB    *gorm.DB
	Terra *terra.Terra
}

func New() *App {
	config := readConfig()
	db := dal.MakeDB()
	return &App{
		DB:    db,
		Terra: terra.NewTerra(db, config.Terra.Credentials),
	}
}

var Instance = New()
