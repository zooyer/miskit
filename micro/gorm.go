package micro

import (
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func OpenDB(name, dsn string, config interface{}) (*gorm.DB, error) {
	var dial gorm.Dialector

	switch name {
	case "sqlite", "sqlite3":
		dial = sqlite.Open(dsn)
	case "mysql":
		if conf, ok := config.(mysql.Config); ok {
			dial = mysql.New(conf)
		} else {
			dial = mysql.Open(dsn)
		}
	}

	return gorm.Open(dial)
}
