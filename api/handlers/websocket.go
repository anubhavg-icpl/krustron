// Package handlers - WebSocket handlers for Krustron Dashboard
// Author: Anubhav Gain <anubhavg@infopercept.com>
package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Channel   string      `json:"channel"`
	Payload   interface{} `json:"payload"`
	Timestamp string      `json:"timestamp"`
}

// DashboardWS handles the main WebSocket connection for the dashboard
func DashboardWS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Identity is already enforced by middleware.WSAuth at upgrade time;
		// surface it for logging/audit if needed later.
		userID, _ := c.Get("user_id")
		_ = userID

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// gorilla/websocket allows one concurrent reader and one concurrent
		// writer; the ping goroutine and the read loop both write, so all
		// writes are serialized through writeMu.
		var writeMu sync.Mutex
		writeJSON := func(v interface{}) bool {
			writeMu.Lock()
			defer writeMu.Unlock()
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			return conn.WriteJSON(v) == nil
		}

		// Set up ping/pong for keepalive
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		// Cap inbound frame size; without this a single client can exhaust
		// server memory with an unbounded message.
		conn.SetReadLimit(512 * 1024)
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		var wg sync.WaitGroup
		done := make(chan struct{})

		// Ping ticker
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					writeMu.Lock()
					err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second))
					writeMu.Unlock()
					if err != nil {
						return
					}
				case <-done:
					return
				}
			}
		}()

		// Read loop
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var msg WebSocketMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}

			now := time.Now().UTC().Format(time.RFC3339)

			switch msg.Type {
			case "auth":
				// Frontend sends an auth handshake and waits for auth_success
				// before treating the socket as authenticated. The token was
				// already validated by WSAuth at upgrade; acknowledge it.
				writeJSON(WebSocketMessage{ID: msg.ID, Type: "auth_success", Channel: "system", Timestamp: now})

			case "ping":
				writeJSON(WebSocketMessage{ID: msg.ID, Type: "pong", Channel: "system", Timestamp: now})

			case "subscribe", "unsubscribe":
				writeJSON(WebSocketMessage{ID: msg.ID, Type: "ack", Channel: msg.Channel,
					Payload: map[string]string{"status": "ok"}, Timestamp: now})
			}
		}

		close(done)
		wg.Wait()
	}
}
