package db

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var keymap = map[string]string{}

func GoDotEnvVariable(key string) string {
	if val, ok := keymap[key]; ok {
		return val
	}
	err := godotenv.Load()
	if err != nil {
		logrus.Error(err)
		return ""
	}
	keymap[key] = os.Getenv(key)
	return keymap[key]
}
