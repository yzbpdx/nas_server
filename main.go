package main

import (
	"fmt"
	"os"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	if _, err := os.Stat("HTML/index.html"); err != nil {
		fmt.Println("error")
	}

	r.LoadHTMLGlob("HTML/*")

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Hello, Gin!",
		})
	})

	r.Run(":9000")
}
