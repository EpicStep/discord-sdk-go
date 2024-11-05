package proto

// EventHandler ...
type EventHandler interface {
	OnEvent(eventType string, data []byte)
}

type noopHandler struct{}

func newNoopHandler() *noopHandler {
	return &noopHandler{}
}

func (noopHandler) OnEvent(_ string, _ []byte) {}
