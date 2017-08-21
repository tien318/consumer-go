package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Load config from json file
func Load() {
	fmt.Println("Load Config")

	viper.SetConfigName("config")
	viper.AddConfigPath("../../")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}
}
