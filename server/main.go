package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[string]*websocket.Conn)
var mu sync.Mutex

func main() {
	e := echo.New()
	e.GET("/ws", func(c echo.Context) error {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		mu.Lock()
		clientId := uuid.New().String()
		clients[clientId] = ws
		mu.Unlock()
		fmt.Println("clients:", clients)

		defer func() {
			mu.Lock()
			delete(clients, clientId)
			ws.Close()
			fmt.Println("client", clientId, "disconnected")
			mu.Unlock()
		}()

		for {
			// read message
			_, message, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					fmt.Println("client closed connection:", err.Error())
				} else {
					fmt.Println("error reading message")
				}
				break
			}
			fmt.Println("message:", clientId, string(message))

			// reply message
			for id, conn := range clients {
				if id != clientId {
					err = conn.WriteMessage(websocket.TextMessage, message)
					if err != nil {
						fmt.Println("ws.WriteMessage:", err)
						break
					}
				}
			}
		}

		return nil
	})
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
