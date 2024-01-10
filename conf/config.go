package config

import (
	"nas_server/logs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Redis RedisConfig `yaml:"redis"`
	MySQL MySQLConfig `yaml:"mysql"`
}

type ServerConfig struct {
	HomePath string `yaml:"homePath"`
	ServerFolder string `yaml:"serverFolder"`
	Listen string `yaml:"listen"`

	RootFolder string
}

type RedisConfig struct {
	Addr string `yaml:"addr"`
	Password string `yaml:"password"`
	DB int `yaml:"db"`
}

type MySQLConfig struct {
	Name string `yaml:"name"`
	PassWord string `yaml:"password"`
	Addr string `yaml:"addr"`
	DB string `yaml:"db"`
}

func InitServerConfig(fileName string) {
	configFile, err := os.ReadFile(fileName)
	if err != nil {
		logs.GetInstance().Logger.Errorf("server config %s not found %s", fileName, err)
		return
	}
	err = yaml.Unmarshal(configFile, config)
	if err != nil {
		logs.GetInstance().Logger.Errorf("yaml unmarshal error %s", err)
	}

	rootFolder, _ := os.UserHomeDir()
	config.Server.RootFolder = filepath.Join(rootFolder, config.Server.ServerFolder)
	if _, err := os.Stat(config.Server.RootFolder); err != nil && os.IsNotExist(err) {
		os.Mkdir(config.Server.RootFolder, 0777)
	}
}

var config = &Config{}

func GetServerConfig() *Config {
	return config
}