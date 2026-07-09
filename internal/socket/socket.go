// Package socket manages an addr's live attachment to the host-owned fabric
// for the addr's whole lifetime: warm while attached, parked to a resumable
// ticket when idle. The addr is the durable identity; the socket is not. The
// fabric is the network — the host stands it up and owns its lifecycle; the
// socket only rides its subtree and never creates or closes it.
package socket

import (
	"context"
	"fmt"
	"sync"

	"github.com/joaomdsg/packets/internal/fabric"
)

// Socket is one addr's attachment onto the shared, host-owned fabric.
type Socket struct {
	mu     sync.Mutex
	fab    *fabric.Fabric
	addr   string
	ticket *Ticket // non-nil while parked
	closed bool
}

// Ticket is a single-use, in-memory handle on a parked addr, carrying the
// shared fabric and the addr so Resume can re-attach without rebuilding the
// network. It is redeemable exactly once.
type Ticket struct {
	mu   sync.Mutex
	fab  *fabric.Fabric
	addr string
	used bool
}

// Open attaches addr's endpoint onto the host-owned fabric fab. addr is the
// durable owner/repo subject token (never a host:port); fab is the shared
// network the host already stood up. The socket rides fab; it does not own it.
func Open(ctx context.Context, fab *fabric.Fabric, addr string) (*Socket, error) {
	if fab == nil {
		return nil, fmt.Errorf("socket: open %q: nil fabric", addr)
	}
	if addr == "" {
		return nil, fmt.Errorf("socket: open: empty addr")
	}
	return &Socket{fab: fab, addr: addr}, nil
}

// Addr is the addr claimed by Open, stable across park/resume.
func (s *Socket) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addr
}

// Fabric is the shared, host-owned fabric the addr rides. It stays valid
// across Park and after Close: the socket never owns the fabric's lifecycle.
func (s *Socket) Fabric() *fabric.Fabric {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.fab
}

// Park releases the addr's warm attachment and returns a single-use Ticket
// that Resume redeems to re-attach. The shared fabric is left untouched — park
// is a socket-only state change, not a teardown of the network.
func (s *Socket) Park() (*Ticket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, fmt.Errorf("socket: park %q: already closed", s.addr)
	}
	if s.ticket != nil {
		return nil, fmt.Errorf("socket: park %q: already parked", s.addr)
	}
	t := &Ticket{fab: s.fab, addr: s.addr}
	s.ticket = t
	return t, nil
}

// Resume re-attaches a parked ticket's addr as a warm Socket on the same
// shared fabric. It never creates a fabric. The ticket is consumed; redeeming
// it a second time, or after the parking socket was closed, errors.
func Resume(ctx context.Context, t *Ticket) (*Socket, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.used {
		return nil, fmt.Errorf("socket: resume: ticket already used")
	}
	t.used = true
	return &Socket{fab: t.fab, addr: t.addr}, nil
}

// Close releases the socket. The host-owned fabric is never closed. While
// parked, Close invalidates the outstanding ticket so a stale ticket can never
// resurrect an endpoint the owner already tore down. Close is idempotent.
func (s *Socket) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if s.ticket != nil {
		t := s.ticket
		s.ticket = nil
		t.mu.Lock()
		t.used = true
		t.mu.Unlock()
	}
	return nil
}
