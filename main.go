package main

import (
	config "nas_server/conf"
	"nas_server/gorm"
	"nas_server/logs"
	"nas_server/redis"
	"nas_server/router"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logs.GetInstance().Logger.Infof("logger started!")
	config.InitServerConfig("conf/server.yaml")
	config := config.GetServerConfig()
	logs.GetInstance().Logger.Infof("config %+v", config)
	ginRouter := router.RouterInit(&config.Server, config.Docker.RegistryPort)
	redis.RedisInit(&config.Redis)
	gorm.MysqlInit(&config.MySQL)

	channel := make(chan os.Signal)
	signal.Notify(channel, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)
	go func() {
		for {
			select {
			case s := <-channel:
				logs.GetInstance().Logger.Infof("sql closed")
				gorm.GetClient().Close()
				logs.GetInstance().Logger.Infof("redis closed")
				redis.GetClient().Close()
				logs.GetInstance().Logger.Infof("log file closed")
				logs.GetInstance().Logger.Infof("server gracefully shutdown %v", s)
				logs.GetInstance().CloseLogFile()
				os.Exit(0)
			}
		}
	}()

	ginRouter.Run(config.Server.Listen)
}
