package fabric_test

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joaomdsg/packets/internal/fabric"
)

func TestBind_resolvesPortZeroToAConcreteAddr(t *testing.T) {
	t.Parallel()
	b, err := fabric.Bind("127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = b.Close() })

	assert.NotContains(t, b.Addr(), ":0")
	assert.Contains(t, b.Addr(), "127.0.0.1:")
}

func TestBind_rejectsAMalformedAddr(t *testing.T) {
	t.Parallel()
	for _, addr := range []string{"no-port", "127.0.0.1:notaport"} {
		_, err := fabric.Bind(addr)
		assert.Error(t, err, "addr %q must be rejected", addr)
	}
}

func TestBinding_listenServesFabricOnTheBoundAddr(t *testing.T) {
	t.Parallel()
	b, err := fabric.Bind("127.0.0.1:0")
	require.NoError(t, err)
	want := b.Addr()

	f, err := b.Listen(context.Background(), t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	assert.Equal(t, want, f.Addr())

	_, err = f.Publish(context.Background(), fabric.EventSubject("A", "i", fabric.StatusMinted, "catch"), []byte("m"))
	assert.NoError(t, err, "the in-process host retains full minting publish after Bind+Listen")
}

func TestBinding_listenPreservesPeerAuthConfinement(t *testing.T) {
	t.Parallel()
	b, err := fabric.Bind("127.0.0.1:0")
	require.NoError(t, err)

	f, err := b.Listen(context.Background(), t.TempDir(),
		fabric.Grant{User: "prodA", Pass: "pwA", Session: "A", Instance: "i"})
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	pc, err := nats.Connect(f.Addr(), nats.UserInfo("prodA", "pwA"))
	require.NoError(t, err)
	defer pc.Close()
	pjs, err := pc.JetStream()
	require.NoError(t, err)

	_, err = pjs.Publish(fabric.EventSubject("A", "i", fabric.StatusClaim, "diff"), []byte("x"))
	assert.NoError(t, err, "a peer may publish its own claim subtree through Bind+Listen")

	_, err = pjs.Publish(fabric.EventSubject("A", "i", fabric.StatusMinted, "catch"), []byte("x"))
	assert.Error(t, err, "a peer must not mint through Bind+Listen")
}

func TestBinding_secondListenOnAConsumedBindingErrors(t *testing.T) {
	t.Parallel()
	b, err := fabric.Bind("127.0.0.1:0")
	require.NoError(t, err)

	f, err := b.Listen(context.Background(), t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	_, err = b.Listen(context.Background(), t.TempDir())
	assert.Error(t, err, "a second Listen on an already-consumed binding must error")
}

func TestBinding_closeReleasesTheAddrForRebindingAndBlocksLaterListen(t *testing.T) {
	t.Parallel()
	b, err := fabric.Bind("127.0.0.1:0")
	require.NoError(t, err)
	addr := b.Addr()

	require.NoError(t, b.Close())

	b2, err := fabric.Bind(addr)
	require.NoError(t, err, "the released addr must be rebindable")
	t.Cleanup(func() { _ = b2.Close() })

	_, err = b.Listen(context.Background(), t.TempDir())
	assert.Error(t, err, "Listen after Close must error")
}
