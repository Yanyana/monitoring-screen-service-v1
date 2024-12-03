package sse

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// SSEClient merepresentasikan klien SSE
type SSEClient struct {
	Chan chan string
}

type SSEHub struct {
	mu      sync.Mutex
	clients map[*SSEClient]bool
}

// NewSSEHub membuat hub SSE baru
func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[*SSEClient]bool),
	}
}

// AddClient menambahkan klien baru ke hub
func (h *SSEHub) AddClient(client *SSEClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client] = true
}

// RemoveClient menghapus klien dari hub
func (h *SSEHub) RemoveClient(client *SSEClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, client)
	close(client.Chan)
}

// Broadcast mengirim pesan ke semua klien
func (h *SSEHub) Broadcast(message string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.clients {
		client.Chan <- message
	}
}

// SSEHandler menangani koneksi SSE dari klien
func SSEHandler(hub *SSEHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Setup header untuk SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		client := &SSEClient{Chan: make(chan string)}
		hub.AddClient(client)
		defer hub.RemoveClient(client)

		log.Println("Client connected to SSE")

		// Kirim data ke klien
		for msg := range client.Chan {
			log.Printf("Sending message to client: %s", msg)
			fmt.Fprint(w, msg)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}

		log.Println("Client disconnected from SSE")
	}
}

// ListenRedisChannel mendengarkan pesan dari Redis dan mengirimkannya ke hub SSE
func ListenRedisChannel(redisClient *redis.Client, hub *SSEHub) {
	ctx := context.Background()
	pubsub := redisClient.Subscribe(ctx, "monitoring-patient")
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Println("Error receiving Redis message:", err)
			continue
		}

		// Gunakan timestamp Unix sebagai ID
		id := fmt.Sprintf("%d", time.Now().UnixMilli())

		// Formatkan payload dengan benar
		payload := fmt.Sprintf("event: new-regis\nid: %s\ndata: %s\n\n", id, msg.Payload)

		fmt.Println("Received message from Redis:", msg.Payload)
		// Kirim pesan ke semua klien SSE
		hub.Broadcast(payload)
	}
}
