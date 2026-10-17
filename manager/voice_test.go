package manager

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/TheTipo01/YADMB/queue"
	"github.com/TheTipo01/YADMB/vc"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
)

type testGateway struct {
	voice.Gateway
	mu     sync.Mutex
	status voice.Status
}

func (g *testGateway) Status() voice.Status {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.status
}

type testConn struct {
	voice.Conn
	mu        sync.Mutex
	channelID snowflake.ID
	status    voice.Status
}

func (c *testConn) Open(_ context.Context, channelID snowflake.ID, _, _ bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.channelID = channelID
	c.status = voice.StatusReady
	return nil
}

func (c *testConn) Close(context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.status = voice.StatusDisconnected
}

func (c *testConn) ChannelID() *snowflake.ID {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channelID == 0 {
		return nil
	}

	channelID := c.channelID
	return &channelID
}

func (c *testConn) Gateway() voice.Gateway {
	c.mu.Lock()
	defer c.mu.Unlock()

	return &testGateway{status: c.status}
}

type testManager struct {
	voice.Manager
	conn *testConn
}

func (m *testManager) CreateConn(snowflake.ID) voice.Conn {
	return m.conn
}

func newTestServer(t *testing.T) (*Server, *testConn) {
	t.Helper()

	conn := &testConn{}
	client := &bot.Client{VoiceManager: &testManager{conn: conn}}
	server := NewServer(snowflake.ID(1), &Clients{Discord: client})

	t.Cleanup(func() {
		server.VC.Disconnect(context.Background())
	})

	return server, conn
}

func TestEnsureVoiceMovesWhenIdle(t *testing.T) {
	server, _ := newTestServer(t)

	channelA := snowflake.ID(10)
	channelB := snowflake.ID(11)

	if err := server.EnsureVoice(channelA); err != nil {
		t.Fatalf("EnsureVoice(A) failed: %v", err)
	}

	if err := server.EnsureVoice(channelA); err != nil {
		t.Fatalf("EnsureVoice(A) again failed: %v", err)
	}

	if err := server.EnsureVoice(channelB); err != nil {
		t.Fatalf("EnsureVoice(B) while idle failed: %v", err)
	}

	channelID, ok := server.VC.ChannelID()
	if !ok || channelID != channelB {
		t.Fatalf("expected channel %s, got %s (ok=%v)", channelB, channelID, ok)
	}
}

func TestEnsureVoiceRefusesToMoveWhilePlaying(t *testing.T) {
	server, _ := newTestServer(t)

	channelA := snowflake.ID(10)
	channelB := snowflake.ID(11)

	if err := server.EnsureVoice(channelA); err != nil {
		t.Fatalf("EnsureVoice(A) failed: %v", err)
	}

	server.Queue.AddElements(queue.Element{Title: "song"})
	server.Started.Store(true)

	err := server.EnsureVoice(channelB)
	if !errors.Is(err, vc.ErrAlreadyConnected) {
		t.Fatalf("expected ErrAlreadyConnected, got %v", err)
	}

	channelID, ok := server.VC.ChannelID()
	if !ok || channelID != channelA {
		t.Fatalf("expected channel %s, got %s (ok=%v)", channelA, channelID, ok)
	}
}
