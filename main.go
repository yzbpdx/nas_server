package main

import (
	"nas_server/logs"
	"nas_server/router"
)

func main() {
	logger := logs.ServerLog{}
	logger.InitLog() 
	logger.Logger.Infof("logger started!")

	ginRouter := router.RouterInit()

	ginRouter.Run(":9000")
}
