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
	HomeUrl string `yaml:"homeUrl"`
	ServerFolder string `yaml:"serverFolder"`
	Listen string `yaml:"listen"`
	Share string `yaml:"share"`

	RootFolder string
	ShareFolder string
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
	initFolder(config.Server.RootFolder)
	config.Server.ShareFolder = filepath.Join(rootFolder, config.Server.Share)
	initFolder(config.Server.Share)
}

var config = &Config{}

func GetServerConfig() *Config {
	return config
}

func initFolder(folder string) {
	if _, err := os.Stat(folder); err != nil && os.IsNotExist(err) {
		os.Mkdir(folder, 0777)
	}
}