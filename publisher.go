package usagi

import (
	"context"
	"encoding/json"
	"log"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rabbitmq/amqp091-go"
)

type PublishOptions struct {
	Persistent bool // Indicates if messages should be marked as persistent.
}

type Publisher struct {
	connectionManager *Connection    // Manages connection to the AMQP server.
	queue             string         // The name of the queue to publish messages to.
	options           PublishOptions // Configuration options for publishing messages.
}

// NewPublisher creates a new instance of a Publisher.
// This function takes a connection manager, the name of the queue to publish messages to,
// and publishing options. It returns a pointer to a new Publisher instance.
func NewPublisher(conn *Connection, queue string, options PublishOptions) *Publisher {
	return &Publisher{
		connectionManager: conn,
		queue:             queue,
		options:           options,
	}
}

// Send marshals the payload into JSON and publishes it to the configured queue.
// It returns a boolean indicating whether the message was sent successfully and an error if something went wrong.
func (p *Publisher) Send(payload interface{}) (sent bool, err error) {
	channel := p.connectionManager.GetChannel() // Obtain a channel from the connection manager.

	body, err := json.Marshal(payload) // Marshal the payload into JSON format.
	if err != nil {
		return false, err
	}

	id, _ := gonanoid.New() // Generate a unique ID for the message.

	err = channel.PublishWithContext(
		context.Background(),              // Use a background context.
		p.connectionManager.GetExchange(), // Get the exchange from the connection manager.
		p.queue,                           // The queue to publish to.
		false,                             // mandatory flag.
		false,                             // immediate flag.
		amqp091.Publishing{ // Message publishing options.
			DeliveryMode: amqp091.Persistent, // Ensure message is persisted.
			MessageId:    id,                 // Unique message ID.
			ContentType:  "application/json", // Content type of the message.
			Body:         body,               // The actual message body.
		},
	)
	if err != nil {
		log.Printf("[queue] couldn't send message to %s, buffer size: %d", p.queue, len(body))
		return false, err
	}

	log.Printf("[queue] message sent to %s, buffer size: %d", p.queue, len(body))
	return true, nil
}
