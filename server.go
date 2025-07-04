package main

//
//import (
//	"bufio"
//	"github.com/gorilla/websocket"
//	"log"
//	"net/http"
//	"os"
//)
//
//var (
//	upgrader = websocket.Upgrader{
//		CheckOrigin: func(r *http.Request) bool { return true },
//	}
//	// 儲存當前連接的客戶端
//	clientConn *websocket.Conn
//	// 用於伺服器輸入的 channel
//	serverInput = make(chan string)
//)
//
//func main() {
//	// 啟動伺服器輸入的 goroutine
//	go handleServerInput()
//
//	http.HandleFunc("/echo", handleWebSocket)
//
//	log.Println("WebSocket 伺服器啟動在 :8899")
//	log.Println("請在另一個終端執行客戶端")
//	log.Println("你可以在這裡輸入訊息發送給客戶端:")
//
//	log.Fatal(http.ListenAndServe(":8899", nil))
//}
//
//func handleWebSocket(w http.ResponseWriter, r *http.Request) {
//	conn, err := upgrader.Upgrade(w, r, nil)
//	if err != nil {
//		log.Printf("升級錯誤: %v", err)
//		return
//	}
//	defer func() {
//		clientConn = nil
//		conn.Close()
//		log.Println("客戶端已斷線")
//	}()
//
//	// 設定當前客戶端連接
//	clientConn = conn
//	log.Println("新客戶端已連接")
//
//	// 監聽客戶端訊息
//	for {
//		_, message, err := conn.ReadMessage()
//		if err != nil {
//			log.Printf("讀取訊息錯誤: %v", err)
//			break
//		}
//
//		log.Printf("📨 客戶端: %s", message)
//	}
//}
//
//func handleServerInput() {
//	scanner := bufio.NewScanner(os.Stdin)
//	for scanner.Scan() {
//		text := scanner.Text()
//		if text == "quit" {
//			log.Println("伺服器關閉...")
//			os.Exit(0)
//		}
//
//		if text != "" && clientConn != nil {
//			// 發送訊息給客戶端
//			serverMsg := append([]byte("伺服器: "), []byte(text)...)
//			err := clientConn.WriteMessage(websocket.TextMessage, serverMsg)
//			if err != nil {
//				log.Printf("發送訊息錯誤: %v", err)
//			}
//		}
//	}
//}
