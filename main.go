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
	logs.GetInstance().Infof("logger started!")
	ginRouter := router.RouterInit()
	redis.RedisInit("localhost:6379", "", 0)
	gorm.MysqlInit("dyf", "123", "localhost:3306", "nas_server")

	channel := make(chan os.Signal)
	signal.Notify(channel, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)
	go func() {
		for {
			select {
			case s := <-channel:
				logs.GetInstance().Infof("server gracefully shutdown %v", s)
				gorm.GetClient().Close()
				redis.GetClient().Close()
				os.Exit(0)
			}
		}
	}()

	ginRouter.Run(":9000")
}
