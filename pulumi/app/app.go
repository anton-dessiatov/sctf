package app

import (
	"github.com/anton-dessiatov/sctf/pulumi/dal"

	"github.com/jinzhu/gorm"
)

type App struct {
	DB *gorm.DB
}

func New() *App {
	// config := readConfig()
	db := dal.MakeDB()

	return &App{
		DB: db,
	}
}

var Instance = New()
