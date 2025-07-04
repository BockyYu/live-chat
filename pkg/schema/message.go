package schema

import (
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type Message struct {
	From      string    `json:"from"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

type Client struct {
	ID       string
	Conn     *websocket.Conn
	Name     string
	LastSeen time.Time
	Messages []Message
	mu       sync.Mutex
}

// ... 其他原本在 types.go 的定義
