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
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Set up ping/pong for keepalive
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
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
					if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
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

			// Handle ping
			if msg.Type == "ping" {
				response := WebSocketMessage{
					ID:        msg.ID,
					Type:      "pong",
					Channel:   "system",
					Payload:   nil,
					Timestamp: time.Now().UTC().Format(time.RFC3339),
				}
				responseBytes, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, responseBytes)
				continue
			}

			// Handle subscribe/unsubscribe
			if msg.Type == "subscribe" || msg.Type == "unsubscribe" {
				response := WebSocketMessage{
					ID:        msg.ID,
					Type:      "ack",
					Channel:   msg.Channel,
					Payload:   map[string]string{"status": "ok"},
					Timestamp: time.Now().UTC().Format(time.RFC3339),
				}
				responseBytes, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, responseBytes)
			}
		}

		close(done)
		wg.Wait()
	}
}
