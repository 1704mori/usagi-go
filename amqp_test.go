package usagi_test

import (
	"testing"

	"github.com/1704mori/usagi-go"
	"github.com/stretchr/testify/assert"
)

func TestNewUsagi(t *testing.T) {
	config := usagi.ConnectionConfig{
		URI:          "amqp://guest:guest@localhost:5672",
		Exchange:     "test-exchange",
		ExchangeType: "topic",
	}
	usagi := usagi.NewUsagi(config)

	assert.NotNil(t, usagi)
}

// todo: Add tests for Initialize, Publish, Queue, and Close methods
