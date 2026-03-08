package realtime

type IPubSub interface {
	Initialize() error
	Publish(channel string, message *ClientMessage) error
	Subscribe(channel string)
	ListenToSubscriptions() <-chan any
}
