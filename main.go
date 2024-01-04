package main

import (
	"nas_server/gorm"
	"nas_server/logs"
	"nas_server/redis"
	"nas_server/router"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	userFolder, _ := os.UserHomeDir()
	rootFolder := userFolder + "/gopath"

	logs.GetInstance().Logger.Infof("logger started!")
	ginRouter := router.RouterInit(rootFolder)
	redis.RedisInit("localhost:6379", "", 0)
	gorm.MysqlInit("dyf", "123", "localhost:3306", "nas_server")

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

	ginRouter.Run(":9000")
}
