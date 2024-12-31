package api

import (
	"gniffer/gniffer"
	"strconv"

	"github.com/gin-gonic/gin"
)

func listen(c *gin.Context) {
	netCard := c.Query("netCard")
	if netCard == "" {
		c.JSON(400, gin.H{
			"error": "netCard is required",
		})
		return
	}

	err := gniffer.Start(netCard)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "start listening",
	})
}

func stop(c *gin.Context) {
	netCard := c.Query("netCard")
	if netCard == "" {
		c.JSON(400, gin.H{
			"error": "netCard is required",
		})
		return
	}

	gniffer.Stop(netCard)

	c.JSON(200, gin.H{
		"message": "stop listening",
	})
}

func require(c *gin.Context) {
	startIndexString := c.Query("start")
	if startIndexString == "" {
		c.JSON(400, gin.H{
			"error": "startIndex is required",
		})
		return
	}

	endIndexString := c.Query("end")
	if endIndexString == "" {
		c.JSON(400, gin.H{
			"error": "endIndex is required",
		})
		return
	}

	startIndex, err := strconv.Atoi(startIndexString)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "startIndex must be a number",
		})
		return
	}

	endIndex, err := strconv.Atoi(endIndexString)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "endIndex must be a number",
		})
		return
	}

	data, index := gniffer.GetDataByIndex(startIndex, endIndex)

	//log.Println(data[0].ID)

	c.JSON(200, gin.H{
		"data":  data,
		"index": index,
	})
}

func setFilter(c *gin.Context) {
	filter := c.Query("filter")
	if filter == "" {
		c.JSON(400, gin.H{
			"error": "filter is required",
		})
		return
	}

	gniffer.FilterType = filter

	c.JSON(200, gin.H{
		"message": "set filter",
	})
}

func setSorter(c *gin.Context) {
	sorter := c.Query("sorter")
	if sorter == "" {
		c.JSON(400, gin.H{
			"error": "sorter is required",
		})
		return
	}

	gniffer.SortType = sorter

	c.JSON(200, gin.H{
		"message": "set sorter",
	})
}
