// Package nats provides NATS messaging client for Krustron
// Author: Anubhav Gain <anubhavg@infopercept.com>
package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Config holds NATS client configuration
type Config struct {
	URL              string
	ClusterID        string
	ClientID         string
	Token            string
	Username         string
	Password         string
	TLSEnabled       bool
	TLSCertFile      string
	TLSKeyFile       string
	TLSCAFile        string
	ReconnectWait    time.Duration
	MaxReconnects    int
	PingInterval     time.Duration
	MaxPingsOut      int
	DrainTimeout     time.Duration
	JetStreamEnabled bool
}

// Client provides NATS messaging operations
type Client struct {
	conn         *nats.Conn
	js           nats.JetStreamContext
	logger       *zap.Logger
	config       *Config
	subscriptions map[string]*nats.Subscription
	subMu        sync.RWMutex
	handlers     map[string][]MessageHandler
	handlerMu    sync.RWMutex
}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, msg *Message) error

// Message represents a NATS message
type Message struct {
	Subject   string                 `json:"subject"`
	Data      []byte                 `json:"data"`
	Headers   map[string]string      `json:"headers"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
	ReplyTo   string                 `json:"reply_to,omitempty"`
	Sequence  uint64                 `json:"sequence,omitempty"`
}

// Event represents a Krustron event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Subject   string                 `json:"subject"`
	Data      interface{}            `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// Subjects for Krustron events
const (
	SubjectClusterEvents     = "krustron.cluster.>"
	SubjectApplicationEvents = "krustron.application.>"
	SubjectPipelineEvents    = "krustron.pipeline.>"
	SubjectDeploymentEvents  = "krustron.deployment.>"
	SubjectSecurityEvents    = "krustron.security.>"
	SubjectAlertEvents       = "krustron.alert.>"
	SubjectAuditEvents       = "krustron.audit.>"
)

// Stream names for JetStream
const (
	StreamCluster     = "KRUSTRON_CLUSTER"
	StreamApplication = "KRUSTRON_APPLICATION"
	StreamPipeline    = "KRUSTRON_PIPELINE"
	StreamDeployment  = "KRUSTRON_DEPLOYMENT"
	StreamSecurity    = "KRUSTRON_SECURITY"
	StreamAlert       = "KRUSTRON_ALERT"
	StreamAudit       = "KRUSTRON_AUDIT"
)

// NewClient creates a new NATS client
func NewClient(logger *zap.Logger, config *Config) (*Client, error) {
	// Set defaults
	if config.URL == "" {
		config.URL = nats.DefaultURL
	}
	if config.ReconnectWait == 0 {
		config.ReconnectWait = 2 * time.Second
	}
	if config.MaxReconnects == 0 {
		config.MaxReconnects = 60
	}
	if config.PingInterval == 0 {
		config.PingInterval = 20 * time.Second
	}
	if config.MaxPingsOut == 0 {
		config.MaxPingsOut = 3
	}
	if config.DrainTimeout == 0 {
		config.DrainTimeout = 30 * time.Second
	}

	// Build connection options
	opts := []nats.Option{
		nats.ReconnectWait(config.ReconnectWait),
		nats.MaxReconnects(config.MaxReconnects),
		nats.PingInterval(config.PingInterval),
		nats.MaxPingsOutstanding(config.MaxPingsOut),
		nats.DrainTimeout(config.DrainTimeout),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logger.Warn("NATS disconnected", zap.Error(err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("NATS reconnected", zap.String("url", nc.ConnectedUrl()))
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			logger.Info("NATS connection closed")
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			logger.Error("NATS error",
				zap.String("subject", sub.Subject),
				zap.Error(err),
			)
		}),
	}

	// Add authentication
	if config.Token != "" {
		opts = append(opts, nats.Token(config.Token))
	} else if config.Username != "" && config.Password != "" {
		opts = append(opts, nats.UserInfo(config.Username, config.Password))
	}

	// Add TLS if enabled
	if config.TLSEnabled {
		if config.TLSCertFile != "" && config.TLSKeyFile != "" {
			opts = append(opts, nats.ClientCert(config.TLSCertFile, config.TLSKeyFile))
		}
		if config.TLSCAFile != "" {
			opts = append(opts, nats.RootCAs(config.TLSCAFile))
		}
	}

	// Connect to NATS
	conn, err := nats.Connect(config.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	client := &Client{
		conn:          conn,
		logger:        logger,
		config:        config,
		subscriptions: make(map[string]*nats.Subscription),
		handlers:      make(map[string][]MessageHandler),
	}

	// Setup JetStream if enabled
	if config.JetStreamEnabled {
		js, err := conn.JetStream()
		if err != nil {
			logger.Warn("Failed to create JetStream context", zap.Error(err))
		} else {
			client.js = js
			if err := client.setupStreams(); err != nil {
				logger.Warn("Failed to setup streams", zap.Error(err))
			}
		}
	}

	logger.Info("Connected to NATS",
		zap.String("url", conn.ConnectedUrl()),
		zap.Bool("jetstream", config.JetStreamEnabled),
	)

	return client, nil
}

// setupStreams creates JetStream streams
func (c *Client) setupStreams() error {
	streams := []struct {
		Name     string
		Subjects []string
	}{
		{StreamCluster, []string{"krustron.cluster.>"}},
		{StreamApplication, []string{"krustron.application.>"}},
		{StreamPipeline, []string{"krustron.pipeline.>"}},
		{StreamDeployment, []string{"krustron.deployment.>"}},
		{StreamSecurity, []string{"krustron.security.>"}},
		{StreamAlert, []string{"krustron.alert.>"}},
		{StreamAudit, []string{"krustron.audit.>"}},
	}

	for _, s := range streams {
		_, err := c.js.StreamInfo(s.Name)
		if err == nats.ErrStreamNotFound {
			_, err = c.js.AddStream(&nats.StreamConfig{
				Name:       s.Name,
				Subjects:   s.Subjects,
				Retention:  nats.LimitsPolicy,
				MaxAge:     7 * 24 * time.Hour, // 7 days retention
				MaxMsgs:    -1,
				MaxBytes:   -1,
				Discard:    nats.DiscardOld,
				Storage:    nats.FileStorage,
				Replicas:   1,
				Duplicates: 2 * time.Minute,
			})
			if err != nil {
				c.logger.Warn("Failed to create stream",
					zap.String("stream", s.Name),
					zap.Error(err),
				)
			} else {
				c.logger.Info("Created stream", zap.String("stream", s.Name))
			}
		}
	}

	return nil
}

// Publish publishes a message to a subject
func (c *Client) Publish(ctx context.Context, subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if c.js != nil {
		_, err = c.js.Publish(subject, payload)
	} else {
		err = c.conn.Publish(subject, payload)
	}

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	c.logger.Debug("Published message",
		zap.String("subject", subject),
		zap.Int("size", len(payload)),
	)

	return nil
}

// PublishEvent publishes a Krustron event
func (c *Client) PublishEvent(ctx context.Context, event *Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	subject := fmt.Sprintf("krustron.%s.%s", event.Source, event.Type)
	return c.Publish(ctx, subject, event)
}

// Request sends a request and waits for a response
func (c *Client) Request(ctx context.Context, subject string, data interface{}, timeout time.Duration) (*Message, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	msg, err := c.conn.Request(subject, payload, timeout)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return &Message{
		Subject:   msg.Subject,
		Data:      msg.Data,
		ReplyTo:   msg.Reply,
		Timestamp: time.Now(),
	}, nil
}

// Subscribe subscribes to a subject with a handler
func (c *Client) Subscribe(subject string, handler MessageHandler) error {
	c.handlerMu.Lock()
	c.handlers[subject] = append(c.handlers[subject], handler)
	c.handlerMu.Unlock()

	c.subMu.Lock()
	defer c.subMu.Unlock()

	// Check if already subscribed
	if _, exists := c.subscriptions[subject]; exists {
		return nil
	}

	sub, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
		c.handleMessage(subject, msg)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	c.subscriptions[subject] = sub
	c.logger.Info("Subscribed to subject", zap.String("subject", subject))

	return nil
}

// SubscribeQueue subscribes to a subject with queue group
func (c *Client) SubscribeQueue(subject, queue string, handler MessageHandler) error {
	c.handlerMu.Lock()
	key := fmt.Sprintf("%s:%s", subject, queue)
	c.handlers[key] = append(c.handlers[key], handler)
	c.handlerMu.Unlock()

	c.subMu.Lock()
	defer c.subMu.Unlock()

	if _, exists := c.subscriptions[key]; exists {
		return nil
	}

	sub, err := c.conn.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		c.handleMessage(key, msg)
	})
	if err != nil {
		return fmt.Errorf("failed to queue subscribe: %w", err)
	}

	c.subscriptions[key] = sub
	c.logger.Info("Queue subscribed",
		zap.String("subject", subject),
		zap.String("queue", queue),
	)

	return nil
}

// SubscribeJetStream subscribes to a JetStream stream
func (c *Client) SubscribeJetStream(stream, consumer string, handler MessageHandler, opts ...nats.SubOpt) error {
	if c.js == nil {
		return fmt.Errorf("JetStream not enabled")
	}

	key := fmt.Sprintf("js:%s:%s", stream, consumer)

	c.handlerMu.Lock()
	c.handlers[key] = append(c.handlers[key], handler)
	c.handlerMu.Unlock()

	c.subMu.Lock()
	defer c.subMu.Unlock()

	if _, exists := c.subscriptions[key]; exists {
		return nil
	}

	// Default options
	defaultOpts := []nats.SubOpt{
		nats.Durable(consumer),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.DeliverNew(),
	}
	opts = append(defaultOpts, opts...)

	sub, err := c.js.Subscribe("", func(msg *nats.Msg) {
		c.handleJetStreamMessage(key, msg)
	}, opts...)
	if err != nil {
		return fmt.Errorf("failed to JetStream subscribe: %w", err)
	}

	c.subscriptions[key] = sub
	c.logger.Info("JetStream subscribed",
		zap.String("stream", stream),
		zap.String("consumer", consumer),
	)

	return nil
}

func (c *Client) handleMessage(key string, msg *nats.Msg) {
	c.handlerMu.RLock()
	handlers := c.handlers[key]
	c.handlerMu.RUnlock()

	m := &Message{
		Subject:   msg.Subject,
		Data:      msg.Data,
		ReplyTo:   msg.Reply,
		Headers:   make(map[string]string),
		Timestamp: time.Now(),
	}

	// Parse headers if present
	if msg.Header != nil {
		for k, v := range msg.Header {
			if len(v) > 0 {
				m.Headers[k] = v[0]
			}
		}
	}

	ctx := context.Background()
	for _, handler := range handlers {
		if err := handler(ctx, m); err != nil {
			c.logger.Error("Handler error",
				zap.String("subject", msg.Subject),
				zap.Error(err),
			)
		}
	}
}

func (c *Client) handleJetStreamMessage(key string, msg *nats.Msg) {
	c.handlerMu.RLock()
	handlers := c.handlers[key]
	c.handlerMu.RUnlock()

	meta, _ := msg.Metadata()
	m := &Message{
		Subject:   msg.Subject,
		Data:      msg.Data,
		ReplyTo:   msg.Reply,
		Headers:   make(map[string]string),
		Timestamp: time.Now(),
	}

	if meta != nil {
		m.Sequence = meta.Sequence.Stream
		m.Timestamp = meta.Timestamp
	}

	if msg.Header != nil {
		for k, v := range msg.Header {
			if len(v) > 0 {
				m.Headers[k] = v[0]
			}
		}
	}

	ctx := context.Background()
	for _, handler := range handlers {
		if err := handler(ctx, m); err != nil {
			c.logger.Error("JetStream handler error",
				zap.String("subject", msg.Subject),
				zap.Error(err),
			)
			// NAK the message on error
			msg.Nak()
			return
		}
	}

	// ACK successful processing
	msg.Ack()
}

// Unsubscribe removes a subscription
func (c *Client) Unsubscribe(subject string) error {
	c.subMu.Lock()
	defer c.subMu.Unlock()

	if sub, exists := c.subscriptions[subject]; exists {
		if err := sub.Unsubscribe(); err != nil {
			return fmt.Errorf("failed to unsubscribe: %w", err)
		}
		delete(c.subscriptions, subject)
	}

	c.handlerMu.Lock()
	delete(c.handlers, subject)
	c.handlerMu.Unlock()

	return nil
}

// CreateConsumer creates a JetStream consumer
func (c *Client) CreateConsumer(stream string, config *nats.ConsumerConfig) (*nats.ConsumerInfo, error) {
	if c.js == nil {
		return nil, fmt.Errorf("JetStream not enabled")
	}

	return c.js.AddConsumer(stream, config)
}

// DeleteConsumer deletes a JetStream consumer
func (c *Client) DeleteConsumer(stream, consumer string) error {
	if c.js == nil {
		return fmt.Errorf("JetStream not enabled")
	}

	return c.js.DeleteConsumer(stream, consumer)
}

// GetStreamInfo returns stream information
func (c *Client) GetStreamInfo(stream string) (*nats.StreamInfo, error) {
	if c.js == nil {
		return nil, fmt.Errorf("JetStream not enabled")
	}

	return c.js.StreamInfo(stream)
}

// Flush flushes the connection buffer
func (c *Client) Flush() error {
	return c.conn.Flush()
}

// Drain gracefully drains the connection
func (c *Client) Drain() error {
	return c.conn.Drain()
}

// Close closes the connection
func (c *Client) Close() {
	c.subMu.Lock()
	for _, sub := range c.subscriptions {
		sub.Unsubscribe()
	}
	c.subscriptions = make(map[string]*nats.Subscription)
	c.subMu.Unlock()

	c.conn.Close()
	c.logger.Info("NATS connection closed")
}

// IsConnected returns true if connected
func (c *Client) IsConnected() bool {
	return c.conn.IsConnected()
}

// Stats returns connection statistics
func (c *Client) Stats() nats.Statistics {
	return c.conn.Stats()
}

// EventBus provides a high-level event bus abstraction
type EventBus struct {
	client *Client
	logger *zap.Logger
}

// NewEventBus creates a new event bus
func NewEventBus(client *Client, logger *zap.Logger) *EventBus {
	return &EventBus{
		client: client,
		logger: logger,
	}
}

// EmitClusterEvent emits a cluster-related event
func (eb *EventBus) EmitClusterEvent(ctx context.Context, eventType string, clusterID string, data interface{}) error {
	event := &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Source:    "cluster",
		Subject:   fmt.Sprintf("krustron.cluster.%s.%s", clusterID, eventType),
		Data:      data,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"cluster_id": clusterID,
		},
	}
	return eb.client.Publish(ctx, event.Subject, event)
}

// EmitApplicationEvent emits an application-related event
func (eb *EventBus) EmitApplicationEvent(ctx context.Context, eventType, appID, namespace string, data interface{}) error {
	event := &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Source:    "application",
		Subject:   fmt.Sprintf("krustron.application.%s.%s", appID, eventType),
		Data:      data,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"app_id":    appID,
			"namespace": namespace,
		},
	}
	return eb.client.Publish(ctx, event.Subject, event)
}

// EmitPipelineEvent emits a pipeline-related event
func (eb *EventBus) EmitPipelineEvent(ctx context.Context, eventType, pipelineID string, data interface{}) error {
	event := &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Source:    "pipeline",
		Subject:   fmt.Sprintf("krustron.pipeline.%s.%s", pipelineID, eventType),
		Data:      data,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"pipeline_id": pipelineID,
		},
	}
	return eb.client.Publish(ctx, event.Subject, event)
}

// EmitDeploymentEvent emits a deployment-related event
func (eb *EventBus) EmitDeploymentEvent(ctx context.Context, eventType, deploymentID, environment string, data interface{}) error {
	event := &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Source:    "deployment",
		Subject:   fmt.Sprintf("krustron.deployment.%s.%s", deploymentID, eventType),
		Data:      data,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"deployment_id": deploymentID,
			"environment":   environment,
		},
	}
	return eb.client.Publish(ctx, event.Subject, event)
}

// EmitSecurityEvent emits a security-related event
func (eb *EventBus) EmitSecurityEvent(ctx context.Context, eventType, severity string, data interface{}) error {
	event := &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Source:    "security",
		Subject:   fmt.Sprintf("krustron.security.%s.%s", severity, eventType),
		Data:      data,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"severity": severity,
		},
	}
	return eb.client.Publish(ctx, event.Subject, event)
}

// EmitAlertEvent emits an alert event
func (eb *EventBus) EmitAlertEvent(ctx context.Context, eventType, severity, source string, data interface{}) error {
	event := &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Source:    "alert",
		Subject:   fmt.Sprintf("krustron.alert.%s.%s", severity, eventType),
		Data:      data,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"severity":     severity,
			"alert_source": source,
		},
	}
	return eb.client.Publish(ctx, event.Subject, event)
}

// EmitAuditEvent emits an audit event
func (eb *EventBus) EmitAuditEvent(ctx context.Context, action, userID, resource string, data interface{}) error {
	event := &Event{
		ID:        generateEventID(),
		Type:      action,
		Source:    "audit",
		Subject:   fmt.Sprintf("krustron.audit.%s.%s", userID, action),
		Data:      data,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"user_id":  userID,
			"resource": resource,
		},
	}
	return eb.client.Publish(ctx, event.Subject, event)
}

// OnClusterEvent registers a handler for cluster events
func (eb *EventBus) OnClusterEvent(handler func(ctx context.Context, event *Event) error) error {
	return eb.client.Subscribe(SubjectClusterEvents, func(ctx context.Context, msg *Message) error {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return err
		}
		return handler(ctx, &event)
	})
}

// OnApplicationEvent registers a handler for application events
func (eb *EventBus) OnApplicationEvent(handler func(ctx context.Context, event *Event) error) error {
	return eb.client.Subscribe(SubjectApplicationEvents, func(ctx context.Context, msg *Message) error {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return err
		}
		return handler(ctx, &event)
	})
}

// OnPipelineEvent registers a handler for pipeline events
func (eb *EventBus) OnPipelineEvent(handler func(ctx context.Context, event *Event) error) error {
	return eb.client.Subscribe(SubjectPipelineEvents, func(ctx context.Context, msg *Message) error {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return err
		}
		return handler(ctx, &event)
	})
}

// OnDeploymentEvent registers a handler for deployment events
func (eb *EventBus) OnDeploymentEvent(handler func(ctx context.Context, event *Event) error) error {
	return eb.client.Subscribe(SubjectDeploymentEvents, func(ctx context.Context, msg *Message) error {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return err
		}
		return handler(ctx, &event)
	})
}

// OnSecurityEvent registers a handler for security events
func (eb *EventBus) OnSecurityEvent(handler func(ctx context.Context, event *Event) error) error {
	return eb.client.Subscribe(SubjectSecurityEvents, func(ctx context.Context, msg *Message) error {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return err
		}
		return handler(ctx, &event)
	})
}

// OnAlertEvent registers a handler for alert events
func (eb *EventBus) OnAlertEvent(handler func(ctx context.Context, event *Event) error) error {
	return eb.client.Subscribe(SubjectAlertEvents, func(ctx context.Context, msg *Message) error {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return err
		}
		return handler(ctx, &event)
	})
}

// OnAuditEvent registers a handler for audit events
func (eb *EventBus) OnAuditEvent(handler func(ctx context.Context, event *Event) error) error {
	return eb.client.Subscribe(SubjectAuditEvents, func(ctx context.Context, msg *Message) error {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return err
		}
		return handler(ctx, &event)
	})
}

func generateEventID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
}
