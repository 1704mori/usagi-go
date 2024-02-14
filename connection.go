package usagi

import (
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type ConnectionConfig struct {
	URI          string
	Exchange     string
	ExchangeType string
}

type Connection struct {
	conn     *amqp091.Connection
	channel  *amqp091.Channel
	exchange string
	config   ConnectionConfig
}

func NewConnection(config ConnectionConfig) *Connection {
	return &Connection{
		config: config,
	}
}

func (c *Connection) Initialize(name string) error {
	var err error
	log.Printf("Attempting to connect to RabbitMQ with URI: %s", c.config.URI)
	config := amqp091.Config{
		Properties: amqp091.NewConnectionProperties(),
	}
	config.Properties.SetClientConnectionName(name)
	c.conn, err = amqp091.DialConfig(c.config.URI, config)
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ: %v", err)
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	log.Printf("Successfully connected to RabbitMQ")

	c.channel, err = c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %v", err)
	}

	err = c.channel.ExchangeDeclare(
		c.config.Exchange,
		c.config.ExchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare an exchange: %v", err)
	}

	c.exchange = c.config.Exchange
	log.Printf("[queue] new connection %s created", name)
	return nil
}

func (c *Connection) Close() error {
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
	log.Println("[queue] connection closed")
	return nil
}

func (c *Connection) GetChannel() *amqp091.Channel {
	return c.channel
}

func (c *Connection) GetExchange() string {
	return c.exchange
}
