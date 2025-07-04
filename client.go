package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

func main() {
	// 連接到 WebSocket 伺服器
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8899/echo", nil)
	if err != nil {
		log.Fatal("連接失敗:", err)
	}
	defer conn.Close()

	log.Println("已連接到伺服器")
	log.Println("你可以開始輸入訊息 (輸入 'quit' 結束):")

	// 設定中斷信號處理
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// 用來通知程式結束的 channel
	done := make(chan struct{})

	// 啟動讀取訊息的 goroutine (只顯示伺服器的訊息)
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("讀取錯誤:", err)
				return
			}

			// 清除當前行並顯示收到的訊息
			fmt.Print("\r\033[K") // 清除當前行
			log.Printf("📨 %s", message)
			fmt.Print("請輸入: ") // 重新顯示輸入提示
		}
	}()

	// 啟動寫入訊息的 goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("請輸入: ")
			if !scanner.Scan() {
				break
			}

			text := strings.TrimSpace(scanner.Text())
			if text == "quit" {
				log.Println("結束程式...")
				// 發送關閉訊息
				conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}

			if text != "" {
				err := conn.WriteMessage(websocket.TextMessage, []byte(text))
				if err != nil {
					log.Println("發送錯誤:", err)
					return
				}
				// 不在這裡顯示自己發送的訊息
			}
		}
	}()

	// 等待程式結束
	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("\n接收到中斷信號")

			// 優雅地關閉連接
			err := conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("關閉訊息發送失敗:", err)
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
