package main

import (
	"nas_server/logs"
	"nas_server/router"
)

func main() {
	logs.GetInstance().Infof("logger started!")

	ginRouter := router.RouterInit()

	ginRouter.Run(":9000")
}
