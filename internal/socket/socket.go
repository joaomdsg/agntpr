// Package socket manages an addr's live attachment to the fabric for the
// addr's whole lifetime: warm while listening, parked to a resumable ticket
// when idle. The addr is the durable identity; the socket is not.
package socket

import (
	"context"
	"fmt"
	"sync"

	"github.com/joaomdsg/packets/internal/fabric"
)

// Socket is one addr's live or parked attachment to the fabric.
type Socket struct {
	mu     sync.Mutex
	addr   string
	dir    string
	grants []fabric.Grant
	fab    *fabric.Fabric // non-nil while listening
	ticket *Ticket        // non-nil while parked
	closed bool
}

// Ticket is a single-use, in-memory handle on a parked addr's placeholder
// binding, redeemable exactly once via Resume.
type Ticket struct {
	mu      sync.Mutex
	binding *fabric.Binding
	dir     string
	grants  []fabric.Grant
	used    bool
}

// Open binds addr and starts it listening, backed by a fabric storing under
// dir with the given peer grants.
func Open(ctx context.Context, addr, dir string, grants ...fabric.Grant) (*Socket, error) {
	b, err := fabric.Bind(addr)
	if err != nil {
		return nil, err
	}
	f, err := b.Listen(ctx, dir, grants...)
	if err != nil {
		return nil, err
	}
	return &Socket{addr: b.Addr(), dir: dir, grants: grants, fab: f}, nil
}

// Addr is the addr claimed by Open, stable across park/resume.
func (s *Socket) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addr
}

// Fabric is the live fabric while listening, or nil while parked or closed.
func (s *Socket) Fabric() *fabric.Fabric {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.fab
}

// Park closes the live fabric and immediately re-binds the same addr via a
// placeholder, so the addr stays claimed while idle. It returns a single-use
// Ticket that Resume redeems to warm the addr back up.
func (s *Socket) Park() (*Ticket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, fmt.Errorf("socket: park %q: already closed", s.addr)
	}
	if s.fab == nil {
		return nil, fmt.Errorf("socket: park %q: already parked", s.addr)
	}
	if err := s.fab.Close(); err != nil {
		return nil, err
	}
	s.fab = nil

	b, err := fabric.Bind(s.addr)
	if err != nil {
		return nil, err
	}
	t := &Ticket{binding: b, dir: s.dir, grants: s.grants}
	s.ticket = t
	return t, nil
}

// Resume warms a parked ticket's held binding back into a listening Socket,
// on the same addr, dir, and grants it was parked with. The ticket is
// consumed; redeeming it a second time errors.
func Resume(ctx context.Context, t *Ticket) (*Socket, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.used {
		return nil, fmt.Errorf("socket: resume: ticket already used")
	}
	t.used = true

	f, err := t.binding.Listen(ctx, t.dir, t.grants...)
	if err != nil {
		return nil, err
	}
	return &Socket{addr: f.Addr(), dir: t.dir, grants: t.grants, fab: f}, nil
}

// Close releases the socket: while listening, it closes the live fabric;
// while parked, it releases the placeholder and invalidates the outstanding
// ticket. Close is idempotent.
func (s *Socket) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true

	if s.fab != nil {
		f := s.fab
		s.fab = nil
		return f.Close()
	}
	if s.ticket != nil {
		t := s.ticket
		s.ticket = nil
		t.mu.Lock()
		t.used = true
		b := t.binding
		t.mu.Unlock()
		return b.Close()
	}
	return nil
}
