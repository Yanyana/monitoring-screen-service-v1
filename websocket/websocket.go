package websocket

import (
	"fmt"
	"log"
	"net/http"

	"context"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

// Hub menyimpan informasi client WebSocket dan channel komunikasi
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Fungsi untuk menjalankan hub dan mengelola komunikasi
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.clients[conn] = true
		case conn := <-h.unregister:
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
		case message := <-h.broadcast:
			for conn := range h.clients {
				conn.WriteMessage(websocket.TextMessage, message)
			}
		}
	}
}

// Fungsi untuk menangani WebSocket upgrade dan menyambungkan client
func ServeWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Daftarkan client ke hub
	hub.register <- conn

	defer func() {
		hub.unregister <- conn
	}()

	// Terima pesan dari WebSocket (optional: jika ada interaksi lainnya)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		hub.broadcast <- message
	}
}

// Fungsi untuk mendengarkan Redis channel dan mengirim pesan ke WebSocket
func ListenRedisChannel(redisClient *redis.Client, hub *Hub) {
	ctx := context.Background()

	// Subscribe ke channel Redis (result-patient)
	pubsub := redisClient.Subscribe(ctx, "ws-monitoring-patient")
	defer pubsub.Close()

	// Menunggu pesan baru dari Redis
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Println("Error receiving Redis message:", err)
			continue
		}

		// Kirim pesan ke semua WebSocket client yang terhubung
		fmt.Println("Received message from Redis:", msg.Payload)
		hub.broadcast <- []byte(msg.Payload) // Kirim pesan ke WebSocket
	}
}
