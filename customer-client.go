package main

import (
	"bufio"
	"encoding/json"
	"exercise/pkg/schema"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

func main() {
	// 連接到客服系統
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8899/customer", nil)
	if err != nil {
		log.Fatal("連接失敗:", err)
	}
	defer conn.Close()

	log.Println("🔗 已連接到客服系統")
	log.Println("請輸入您的問題，客服會盡快回覆您 (輸入 'quit' 結束)")

	// 設定中斷信號處理
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	// 啟動接收訊息的 goroutine
	go func() {
		defer close(done)
		for {
			_, messageBytes, err := conn.ReadMessage()
			if err != nil {
				log.Println("連接中斷:", err)
				return
			}

			var message schema.Message
			if err := json.Unmarshal(messageBytes, &message); err != nil {
				log.Printf("解析訊息失敗: %v", err)
				continue
			}

			// 清除當前行並顯示收到的訊息
			fmt.Print("\r\033[K")
			timeStr := message.Timestamp.Format("15:04:05")

			switch message.Type {
			case "system":
				fmt.Printf("🔔 [%s] %s\n", timeStr, message.Content)
			case "service":
				fmt.Printf("🎧 [%s] 客服: %s\n", timeStr, message.Content)
			}

			fmt.Print("您> ")
		}
	}()

	// 啟動發送訊息的 goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("您> ")
			if !scanner.Scan() {
				break
			}

			text := strings.TrimSpace(scanner.Text())
			if text == "quit" {
				log.Println("結束對話...")
				conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}

			if text != "" {
				err := conn.WriteMessage(websocket.TextMessage, []byte(text))
				if err != nil {
					log.Println("發送失敗:", err)
					return
				}
			}
		}
	}()

	// 等待程式結束
	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("\n正在離開...")

			err := conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("關閉連接失敗:", err)
				return
			}

			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
