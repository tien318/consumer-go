package config

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Load config from json file
func Load() {
	log.Info("Load Config")

	viper.SetConfigName("config")
	viper.AddConfigPath("../../")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}
}
