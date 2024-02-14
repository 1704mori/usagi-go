package usagi

type Usagi struct {
	ConnectionManager *Connection
}

func NewUsagi(config ConnectionConfig) *Usagi {
	return &Usagi{
		ConnectionManager: NewConnection(config),
	}
}

func (u *Usagi) Initialize(name string) error {
	return u.ConnectionManager.Initialize(name)
}

func (u *Usagi) Publish(queue string, payload interface{}, options PublishOptions) (sent bool, err error) {
	publisher := NewPublisher(u.ConnectionManager, queue, options)
	return publisher.Send(payload)
}

func (u *Usagi) Queue(queueName string) *Listener {
	return NewListener(u.ConnectionManager, queueName)
}

func (u *Usagi) Close() error {
	return u.ConnectionManager.Close()
}
