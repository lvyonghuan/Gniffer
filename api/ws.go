package api

import (
	"gniffer/gniffer"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func listen(c *gin.Context) {
	netCard := c.Query("netCard")
	if netCard == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "netCard is required"})
		return
	}
	err := gniffer.Start(netCard)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn, err := upgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	go write(conn)
	go read(conn)
}

func write(conn *websocket.Conn) {
	for {
		select {
		case data := <-gniffer.OutPutChan:
			err := conn.WriteJSON(data)
			if err != nil {
				return
			}
		}
	}
}

func read(conn *websocket.Conn) {
	for {
		_, codeByte, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
		}

		code := string(codeByte)
		switch code {
		case "8":
			//TODO
		}
	}
}
