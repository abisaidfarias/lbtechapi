package utils

import (
	"log"

	"github.com/spf13/viper"
)

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() {

	viper.SetConfigType("yaml")

	viper.SetConfigFile("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	return
}
