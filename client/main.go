package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

func main() {
	serverUrl := "ws://localhost:8080/ws"
	connection, _, err := websocket.DefaultDialer.Dial(serverUrl, nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer connection.Close()

	done := make(chan struct{})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	message := "message1"
	err = connection.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Fatal("WriteMessage:", err)
	}

	go func() {
		defer close(done)
		for {
			_, response, err := connection.ReadMessage()
			if err != nil {
				log.Fatal("ReadMessage:", err)
				return
			}
			fmt.Println("response:", string(response))
		}
	}()

	<-ctx.Done()
	err = connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		fmt.Println("write close message:", err)
		return
	}

	<-done
}
