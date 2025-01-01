package api

import (
	"gniffer/gniffer"

	"github.com/gin-gonic/gin"
)

func InitRouter() {
	r := gin.Default()
	//r.Use(gin.Logger())
	r.Use(corsMiddleware())

	r.GET("/list", getNetCardsList)
	r.GET("/listen", listen)
	r.GET("/stop", stop)
	r.GET("/require", require)
	r.GET("/setFilter", setFilter)
	r.GET("/setSorter", setSorter)

	r.Run()
}

func getNetCardsList(c *gin.Context) {
	cards := gniffer.GetCards()
	c.JSON(200, gin.H{
		"cards": cards,
	})
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
