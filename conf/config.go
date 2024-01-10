package config

import (
	"nas_server/logs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HomePath string `yaml:"homePath"`
	ServerFolder string `yaml:"serverFolder"`

	UserName string
	RootFolder string
}

func InitServerConfig(fileName string) {
	configFile, err := os.ReadFile(fileName)
	if err != nil {
		logs.GetInstance().Logger.Errorf("server config %s not found %s", fileName, err)
		return
	}
	err = yaml.Unmarshal(configFile, serverConfig)
	if err != nil {
		logs.GetInstance().Logger.Errorf("yaml unmarshal error %s", err)
	}

	rootFolder, _ := os.UserHomeDir()
	serverConfig.RootFolder = filepath.Join(rootFolder, serverConfig.ServerFolder)
	if _, err := os.Stat(serverConfig.RootFolder); err != nil && os.IsNotExist(err) {
		os.Mkdir(serverConfig.RootFolder, 0777)
	}
}

var serverConfig = &Config{}

func GetServerConfig() *Config {
	return serverConfig
}