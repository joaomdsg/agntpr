package fabric

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/nats-io/nats-server/v2/server"
)

// Binding is a TCP addr claimed ahead of the fabric being live on it: bound =
// addr reserved + announceable, independent of any running server. Listen
// warms it up into a live fabric; Close releases the addr unused.
type Binding struct {
	mu          sync.Mutex
	placeholder net.Listener
	addr        string
	consumed    bool
}

// Bind claims addr (host:port; port 0 resolves to a random free port) now, via
// a placeholder TCP listener, without booting a fabric on it yet.
func Bind(addr string) (*Binding, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("fabric: bind %q: %v", addr, err)
	}
	return &Binding{placeholder: ln, addr: ln.Addr().String()}, nil
}

// Addr is the concrete host:port claimed by Bind (port 0 resolved).
func (b *Binding) Addr() string {
	return b.addr
}

// Listen warms the binding up into a live fabric: it releases the placeholder
// listener and boots the embedded server on the exact same host:port, with the
// same host/peer auth model as StartListening. A Binding is consumed by its
// first successful Listen; a second Listen, or a Listen after Close, errors.
func (b *Binding) Listen(ctx context.Context, dir string, grants ...Grant) (*Fabric, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.consumed {
		return nil, fmt.Errorf("fabric: binding %q already consumed", b.addr)
	}
	host, port, err := splitAddr(b.addr)
	if err != nil {
		return nil, err
	}
	b.placeholder.Close()
	b.consumed = true

	f, err := boot(&server.Options{
		StoreDir:   dir,
		Host:       host,
		Port:       port,
		Users:      authUsers(grants),
		NoAuthUser: hostUser,
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Close releases the bound addr without ever listening. It is safe to call
// after a successful Listen (no-op), since Listen already released it.
func (b *Binding) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.consumed {
		return nil
	}
	b.consumed = true
	return b.placeholder.Close()
}
