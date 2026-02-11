package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	AppName string       `mapstructure:"app_name"`
	Server  ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
	Http HttpConfig `mapstructure:"http"`
	Grpc GrpcConfig `mapstructure:"grpc"`
}

type HttpConfig struct {
	Port           string `mapstructure:"port"`
	ReadTimeout    int64  `mapstructure:"readtimeout"`
	WriteTimeout   int64  `mapstructure:"writetimeout"`
	IdleTimeout    int64  `mapstructure:"idletimeout"`
	MaxHeaderBytes int    `mapstructure:"maxheaderbytes"`
}

type GrpcConfig struct {
}

func Initialize() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config/")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Error Reading config. Err:", err)
	}

	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal("Unable to unmarshall configs into config struct")
	}

	return &config

	// fmt.Println("Config app_name:", config.AppName)
	// fmt.Println("Config server http:")
	// fmt.Println("Port:", config.Server.Http.Port)
	// fmt.Println("Readtimeout:", config.Server.Http.ReadTimeout)
	// fmt.Println("Writetimeout:", config.Server.Http.WriteTimeout)
	// fmt.Println("Idletimeout:", config.Server.Http.IdleTimeout)
}

func SetDefaults() {
	viper.SetDefault("app_name", "Neo's api server")
	viper.SetDefault("server.http.port", "9000")
	viper.SetDefault("server.http.idletimeout", 15)
	viper.SetDefault("server.http.readtimeout", 15)
	viper.SetDefault("server.http.writetimeout", 15)
	viper.SetDefault("server.http.maxheaderbytes", 1024)
}
