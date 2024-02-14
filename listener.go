package usagi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Listener struct {
	connectionManager *Connection
	queueName         string
	RetryCount        int
	RetryTimeout      time.Duration
	DontSetupNack     bool
}

func NewListener(conn *Connection, queueName string) *Listener {
	return &Listener{
		connectionManager: conn,
		queueName:         queueName,
		RetryCount:        5,                // default retry count
		RetryTimeout:      60 * time.Second, // default retry timeout
	}
}

// Listen starts listening to the queue and processes messages using the provided callback.
func (l *Listener) Listen(callback func(message interface{}) bool) error {
	channel := l.connectionManager.GetChannel()

	// Declare the main queue
	_, err := channel.QueueDeclare(
		l.queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp091.Table{
			"x-dead-letter-exchange":    l.connectionManager.GetExchange(),
			"x-dead-letter-routing-key": fmt.Sprintf("%s.nack", l.queueName),
		},
	)
	if err != nil {
		return err
	}

	err = channel.QueueBind(l.queueName,
		l.queueName,
		l.connectionManager.GetExchange(),
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Declare a dead-letter queue for nacked messages
	nackQueueName := l.queueName + ".nack"
	if !l.DontSetupNack {
		err := l.setupNackQueue(channel, nackQueueName)
		if err != nil {
			return err
		}
	}

	msgs, err := channel.Consume(
		l.queueName,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			message := make(map[string]interface{})
			if err := json.Unmarshal(d.Body, &message); err != nil {
				log.Printf("Error decoding JSON: %s", err)
				d.Nack(false, false) // nack without requeue
				continue
			}

			if callback(message) {
				d.Ack(false)
			} else {
				l.handleNack(channel, &d, nackQueueName)
			}
		}
	}()

	return nil
}

// setupNackQueue sets up a dead-letter queue for nacked messages.
func (l *Listener) setupNackQueue(channel *amqp091.Channel, nackQueueName string) error {
	_, err := channel.QueueDeclare(
		nackQueueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp091.Table{
			"x-dead-letter-exchange":    l.connectionManager.GetExchange(),
			"x-dead-letter-routing-key": l.queueName,
			"x-message-ttl":             int64(l.RetryTimeout / time.Millisecond),
		},
	)
	if err != nil {
		return err
	}
	err = channel.QueueBind(l.queueName,
		l.queueName,
		l.connectionManager.GetExchange(),
		false,
		nil,
	)
	if err != nil {
		return err
	}
	return nil
}

// handleNack processes message nacks, including requeuing or sending to a dead-letter queue after too many retries.
func (l *Listener) handleNack(channel *amqp091.Channel, msg *amqp091.Delivery, nackQueueName string) {
	retryCount, ok := msg.Headers["x-retry-count"].(int)
	if !ok {
		retryCount = 0
	}

	if retryCount < l.RetryCount {
		headers := amqp091.Table{"x-retry-count": retryCount + 1}
		err := channel.PublishWithContext(
			context.Background(),
			l.connectionManager.GetExchange(),
			l.queueName,
			false,
			false,
			amqp091.Publishing{
				Headers:      headers,
				ContentType:  "application/json",
				Body:         msg.Body,
				DeliveryMode: amqp091.Persistent,
			},
		)
		if err != nil {
			log.Printf("[queue] failed to requeue message: %s", err)
		} else {
			log.Printf("[queue] message requeued (attempt %d)", retryCount+1)
		}
	} else {
		// Send to dead-letter queue after max retries
		err := channel.PublishWithContext(
			context.Background(),
			"", // default exchange
			nackQueueName,
			false,
			false,
			amqp091.Publishing{
				ContentType:  "application/json",
				Body:         msg.Body,
				DeliveryMode: amqp091.Persistent,
			},
		)
		if err != nil {
			log.Printf("[queue] failed to send message to nack queue: %s", err)
		} else {
			log.Printf("[queue] message sent to nack queue %s", nackQueueName)
		}
	}

	msg.Ack(false) // acknowledge the original message
}
