package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/voice-diary/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

// Client represents a RabbitMQ client
type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewClient creates a new RabbitMQ client
func NewClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	logger.Info("Successfully connected to RabbitMQ")

	return &Client{
		conn:    conn,
		channel: channel,
	}, nil
}

// Close closes the RabbitMQ connection
func (c *Client) Close() error {
	logger.Info("Closing RabbitMQ connection")

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			return err
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}

	return nil
}

// DeclareQueue declares a queue
func (c *Client) DeclareQueue(name string) error {
	_, err := c.channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return err
}

// Publish publishes a message to a queue
func (c *Client) Publish(ctx context.Context, queueName string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = c.channel.PublishWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	logger.Debug("Published message to queue",
		zap.String("queue", queueName),
		zap.Int("size", len(body)),
	)

	return nil
}

// Consume consumes messages from a queue
func (c *Client) Consume(queueName string, handler func([]byte) error) error {
	msgs, err := c.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	logger.Info("Started consuming messages", zap.String("queue", queueName))

	go func() {
		for msg := range msgs {
			logger.Debug("Received message", zap.String("queue", queueName))

			if err := handler(msg.Body); err != nil {
				logger.Error("Failed to handle message",
					zap.String("queue", queueName),
					zap.Error(err),
				)
				msg.Nack(false, true) // Requeue the message
			} else {
				msg.Ack(false)
			}
		}
	}()

	return nil
}

// HealthCheck checks if RabbitMQ is healthy
func (c *Client) HealthCheck() error {
	if c.conn == nil || c.conn.IsClosed() {
		return fmt.Errorf("RabbitMQ connection is closed")
	}
	return nil
}
