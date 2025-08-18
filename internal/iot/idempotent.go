package iot

import "sync"

// IdempotentSender wraps a CommandSender and suppresses duplicate consecutive
// commands (e.g. sending OFF twice). Useful to avoid double publish when
// auto-expire and manual stop happen nearly simultaneously.
type IdempotentSender struct {
	inner CommandSender
	mu    sync.Mutex
	last  map[int64]string
}

func NewIdempotentSender(inner CommandSender) *IdempotentSender {
	return &IdempotentSender{inner: inner, last: make(map[int64]string)}
}

func (s *IdempotentSender) Send(consoleID int64, cmd string) error {
	s.mu.Lock()
	prev := s.last[consoleID]
	if prev == cmd { // suppress duplicate
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()
	if err := s.inner.Send(consoleID, cmd); err != nil {
		return err
	}
	s.mu.Lock()
	s.last[consoleID] = cmd
	s.mu.Unlock()
	return nil
}
