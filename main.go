package main

import (
	"nas_server/logs"
	"nas_server/router"
	"nas_server/redis"
)

func main() {
	logs.GetInstance().Infof("logger started!")
	ginRouter := router.RouterInit()
	redis.RedisInit("localhost:6379", "", 0)

	ginRouter.Run(":9000")
}
