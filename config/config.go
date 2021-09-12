package config

import (
	"github.com/spf13/viper"
)

// ReadConfig reads config file
func ReadConfig() (cfg Config, err error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")

	err = viper.ReadInConfig()
	if err != nil {
		return cfg, err
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
