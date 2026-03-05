package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		GoDotEnvVariable("DB_HOST"),
		GoDotEnvVariable("DB_USER"),
		GoDotEnvVariable("DB_PASSWORD"),
		GoDotEnvVariable("DB_NAME"),
		GoDotEnvVariable("DB_PORT"))
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if err := DB.Exec("SELECT 1").Error; err != nil {
		panic(err)
	}
}
