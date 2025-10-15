package api

type StreamEvent string

const (
	EVENT_INITIAL StreamEvent = "event_initial"
	EVENT_APPEND  StreamEvent = "event_append"
	EVENT_ERROR   StreamEvent = "event_error"
)

type StreamMessage struct {
	Type StreamEvent `json:"type"`
	Data any         `json:"data"`
}
