package api

import (
	"gniffer/gniffer"

	"github.com/gin-gonic/gin"
)

func InitRouter() {
	r := gin.Default()

	card := r.Group("/card")
	{
		card.GET("/list", getNetCardsList)
		card.GET("/listen")
	}

	r.Run()
}

func getNetCardsList(c *gin.Context) {
	cards := gniffer.GetNetCards()
	c.JSON(200, gin.H{
		"cards": cards,
	})
}
