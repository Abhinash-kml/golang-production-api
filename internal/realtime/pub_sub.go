package realtime

type IPubSub interface {
	Publish(channel string, message *ClientMessage) error
	Subscribe(channel string) any
}
