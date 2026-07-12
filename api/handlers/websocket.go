// Package handlers - WebSocket handlers for Krustron Dashboard
// Author: Anubhav Gain <anubhavg@infopercept.com>
package handlers

import (
	"github.com/anubhavg-icpl/krustron/pkg/websocket"
	"github.com/gin-gonic/gin"
)

// DashboardWS upgrades the dashboard WebSocket and hands the connection to the
// real-time hub, which drives ping/pong, subscribe/unsubscribe, auth handshake,
// and event broadcast. Authentication was already enforced by middleware.WSAuth
// at upgrade time.
func DashboardWS(hub *websocket.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		uid, _ := userID.(string)
		hub.HandleWebSocket(c.Writer, c.Request, uid)
	}
}
