package vc

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
)

type fakeGateway struct {
	voice.Gateway

	mu     sync.Mutex
	status voice.Status
}

func (g *fakeGateway) Status() voice.Status {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.status
}

type fakeConn struct {
	voice.Conn

	mu        sync.Mutex
	channelID snowflake.ID
	status    voice.Status
	openErr   error
	openFn    func(ctx context.Context) error
	closed    int
}

func (c *fakeConn) Open(ctx context.Context, channelID snowflake.ID, _, _ bool) error {
	if c.openFn != nil {
		if err := c.openFn(ctx); err != nil {
			return err
		}
	}

	if c.openErr != nil {
		return c.openErr
	}

	c.mu.Lock()
	c.channelID = channelID
	c.status = voice.StatusReady
	c.mu.Unlock()

	return nil
}

func (c *fakeConn) Close(context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed++
	c.channelID = 0
	c.status = voice.StatusDisconnected
}

func (c *fakeConn) ChannelID() *snowflake.ID {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channelID == 0 {
		return nil
	}

	channelID := c.channelID
	return &channelID
}

func (c *fakeConn) Gateway() voice.Gateway {
	c.mu.Lock()
	defer c.mu.Unlock()

	return &fakeGateway{status: c.status}
}

func (c *fakeConn) SetSpeaking(context.Context, voice.SpeakingFlags) error {
	return nil
}

func (c *fakeConn) externalDisconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.channelID = 0
	c.status = voice.StatusDisconnected
}

func (c *fakeConn) closedCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.closed
}

type fakeManager struct {
	voice.Manager

	mu      sync.Mutex
	factory func() *fakeConn
	created []*fakeConn
}

func (m *fakeManager) CreateConn(_ snowflake.ID) voice.Conn {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn := &fakeConn{}
	if m.factory != nil {
		conn = m.factory()
	}

	m.created = append(m.created, conn)
	return conn
}

func (m *fakeManager) conn(i int) *fakeConn {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.created[i]
}

func (m *fakeManager) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.created)
}

const (
	guildID  snowflake.ID = 1
	channelA snowflake.ID = 2
	channelB snowflake.ID = 3
)

func TestNotConnectedIsSafe(t *testing.T) {
	v := NewVC(guildID)

	if v.IsConnected() {
		t.Fatal("new VC should not be connected")
	}

	if _, ok := v.ChannelID(); ok {
		t.Fatal("new VC should not have a channel")
	}

	if _, err := v.UDP(); !errors.Is(err, ErrNotConnected) {
		t.Fatalf("expected ErrNotConnected from UDP, got %v", err)
	}

	if err := v.SetSpeaking(context.Background(), voice.SpeakingFlagMicrophone); !errors.Is(err, ErrNotConnected) {
		t.Fatalf("expected ErrNotConnected from SetSpeaking, got %v", err)
	}

	v.Disconnect(context.Background())
}

func TestJoinSameChannelAndWrongChannel(t *testing.T) {
	manager := &fakeManager{}
	v := NewVC(guildID)

	if err := v.Join(context.Background(), channelA, manager); err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	if !v.IsConnected() {
		t.Fatal("expected connected after Join")
	}

	channelID, ok := v.ChannelID()
	if !ok || channelID != channelA {
		t.Fatalf("expected channel %s, got %s (ok=%v)", channelA, channelID, ok)
	}

	if err := v.Join(context.Background(), channelA, manager); err != nil {
		t.Fatalf("joining the same channel should be a no-op: %v", err)
	}

	if err := v.Join(context.Background(), channelB, manager); !errors.Is(err, ErrAlreadyConnected) {
		t.Fatalf("expected ErrAlreadyConnected, got %v", err)
	}

	v.Disconnect(context.Background())

	if v.IsConnected() {
		t.Fatal("expected disconnected after Disconnect")
	}

	if manager.count() != 1 {
		t.Fatalf("expected one connection, got %d", manager.count())
	}

	if manager.conn(0).closedCount() != 1 {
		t.Fatalf("expected connection to be closed once, got %d", manager.conn(0).closedCount())
	}
}

func TestJoinTimeoutClosesConnection(t *testing.T) {
	manager := &fakeManager{factory: func() *fakeConn {
		return &fakeConn{
			openFn: func(ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			},
		}
	}}
	v := NewVC(guildID)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if err := v.Join(ctx, channelA, manager); err == nil {
		t.Fatal("expected Join to fail")
	}

	if v.IsConnected() {
		t.Fatal("expected disconnected after failed Join")
	}

	if manager.conn(0).closedCount() != 1 {
		t.Fatalf("expected connection to be closed after failed Join, got %d", manager.conn(0).closedCount())
	}
}

func TestJoinErrorClosesConnection(t *testing.T) {
	wantErr := errors.New("open failed")
	manager := &fakeManager{factory: func() *fakeConn {
		return &fakeConn{openErr: wantErr}
	}}
	v := NewVC(guildID)

	if err := v.Join(context.Background(), channelA, manager); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}

	if v.IsConnected() {
		t.Fatal("expected disconnected after failed Join")
	}

	if manager.conn(0).closedCount() != 1 {
		t.Fatalf("expected connection to be closed after failed Join, got %d", manager.conn(0).closedCount())
	}
}

func TestReconnectUsesLastChannel(t *testing.T) {
	manager := &fakeManager{}
	v := NewVC(guildID)

	if err := v.Join(context.Background(), channelA, manager); err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	if err := v.Reconnect(context.Background(), manager); err != nil {
		t.Fatalf("Reconnect failed: %v", err)
	}

	if manager.count() != 2 {
		t.Fatalf("expected a second connection, got %d", manager.count())
	}

	channelID, ok := v.ChannelID()
	if !ok || channelID != channelA {
		t.Fatalf("expected channel %s after reconnect, got %s (ok=%v)", channelA, channelID, ok)
	}

	if manager.conn(0).closedCount() != 1 {
		t.Fatalf("expected old connection to be closed, got %d", manager.conn(0).closedCount())
	}
}

func TestReconnectAfterExternalDisconnect(t *testing.T) {
	manager := &fakeManager{}
	v := NewVC(guildID)

	if err := v.Join(context.Background(), channelA, manager); err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	manager.conn(0).externalDisconnect()

	if v.IsConnected() {
		t.Fatal("expected external disconnect to be detected")
	}

	if err := v.Reconnect(context.Background(), manager); err != nil {
		t.Fatalf("Reconnect failed: %v", err)
	}

	if manager.count() != 2 {
		t.Fatalf("expected a second connection, got %d", manager.count())
	}

	channelID, ok := v.ChannelID()
	if !ok || channelID != channelA {
		t.Fatalf("expected channel %s after reconnect, got %s (ok=%v)", channelA, channelID, ok)
	}
}

func TestCloseKeepsLastChannelForReconnect(t *testing.T) {
	manager := &fakeManager{}
	v := NewVC(guildID)

	if err := v.Join(context.Background(), channelA, manager); err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	v.Close(context.Background())

	if v.IsConnected() {
		t.Fatal("expected connection to be closed")
	}

	if err := v.Reconnect(context.Background(), manager); err != nil {
		t.Fatalf("Reconnect after Close failed: %v", err)
	}

	channelID, ok := v.ChannelID()
	if !ok || channelID != channelA {
		t.Fatalf("expected channel %s after reconnect, got %s (ok=%v)", channelA, channelID, ok)
	}
}

func TestDisconnectClearsLastChannel(t *testing.T) {
	manager := &fakeManager{}
	v := NewVC(guildID)

	if err := v.Join(context.Background(), channelA, manager); err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	v.Disconnect(context.Background())

	if err := v.Reconnect(context.Background(), manager); !errors.Is(err, ErrNotConnected) {
		t.Fatalf("expected ErrNotConnected after explicit disconnect, got %v", err)
	}
}

func TestConcurrentJoinAndDisconnect(t *testing.T) {
	manager := &fakeManager{}
	v := NewVC(guildID)

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = v.Join(context.Background(), channelA, manager)
			_ = v.Join(context.Background(), channelA, manager)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			v.Disconnect(context.Background())
		}()
	}

	wg.Wait()
}
