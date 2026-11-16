package ipmi

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/server"
	"github.com/bougou/go-ipmi/pkg/transport/udp"
	"github.com/metal-stack/go-hal/pkg/logger"
	"github.com/stretchr/testify/require"
)

// syncedBuffer is an io.Writer safe for concurrent reads from the test.
type syncedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func waitFor(t *testing.T, what string, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

// TestActivateSOL exercises the console wiring against the go-ipmi reference
// BMC served over UDP loopback: console output must reach the output stream,
// keystrokes must reach the system serial port, and closing the input must end
// the session and deactivate the payload.
func TestActivateSOL(t *testing.T) {
	fake := &mock.FakeConsoleConn{}

	halMock := mock.New()
	halMock.SetConsole(&mock.Console{Conn: fake})

	b := bmc.New(bmc.DeviceInfo{IPMIVersion: 0x20}, [16]byte{}, halMock, bmc.WithClock(clock.Real))
	b.SOL.SetReconnectPolicy(&bmc.DefaultReconnectPolicy)
	user, err := b.Users.Add(2, "ADMIN")
	require.NoError(t, err)
	user.SetPassword([]byte("ADMIN"))
	user.Enabled = true
	user.ChannelAccess[1] = bmc.UserChannelAccess{MaxPrivilege: bmc.PrivilegeLevelAdministrator, Enabled: true}

	pc, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	require.NoError(t, err)
	t.Cleanup(func() { _ = pc.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	srv := server.NewServer(b, udp.Wrap(pc))
	t.Cleanup(func() { _ = srv.Close() })
	go func() { _ = srv.Serve(ctx) }()

	addr := pc.LocalAddr().(*net.UDPAddr)
	tool, err := NewOutBand(addr.IP.String(), addr.Port, "ADMIN", "ADMIN", logger.New())
	require.NoError(t, err)
	client := tool.(*Client)
	t.Cleanup(func() { _ = client.Close() })

	inR, inW := io.Pipe()
	out := &syncedBuffer{}

	solErr := make(chan error, 1)
	go func() {
		solErr <- client.activateSOL(inR, out)
	}()

	waitFor(t, "SOL activation", func() bool {
		return b.SOL.ActiveSessionID(1) != 0
	})

	fake.FeedRX([]byte("login: "))
	waitFor(t, "console output at remote", func() bool {
		return bytes.Contains([]byte(out.String()), []byte("login: "))
	})

	_, err = inW.Write([]byte("root\r"))
	require.NoError(t, err)
	waitFor(t, "keystrokes at console", func() bool {
		return bytes.Contains([]byte(fake.TXString()), []byte("root\r"))
	})

	// Closing the input ends the session and deactivates the payload.
	require.NoError(t, inW.Close())
	select {
	case err := <-solErr:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("activateSOL did not return after input EOF")
	}

	waitFor(t, "payload deactivation", func() bool {
		return b.SOL.ActiveSessionID(1) == 0
	})
}
