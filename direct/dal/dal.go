package dal

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // Default to MySQL dialects
)

func MakeDB() *gorm.DB {
	db, err := gorm.Open("mysql", "root:101202@tcp(127.0.0.1:5558)/direct?parseTime=true")
	if err != nil {
		log.Fatalf("gorm.Open: %w", err)
	}
	db.SingularTable(true)

	return db
}
