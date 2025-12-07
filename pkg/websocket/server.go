// Package websocket provides WebSocket server for real-time communication
// Author: Anubhav Gain <anubhavg@infopercept.com>
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// System messages
	MessageTypeConnect    MessageType = "connect"
	MessageTypeDisconnect MessageType = "disconnect"
	MessageTypePing       MessageType = "ping"
	MessageTypePong       MessageType = "pong"
	MessageTypeError      MessageType = "error"

	// Cluster messages
	MessageTypeClusterStatus  MessageType = "cluster.status"
	MessageTypeClusterHealth  MessageType = "cluster.health"
	MessageTypeClusterMetrics MessageType = "cluster.metrics"
	MessageTypeClusterEvent   MessageType = "cluster.event"

	// Application messages
	MessageTypeAppStatus   MessageType = "app.status"
	MessageTypeAppSync     MessageType = "app.sync"
	MessageTypeAppHealth   MessageType = "app.health"
	MessageTypeAppEvent    MessageType = "app.event"

	// Pipeline messages
	MessageTypePipelineStatus MessageType = "pipeline.status"
	MessageTypePipelineLog    MessageType = "pipeline.log"
	MessageTypePipelineStage  MessageType = "pipeline.stage"

	// Resource messages
	MessageTypePodStatus    MessageType = "pod.status"
	MessageTypePodLogs      MessageType = "pod.logs"
	MessageTypeNodeStatus   MessageType = "node.status"
	MessageTypeNodeMetrics  MessageType = "node.metrics"

	// Alert messages
	MessageTypeAlert       MessageType = "alert"
	MessageTypeAlertAck    MessageType = "alert.ack"
	MessageTypeAlertResolve MessageType = "alert.resolve"

	// Cost messages
	MessageTypeCostUpdate MessageType = "cost.update"
	MessageTypeCostAlert  MessageType = "cost.alert"

	// AI messages
	MessageTypeAIResponse MessageType = "ai.response"
	MessageTypeAIStream   MessageType = "ai.stream"
)

// Message represents a WebSocket message
type Message struct {
	ID        string                 `json:"id"`
	Type      MessageType            `json:"type"`
	Channel   string                 `json:"channel,omitempty"`
	Data      interface{}            `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID            string
	UserID        string
	Conn          *websocket.Conn
	Hub           *Hub
	Send          chan *Message
	Subscriptions map[string]bool
	mu            sync.RWMutex
	lastPing      time.Time
	metadata      map[string]interface{}
}

// Hub manages all WebSocket connections
type Hub struct {
	clients       map[string]*Client
	channels      map[string]map[string]*Client
	broadcast     chan *Message
	register      chan *Client
	unregister    chan *Client
	subscribe     chan *Subscription
	unsubscribe   chan *Subscription
	mu            sync.RWMutex
	logger        *zap.Logger
	messageBuffer []*Message
	bufferSize    int
}

// Subscription represents a channel subscription request
type Subscription struct {
	Client  *Client
	Channel string
}

// Config holds WebSocket server configuration
type Config struct {
	PingInterval    time.Duration
	PongTimeout     time.Duration
	WriteTimeout    time.Duration
	ReadTimeout     time.Duration
	MaxMessageSize  int64
	BufferSize      int
	EnableCompression bool
}

// DefaultConfig returns default WebSocket configuration
func DefaultConfig() *Config {
	return &Config{
		PingInterval:    30 * time.Second,
		PongTimeout:     60 * time.Second,
		WriteTimeout:    10 * time.Second,
		ReadTimeout:     60 * time.Second,
		MaxMessageSize:  512 * 1024, // 512KB
		BufferSize:      1000,
		EnableCompression: true,
	}
}

// NewHub creates a new WebSocket hub
func NewHub(logger *zap.Logger, config *Config) *Hub {
	if config == nil {
		config = DefaultConfig()
	}

	return &Hub{
		clients:       make(map[string]*Client),
		channels:      make(map[string]map[string]*Client),
		broadcast:     make(chan *Message, 256),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		subscribe:     make(chan *Subscription, 64),
		unsubscribe:   make(chan *Subscription, 64),
		logger:        logger,
		messageBuffer: make([]*Message, 0, config.BufferSize),
		bufferSize:    config.BufferSize,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("Hub shutting down")
			h.closeAllClients()
			return

		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case sub := <-h.subscribe:
			h.subscribeClient(sub)

		case sub := <-h.unsubscribe:
			h.unsubscribeClient(sub)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-ticker.C:
			h.checkConnections()
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client
	h.logger.Info("Client connected",
		zap.String("client_id", client.ID),
		zap.String("user_id", client.UserID),
	)

	// Send connection confirmation
	client.Send <- &Message{
		ID:        uuid.New().String(),
		Type:      MessageTypeConnect,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"client_id": client.ID,
			"message":   "Connected to Krustron WebSocket",
		},
	}

	// Send buffered messages
	for _, msg := range h.messageBuffer {
		select {
		case client.Send <- msg:
		default:
		}
	}
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)
		close(client.Send)

		// Remove from all channels
		for channel := range client.Subscriptions {
			if clients, ok := h.channels[channel]; ok {
				delete(clients, client.ID)
				if len(clients) == 0 {
					delete(h.channels, channel)
				}
			}
		}

		h.logger.Info("Client disconnected",
			zap.String("client_id", client.ID),
			zap.String("user_id", client.UserID),
		)
	}
}

func (h *Hub) subscribeClient(sub *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.channels[sub.Channel]; !ok {
		h.channels[sub.Channel] = make(map[string]*Client)
	}
	h.channels[sub.Channel][sub.Client.ID] = sub.Client

	sub.Client.mu.Lock()
	sub.Client.Subscriptions[sub.Channel] = true
	sub.Client.mu.Unlock()

	h.logger.Debug("Client subscribed to channel",
		zap.String("client_id", sub.Client.ID),
		zap.String("channel", sub.Channel),
	)
}

func (h *Hub) unsubscribeClient(sub *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.channels[sub.Channel]; ok {
		delete(clients, sub.Client.ID)
		if len(clients) == 0 {
			delete(h.channels, sub.Channel)
		}
	}

	sub.Client.mu.Lock()
	delete(sub.Client.Subscriptions, sub.Channel)
	sub.Client.mu.Unlock()
}

func (h *Hub) broadcastMessage(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Add to buffer
	if len(h.messageBuffer) >= h.bufferSize {
		h.messageBuffer = h.messageBuffer[1:]
	}
	h.messageBuffer = append(h.messageBuffer, message)

	// Send to channel subscribers if channel specified
	if message.Channel != "" {
		if clients, ok := h.channels[message.Channel]; ok {
			for _, client := range clients {
				select {
				case client.Send <- message:
				default:
					h.logger.Warn("Client send buffer full",
						zap.String("client_id", client.ID),
					)
				}
			}
		}
		return
	}

	// Broadcast to all clients
	for _, client := range h.clients {
		select {
		case client.Send <- message:
		default:
			h.logger.Warn("Client send buffer full",
				zap.String("client_id", client.ID),
			)
		}
	}
}

func (h *Hub) checkConnections() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if time.Since(client.lastPing) > 90*time.Second {
			h.logger.Warn("Client connection stale",
				zap.String("client_id", client.ID),
			)
		}
	}
}

func (h *Hub) closeAllClients() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, client := range h.clients {
		close(client.Send)
		client.Conn.Close()
	}
	h.clients = make(map[string]*Client)
	h.channels = make(map[string]map[string]*Client)
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(message *Message) {
	if message.ID == "" {
		message.ID = uuid.New().String()
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	h.broadcast <- message
}

// BroadcastToChannel sends a message to a specific channel
func (h *Hub) BroadcastToChannel(channel string, message *Message) {
	message.Channel = channel
	h.Broadcast(message)
}

// SendToClient sends a message to a specific client
func (h *Hub) SendToClient(clientID string, message *Message) error {
	h.mu.RLock()
	client, ok := h.clients[clientID]
	h.mu.RUnlock()

	if !ok {
		return fmt.Errorf("client not found: %s", clientID)
	}

	if message.ID == "" {
		message.ID = uuid.New().String()
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	select {
	case client.Send <- message:
		return nil
	default:
		return fmt.Errorf("client send buffer full: %s", clientID)
	}
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetChannelCount returns the number of active channels
func (h *Hub) GetChannelCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.channels)
}

// GetClientsByChannel returns clients subscribed to a channel
func (h *Hub) GetClientsByChannel(channel string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clientIDs []string
	if clients, ok := h.channels[channel]; ok {
		for id := range clients {
			clientIDs = append(clientIDs, id)
		}
	}
	return clientIDs
}

// HandleWebSocket handles WebSocket upgrade and client connection
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request, userID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}

	client := &Client{
		ID:            uuid.New().String(),
		UserID:        userID,
		Conn:          conn,
		Hub:           h,
		Send:          make(chan *Message, 256),
		Subscriptions: make(map[string]bool),
		lastPing:      time.Now(),
		metadata:      make(map[string]interface{}),
	}

	h.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.lastPing = time.Now()
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.logger.Error("WebSocket read error",
					zap.String("client_id", c.ID),
					zap.Error(err),
				)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.Hub.logger.Warn("Invalid message format",
				zap.String("client_id", c.ID),
				zap.Error(err),
			)
			continue
		}

		c.handleMessage(&msg)
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				c.Hub.logger.Error("Message marshal error",
					zap.String("client_id", c.ID),
					zap.Error(err),
				)
				continue
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				c.Hub.logger.Error("WebSocket write error",
					zap.String("client_id", c.ID),
					zap.Error(err),
				)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from the client
func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case MessageTypePing:
		c.lastPing = time.Now()
		c.Send <- &Message{
			ID:        uuid.New().String(),
			Type:      MessageTypePong,
			Timestamp: time.Now(),
		}

	case "subscribe":
		if channel, ok := msg.Data.(map[string]interface{})["channel"].(string); ok {
			c.Hub.subscribe <- &Subscription{
				Client:  c,
				Channel: channel,
			}
		}

	case "unsubscribe":
		if channel, ok := msg.Data.(map[string]interface{})["channel"].(string); ok {
			c.Hub.unsubscribe <- &Subscription{
				Client:  c,
				Channel: channel,
			}
		}

	default:
		c.Hub.logger.Debug("Received message",
			zap.String("client_id", c.ID),
			zap.String("type", string(msg.Type)),
		)
	}
}

// Subscribe subscribes the client to a channel
func (c *Client) Subscribe(channel string) {
	c.Hub.subscribe <- &Subscription{
		Client:  c,
		Channel: channel,
	}
}

// Unsubscribe unsubscribes the client from a channel
func (c *Client) Unsubscribe(channel string) {
	c.Hub.unsubscribe <- &Subscription{
		Client:  c,
		Channel: channel,
	}
}

// EventEmitter provides convenience methods for emitting events
type EventEmitter struct {
	hub *Hub
}

// NewEventEmitter creates a new event emitter
func NewEventEmitter(hub *Hub) *EventEmitter {
	return &EventEmitter{hub: hub}
}

// EmitClusterStatus emits cluster status update
func (e *EventEmitter) EmitClusterStatus(clusterID string, status interface{}) {
	e.hub.BroadcastToChannel("cluster:"+clusterID, &Message{
		Type: MessageTypeClusterStatus,
		Data: status,
	})
}

// EmitClusterHealth emits cluster health update
func (e *EventEmitter) EmitClusterHealth(clusterID string, health interface{}) {
	e.hub.BroadcastToChannel("cluster:"+clusterID, &Message{
		Type: MessageTypeClusterHealth,
		Data: health,
	})
}

// EmitClusterMetrics emits cluster metrics
func (e *EventEmitter) EmitClusterMetrics(clusterID string, metrics interface{}) {
	e.hub.BroadcastToChannel("cluster:"+clusterID, &Message{
		Type: MessageTypeClusterMetrics,
		Data: metrics,
	})
}

// EmitAppStatus emits application status update
func (e *EventEmitter) EmitAppStatus(appID string, status interface{}) {
	e.hub.BroadcastToChannel("app:"+appID, &Message{
		Type: MessageTypeAppStatus,
		Data: status,
	})
}

// EmitAppSync emits application sync event
func (e *EventEmitter) EmitAppSync(appID string, sync interface{}) {
	e.hub.BroadcastToChannel("app:"+appID, &Message{
		Type: MessageTypeAppSync,
		Data: sync,
	})
}

// EmitPipelineStatus emits pipeline status update
func (e *EventEmitter) EmitPipelineStatus(pipelineID string, status interface{}) {
	e.hub.BroadcastToChannel("pipeline:"+pipelineID, &Message{
		Type: MessageTypePipelineStatus,
		Data: status,
	})
}

// EmitPipelineLog emits pipeline log line
func (e *EventEmitter) EmitPipelineLog(pipelineID string, log interface{}) {
	e.hub.BroadcastToChannel("pipeline:"+pipelineID, &Message{
		Type: MessageTypePipelineLog,
		Data: log,
	})
}

// EmitPodStatus emits pod status update
func (e *EventEmitter) EmitPodStatus(namespace, podName string, status interface{}) {
	e.hub.BroadcastToChannel("pod:"+namespace+"/"+podName, &Message{
		Type: MessageTypePodStatus,
		Data: status,
	})
}

// EmitPodLogs emits pod log lines
func (e *EventEmitter) EmitPodLogs(namespace, podName string, logs interface{}) {
	e.hub.BroadcastToChannel("pod:"+namespace+"/"+podName, &Message{
		Type: MessageTypePodLogs,
		Data: logs,
	})
}

// EmitAlert emits an alert
func (e *EventEmitter) EmitAlert(alert interface{}) {
	e.hub.Broadcast(&Message{
		Type:    MessageTypeAlert,
		Channel: "alerts",
		Data:    alert,
	})
}

// EmitCostUpdate emits cost update
func (e *EventEmitter) EmitCostUpdate(data interface{}) {
	e.hub.BroadcastToChannel("cost", &Message{
		Type: MessageTypeCostUpdate,
		Data: data,
	})
}

// EmitAIResponse emits AI response
func (e *EventEmitter) EmitAIResponse(clientID string, response interface{}) {
	e.hub.SendToClient(clientID, &Message{
		Type: MessageTypeAIResponse,
		Data: response,
	})
}

// EmitAIStream emits AI streaming response chunk
func (e *EventEmitter) EmitAIStream(clientID string, chunk interface{}) {
	e.hub.SendToClient(clientID, &Message{
		Type: MessageTypeAIStream,
		Data: chunk,
	})
}
