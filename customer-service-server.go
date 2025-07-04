package main

import (
	"bufio"
	"encoding/json"
	"exercise/pkg/schema"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Client struct {
	ID       string
	Conn     *websocket.Conn
	Name     string
	LastSeen time.Time
	Messages []schema.Message
	mu       sync.Mutex
}

type ClientManager struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[string]*Client),
	}
}

func (cm *ClientManager) AddClient(client *Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.clients[client.ID] = client
}

func (cm *ClientManager) RemoveClient(clientID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.clients, clientID)
}

func (cm *ClientManager) GetClient(clientID string) (*Client, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	client, exists := cm.clients[clientID]
	return client, exists
}

func (cm *ClientManager) GetAllClients() map[string]*Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[string]*Client)
	for id, client := range cm.clients {
		result[id] = client
	}
	return result
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clientManager     = NewClientManager()
	currentChatClient = "" // 當前正在對話的客戶端 ID
)

func main() {
	go handleServiceInput()

	http.HandleFunc("/customer", handleCustomerWebSocket)

	log.Println("🎧 客服系統啟動在 :8899")
	log.Println("客戶可以連接到: ws://localhost:8899/customer")
	log.Println("\n客服指令:")
	log.Println("  list          - 查看所有客戶")
	log.Println("  chat <客戶ID>  - 切換到指定客戶對話")
	log.Println("  history <客戶ID> - 查看客戶對話記錄")
	log.Println("  quit          - 關閉系統")
	log.Println("\n等待客戶連接...")

	log.Fatal(http.ListenAndServe(":8899", nil))
}

func handleCustomerWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket 升級失敗: %v", err)
		return
	}

	// 生成客戶 ID
	clientID := fmt.Sprintf("C%d", time.Now().Unix()%10000)

	client := &Client{
		ID:       clientID,
		Conn:     conn,
		Name:     fmt.Sprintf("客戶-%s", clientID),
		LastSeen: time.Now(),
		Messages: make([]schema.Message, 0),
	}

	clientManager.AddClient(client)
	log.Printf("🟢 新客戶連接: %s (%s)", client.Name, client.ID)

	defer func() {
		clientManager.RemoveClient(clientID)
		conn.Close()
		log.Printf("🔴 客戶離線: %s (%s)", client.Name, client.ID)
	}()

	// 發送歡迎訊息
	welcomeMsg := schema.Message{
		From:      "系統",
		Content:   fmt.Sprintf("歡迎！您的客戶編號是 %s，客服將為您服務", clientID),
		Timestamp: time.Now(),
		Type:      "system",
	}
	sendMessageToClient(client, welcomeMsg)

	// 監聽客戶訊息
	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}

		message := schema.Message{
			From:      client.Name,
			Content:   string(messageBytes),
			Timestamp: time.Now(),
			Type:      "client",
		}

		// 儲存訊息
		client.mu.Lock()
		client.Messages = append(client.Messages, message)
		client.LastSeen = time.Now()
		client.mu.Unlock()

		// 顯示客戶訊息
		log.Printf("💬 [%s] %s: %s", client.ID, client.Name, message.Content)

		// 如果當前正在與此客戶對話，顯示提示
		if currentChatClient == client.ID {
			fmt.Printf("\n📨 [%s] %s: %s\n", client.ID, client.Name, message.Content)
			fmt.Print("回覆> ")
		}
	}
}

func sendMessageToClient(client *Client, message schema.Message) {
	messageBytes, _ := json.Marshal(message)
	client.Conn.WriteMessage(websocket.TextMessage, messageBytes)
}

func handleServiceInput() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.SplitN(input, " ", 2)
		command := parts[0]

		switch command {
		case "quit":
			log.Println("系統關閉...")
			os.Exit(0)

		case "list":
			listClients()

		case "chat":
			if len(parts) < 2 {
				fmt.Println("用法: chat <客戶ID>")
				continue
			}
			startChat(parts[1])

		case "history":
			if len(parts) < 2 {
				fmt.Println("用法: history <客戶ID>")
				continue
			}
			showHistory(parts[1])

		default:
			// 如果當前在對話模式，發送訊息給選定的客戶
			if currentChatClient != "" {
				sendMessageToCurrentClient(input)
			} else {
				fmt.Println("未知指令。輸入 'list' 查看客戶，'chat <客戶ID>' 開始對話")
			}
		}
	}
}

func listClients() {
	clients := clientManager.GetAllClients()
	if len(clients) == 0 {
		fmt.Println("📭 目前沒有客戶在線")
		return
	}

	fmt.Println("\n📋 在線客戶列表:")
	fmt.Println("ID\t客戶名稱\t\t最後活動\t\t訊息數")
	fmt.Println("-----------------------------------------------------------")

	for _, client := range clients {
		client.mu.Lock()
		messageCount := len(client.Messages)
		lastSeen := client.LastSeen.Format("15:04:05")
		client.mu.Unlock()

		status := ""
		if currentChatClient == client.ID {
			status = " [當前對話]"
		}

		fmt.Printf("%s\t%s\t\t%s\t\t%d%s\n",
			client.ID, client.Name, lastSeen, messageCount, status)
	}
	fmt.Println()
}

func startChat(clientID string) {
	client, exists := clientManager.GetClient(clientID)
	if !exists {
		fmt.Printf("❌ 客戶 %s 不存在\n", clientID)
		return
	}

	currentChatClient = clientID
	fmt.Printf("💬 開始與 %s (%s) 對話\n", client.Name, clientID)
	fmt.Println("輸入訊息直接發送，輸入 'end' 結束對話，'list' 查看客戶列表")
	fmt.Print("回覆> ")
}

func sendMessageToCurrentClient(content string) {
	if content == "end" {
		fmt.Printf("結束與客戶 %s 的對話\n", currentChatClient)
		currentChatClient = ""
		return
	}

	client, exists := clientManager.GetClient(currentChatClient)
	if !exists {
		fmt.Printf("❌ 客戶 %s 已離線\n", currentChatClient)
		currentChatClient = ""
		return
	}

	message := schema.Message{
		From:      "客服",
		Content:   content,
		Timestamp: time.Now(),
		Type:      "service",
	}

	// 儲存訊息
	client.mu.Lock()
	client.Messages = append(client.Messages, message)
	client.mu.Unlock()

	// 發送給客戶
	sendMessageToClient(client, message)

	fmt.Printf("✅ 已發送給 %s: %s\n", client.Name, content)
	fmt.Print("回覆> ")
}

func showHistory(clientID string) {
	client, exists := clientManager.GetClient(clientID)
	if !exists {
		fmt.Printf("❌ 客戶 %s 不存在\n", clientID)
		return
	}

	client.mu.Lock()
	messages := make([]schema.Message, len(client.Messages))
	copy(messages, client.Messages)
	client.mu.Unlock()

	if len(messages) == 0 {
		fmt.Printf("📭 客戶 %s 還沒有對話記錄\n", clientID)
		return
	}

	fmt.Printf("\n📜 客戶 %s (%s) 的對話記錄:\n", client.Name, clientID)
	fmt.Println("-------------------------------------------")

	for _, msg := range messages {
		timeStr := msg.Timestamp.Format("15:04:05")
		fmt.Printf("[%s] %s: %s\n", timeStr, msg.From, msg.Content)
	}
	fmt.Println("-------------------------------------------\n")
}
