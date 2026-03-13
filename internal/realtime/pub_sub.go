package realtime

type IPubSub interface {
	Initialize() error
	Publish(channel string, message *Envelope) error
	Subscribe(channel string)
	ListenToSubscriptions() <-chan any
}
