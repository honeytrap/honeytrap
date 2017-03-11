package message

// Message defines a struct which contains specific data relating to
// different messages to provide Push notifications for the pusher api.
type PushMessage struct {
	Sensor      string
	Category    string
	SessionID   string
	ContainerID string
	Data        interface{}
}
