//go:build js && wasm

package tgwasm

import "sync"

// serial runs funcs one at a time, in order, on a goroutine of its own.
// do never blocks, so it is safe on a JavaScript call.
type serial struct {
	mu    sync.Mutex
	queue []func()
	wake  chan struct{}
}

func newSerial() *serial {
	s := &serial{wake: make(chan struct{}, 1)}
	go s.loop()
	return s
}

func (s *serial) do(f func()) {
	s.mu.Lock()
	s.queue = append(s.queue, f)
	s.mu.Unlock()

	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *serial) loop() {
	for range s.wake {
		for {
			s.mu.Lock()
			if len(s.queue) == 0 {
				s.mu.Unlock()
				break
			}
			f := s.queue[0]
			// Cleared so a done func, and the session it holds, can be freed.
			s.queue[0] = nil
			s.queue = s.queue[1:]
			s.mu.Unlock()

			f()
		}
	}
}
