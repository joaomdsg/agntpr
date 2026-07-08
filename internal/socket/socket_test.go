package socket_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joaomdsg/packets/internal/fabric"
	"github.com/joaomdsg/packets/internal/socket"
)

func TestOpen_startsListeningAndAcceptsPublish(t *testing.T) {
	t.Parallel()
	s, err := socket.Open(context.Background(), "127.0.0.1:0", t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })

	require.NotNil(t, s.Fabric())
	_, err = s.Fabric().Publish(context.Background(), fabric.EventSubject("A", "i", fabric.StatusMinted, "catch"), []byte("m"))
	assert.NoError(t, err)
}

func TestSocket_addrStaysStableAcrossParkAndResume(t *testing.T) {
	t.Parallel()
	s, err := socket.Open(context.Background(), "127.0.0.1:0", t.TempDir())
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

func TestSocket_parkReleasesLiveFabricButKeepsAddrClaimed(t *testing.T) {
	t.Parallel()
	s, err := socket.Open(context.Background(), "127.0.0.1:0", t.TempDir())
	require.NoError(t, err)
	addr := s.Addr()

	ticket, err := s.Park()
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })

	assert.Nil(t, s.Fabric())

	_, err = fabric.Bind(addr)
	assert.Error(t, err, "the addr must stay claimed by the parked socket's placeholder")

	s2, err := socket.Resume(context.Background(), ticket)
	require.NoError(t, err)
	_ = s2.Close()
}

func TestSocket_resumeRewarmsPublishOnTheSameAddr(t *testing.T) {
	t.Parallel()
	s, err := socket.Open(context.Background(), "127.0.0.1:0", t.TempDir())
	require.NoError(t, err)
	addr := s.Addr()

	ticket, err := s.Park()
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })

	s2, err := socket.Resume(context.Background(), ticket)
	require.NoError(t, err)
	t.Cleanup(func() { _ = s2.Close() })

	assert.Equal(t, addr, s2.Addr())
	_, err = s2.Fabric().Publish(context.Background(), fabric.EventSubject("A", "i", fabric.StatusMinted, "catch"), []byte("m"))
	assert.NoError(t, err)
}

func TestTicket_isSingleUse(t *testing.T) {
	t.Parallel()
	s, err := socket.Open(context.Background(), "127.0.0.1:0", t.TempDir())
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

func TestSocket_closeIsIdempotentWhenListening(t *testing.T) {
	t.Parallel()
	s, err := socket.Open(context.Background(), "127.0.0.1:0", t.TempDir())
	require.NoError(t, err)

	require.NoError(t, s.Close())
	assert.NoError(t, s.Close())
}

func TestSocket_closeIsIdempotentWhenParked(t *testing.T) {
	t.Parallel()
	s, err := socket.Open(context.Background(), "127.0.0.1:0", t.TempDir())
	require.NoError(t, err)

	_, err = s.Park()
	require.NoError(t, err)

	require.NoError(t, s.Close())
	assert.NoError(t, s.Close())
}

func TestSocket_parkAfterCloseErrors(t *testing.T) {
	t.Parallel()
	s, err := socket.Open(context.Background(), "127.0.0.1:0", t.TempDir())
	require.NoError(t, err)
	require.NoError(t, s.Close())

	_, err = s.Park()
	assert.Error(t, err)
}

func TestSocket_parkWhileAlreadyParkedErrors(t *testing.T) {
	t.Parallel()
	s, err := socket.Open(context.Background(), "127.0.0.1:0", t.TempDir())
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
