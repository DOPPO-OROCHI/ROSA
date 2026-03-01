package transport

import "sync"

type StreamKey struct {
	MatchID      uint
	ViewerUserID uint
}

type Hub struct {
	mu   sync.RWMutex
	subs map[StreamKey]map[chan []byte]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: make(map[StreamKey]map[chan []byte]struct{})}
}

func (h *Hub) Subscribe(key StreamKey) chan []byte {
	ch := make(chan []byte, 16)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subs[key] == nil {
		h.subs[key] = make(map[chan []byte]struct{})
	}
	h.subs[key][ch] = struct{}{}
	return ch
}

func (h *Hub) Unsubscribe(key StreamKey, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.subs[key]; ok {
		delete(set, ch)
		if len(set) == 0 {
			delete(h.subs, key)
		}
	}
	close(ch)
}

func (h *Hub) Publish(key StreamKey, msg []byte) {
	h.mu.RLock()
	set := h.subs[key]
	h.mu.RUnlock()
	for ch := range set {
		select {
		case ch <- msg:
		default:
		}
	}
}
