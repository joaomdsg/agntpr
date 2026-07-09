package socket_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joaomdsg/packets/internal/fabric"
	"github.com/joaomdsg/packets/internal/socket"
)

// newFabric stands up a host-owned in-process fabric for a test. The socket
// rides it; the host (here, the test) owns its lifecycle.
func newFabric(t *testing.T) *fabric.Fabric {
	t.Helper()
	f, err := fabric.Start(context.Background(), t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	return f
}

// originates publishes on the addr's subtree to prove the fabric is live and
// writable — the socket "originating" a send.
func originates(t *testing.T, f *fabric.Fabric) error {
	t.Helper()
	_, err := f.Publish(context.Background(), fabric.EventSubject("A", "i", fabric.StatusMinted, "catch"), []byte("m"))
	return err
}

func TestOpenAttachesAddrToTheSharedFabric(t *testing.T) {
	t.Parallel()
	fab := newFabric(t)

	s, err := socket.Open(context.Background(), fab, "octocat/hello")
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })

	f := s.Fabric()
	require.NotNil(t, f)
	assert.Same(t, fab, f, "the socket must expose the host-owned fabric it was opened onto, not a new one")
	assert.Equal(t, "octocat/hello", s.Addr())
	assert.NoError(t, originates(t, f))
}

func TestOpenRejectsANilFabric(t *testing.T) {
	t.Parallel()
	_, err := socket.Open(context.Background(), nil, "octocat/hello")
	assert.Error(t, err, "a socket cannot attach onto a nil fabric")
}

func TestOpenRejectsAnEmptyAddr(t *testing.T) {
	t.Parallel()
	fab := newFabric(t)
	_, err := socket.Open(context.Background(), fab, "")
	assert.Error(t, err, "a socket needs a durable addr identity")
}

func TestAddrIsStableAcrossParkAndResume(t *testing.T) {
	t.Parallel()
	fab := newFabric(t)
	s, err := socket.Open(context.Background(), fab, "octocat/hello")
	require.NoError(t, err)
	addr := s.Addr()

	ticket, err := s.Park()
	require.NoError(t, err)
	assert.Equal(t, addr, s.Addr())

	s2, err := socket.Resume(context.Background(), ticket)
	require.NoError(t, err)
	t.Cleanup(func() { _ = s2.Close() })

	assert.Equal(t, addr, s2.Addr())
}

// Parking used to tear the fabric down; now the fabric is host-owned, so park
// must leave it live — the addr's warm attachment is released, the network is
// not.
func TestParkLeavesTheSharedFabricLive(t *testing.T) {
	t.Parallel()
	fab := newFabric(t)
	s, err := socket.Open(context.Background(), fab, "octocat/hello")
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })

	_, err = s.Park()
	require.NoError(t, err)

	f := s.Fabric()
	require.NotNil(t, f, "Fabric() must stay valid across park")
	assert.Same(t, fab, f, "Fabric() must stay valid across park")
	assert.NoError(t, originates(t, f), "the shared fabric must still be writable after park")
}

func TestResumeRidesTheSameSharedFabric(t *testing.T) {
	t.Parallel()
	fab := newFabric(t)
	s, err := socket.Open(context.Background(), fab, "octocat/hello")
	require.NoError(t, err)
	addr := s.Addr()

	ticket, err := s.Park()
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })

	s2, err := socket.Resume(context.Background(), ticket)
	require.NoError(t, err)
	t.Cleanup(func() { _ = s2.Close() })

	assert.Equal(t, addr, s2.Addr())
	f := s2.Fabric()
	require.NotNil(t, f)
	assert.Same(t, fab, f, "resume must ride the same host-owned fabric, not create one")
	assert.NoError(t, originates(t, f))
}

// The socket never owns the fabric's lifecycle, so closing it must never take
// the host's network down with it.
func TestClosingASocketNeverClosesTheSharedFabric(t *testing.T) {
	t.Parallel()
	fab := newFabric(t)
	s, err := socket.Open(context.Background(), fab, "octocat/hello")
	require.NoError(t, err)

	require.NoError(t, s.Close())
	assert.Same(t, fab, s.Fabric(), "Fabric() must stay valid after Close")
	assert.NoError(t, originates(t, fab), "the host-owned fabric must survive the socket's Close")
}

func TestATicketRedeemsOnlyOnce(t *testing.T) {
	t.Parallel()
	fab := newFabric(t)
	s, err := socket.Open(context.Background(), fab, "octocat/hello")
	require.NoError(t, err)

	ticket, err := s.Park()
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })

	s2, err := socket.Resume(context.Background(), ticket)
	require.NoError(t, err)
	t.Cleanup(func() { _ = s2.Close() })

	_, err = socket.Resume(context.Background(), ticket)
	assert.Error(t, err, "a second Resume on the same ticket must error")
}

func TestCloseIsIdempotentWhileListening(t *testing.T) {
	t.Parallel()
	fab := newFabric(t)
	s, err := socket.Open(context.Background(), fab, "octocat/hello")
	require.NoError(t, err)

	require.NoError(t, s.Close())
	assert.NoError(t, s.Close())
}

func TestCloseIsIdempotentWhileParked(t *testing.T) {
	t.Parallel()
	fab := newFabric(t)
	s, err := socket.Open(context.Background(), fab, "octocat/hello")
	require.NoError(t, err)

	_, err = s.Park()
	require.NoError(t, err)

	require.NoError(t, s.Close())
	assert.NoError(t, s.Close())
}

func TestParkAfterCloseIsRejected(t *testing.T) {
	t.Parallel()
	fab := newFabric(t)
	s, err := socket.Open(context.Background(), fab, "octocat/hello")
	require.NoError(t, err)
	require.NoError(t, s.Close())

	_, err = s.Park()
	assert.Error(t, err)
}

func TestParkingAnAlreadyParkedSocketIsRejected(t *testing.T) {
	t.Parallel()
	fab := newFabric(t)
	s, err := socket.Open(context.Background(), fab, "octocat/hello")
	require.NoError(t, err)

	ticket, err := s.Park()
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })

	_, err = s.Park()
	assert.Error(t, err)

	s2, err := socket.Resume(context.Background(), ticket)
	require.NoError(t, err)
	_ = s2.Close()
}

// Closing a parked socket invalidates its outstanding ticket, so a stale
// ticket can never resurrect an endpoint the owner already tore down.
func TestATicketFromAClosedSocketCannotResume(t *testing.T) {
	t.Parallel()
	fab := newFabric(t)
	s, err := socket.Open(context.Background(), fab, "octocat/hello")
	require.NoError(t, err)

	ticket, err := s.Park()
	require.NoError(t, err)
	require.NoError(t, s.Close())

	_, err = socket.Resume(context.Background(), ticket)
	assert.Error(t, err, "a ticket from a closed socket must not resume")
}
