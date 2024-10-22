package config

import (
	"anti-apt-backend/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Db *gorm.DB

func DBconfig() error {
	dsn := "root:tekken@tcp(127.0.0.1:3306)/antiaptdb?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		// slog.Println("Error connecting to database: ", err)
		return err
	}

	err = db.AutoMigrate(
		&model.FileOnDemand{},
		&model.UrlOnDemand{},
		&model.TaskLiveAnalysisTable{},
		&model.TaskFinishedTable{},
		&model.TaskDuplicateTable{},
		&model.AuditTable{},
		&model.FileHashes{},
	)
	if err != nil {
		// slog.Println("Error migrating database: ", err)
		return err
	}

	Db = db
	// slog.Println("Connected to database")
	return nil
}
