package main

import (
	"bytes"
	"strconv"
	"sync"
)

func itoa(n int) string { return strconv.Itoa(n) }

// syncBuf is a goroutine-safe writer for the server's log output.
type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}
