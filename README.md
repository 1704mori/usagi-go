# Usagi Go

!This project is still in development and is not ready for production use (kinda)!
<hr>
Usagi provides a simple and flexible interface for interacting with RabbitMQ, including message publishing, listening, and retry mechanisms.

## Install

Via go get tool

``` bash
go get github.com/1704mori/usagi-go
```

## Usage

Create and connect instance

``` go
client := usagi.NewUsagi(amqp.ConnectionConfig{
    URI:          "amqp://user:passwd@localhost:5672/vhost,
    Exchange:     "my_exchange",
    ExchangeType: "topic",
})

err := client.Initialize("connection_name")
if err != nil {
    defer client.Close()
    log.Fatalf("Failed to initialize: %v", err)
}
```

Publish a message

``` go
publisher := usagi.NewPublisher(client.ConnectionManager, "my_queue", usagi.PublishOptions{})
sent, err := publisher.Send(payload interface{})

// or
sent, err := client.Publish("my_queue", payload interface{}, usagi.PublishOptions{})
```

Listen to a queue

``` go
listener := usagi.NewListener(client.ConnectionManager, "my_queue")

err := listener.Listen(func(message interface{}) bool {
    value, err := someFunctionCall()
    if err != nil {
        return false // returning false will nack the message, triggering the retry mechanism, which tries 3 times before removing it from the queue
    }

    return true // ack the message
})

if err != nil {
  log.Fatalf("Failed to start listening: %v", err)
}

// or
client.Queue("my_queue").Listen(func(message interface{}) bool {
  // return true or false
})

```

## License

This project is licensed under the [MIT License](LICENSE).
