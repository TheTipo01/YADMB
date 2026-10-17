package vc

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
)

var (
	// ErrNotConnected is returned when an operation needs an active voice connection.
	ErrNotConnected = errors.New("voice: not connected")
	// ErrAlreadyConnected is returned when the bot is already connected to another channel.
	ErrAlreadyConnected = errors.New("voice: already connected to another channel")
)

const (
	joinTimeout  = 10 * time.Second
	closeTimeout = 5 * time.Second
	opTimeout    = 5 * time.Second
)

// VC is a thread-safe wrapper around a disgo voice connection.
type VC struct {
	guildID snowflake.ID

	// opMu serializes operations that create or destroy a connection.
	// It is never held while calling into a caller-provided context-bound
	// playback path, so it cannot block the playback loop forever.
	opMu sync.Mutex

	// mu protects the fields below.
	mu          sync.RWMutex
	conn        voice.Conn
	ready       bool
	lastChannel snowflake.ID
}

func NewVC(guild snowflake.ID) *VC {
	return &VC{guildID: guild}
}

// GuildID returns the guild this connection belongs to.
func (v *VC) GuildID() snowflake.ID {
	return v.guildID
}

// IsConnected reports whether the connection is usable right now.
// The state is derived from disgo instead of being cached in a separate bool,
// so an external disconnect is reflected immediately.
func (v *VC) IsConnected() bool {
	_, ok := v.liveConn()
	return ok
}

// ChannelID returns the channel the bot is currently connected to, if any.
func (v *VC) ChannelID() (snowflake.ID, bool) {
	conn, ok := v.liveConn()
	if !ok {
		return 0, false
	}

	channelID := conn.ChannelID()
	if channelID == nil {
		return 0, false
	}

	return *channelID, true
}

// GatewayStatus returns the current voice gateway status.
func (v *VC) GatewayStatus() voice.Status {
	v.mu.RLock()
	conn := v.conn
	v.mu.RUnlock()

	if conn == nil {
		return voice.StatusUnconnected
	}

	return conn.Gateway().Status()
}

// Join joins channelID. Joining the channel the bot is already in is a no-op.
// Joining another channel while already connected returns ErrAlreadyConnected;
// the caller decides whether it is allowed to move the bot.
func (v *VC) Join(ctx context.Context, channelID snowflake.ID, manager voice.Manager) error {
	v.opMu.Lock()
	defer v.opMu.Unlock()

	if conn, ok := v.liveConn(); ok {
		if current := conn.ChannelID(); current != nil && *current == channelID {
			return nil
		}
		return ErrAlreadyConnected
	}

	return v.joinLocked(ctx, channelID, manager)
}

// Reconnect closes the current connection, if any, and joins the last known
// channel again. It is used by the playback loop after a voice write failure.
func (v *VC) Reconnect(ctx context.Context, manager voice.Manager) error {
	v.opMu.Lock()
	defer v.opMu.Unlock()

	v.mu.RLock()
	conn := v.conn
	channelID := v.lastChannel
	v.mu.RUnlock()

	if conn != nil {
		if current := conn.ChannelID(); current != nil {
			channelID = *current
		}

		v.closeConn(conn)
	}

	v.mu.Lock()
	v.conn = nil
	v.ready = false
	v.mu.Unlock()

	if channelID == 0 {
		return ErrNotConnected
	}

	return v.joinLocked(ctx, channelID, manager)
}

// Disconnect closes the voice connection and forgets the last channel.
func (v *VC) Disconnect(ctx context.Context) {
	v.close(ctx, true)
}

// Close closes the voice connection but keeps the last channel, so a later
// Reconnect can rejoin it. It is used for the idle-disconnect timer and for
// transient gateway failures.
func (v *VC) Close(ctx context.Context) {
	v.close(ctx, false)
}

func (v *VC) close(ctx context.Context, clearChannel bool) {
	v.opMu.Lock()
	defer v.opMu.Unlock()

	v.mu.Lock()
	conn := v.conn
	v.conn = nil
	v.ready = false
	if clearChannel {
		v.lastChannel = 0
	}
	v.mu.Unlock()

	if conn != nil {
		v.closeConn(conn)
	}
}

// UDP returns the UDP connection used to send audio.
func (v *VC) UDP() (voice.UDPConn, error) {
	conn, ok := v.liveConn()
	if !ok {
		return nil, ErrNotConnected
	}

	return conn.UDP(), nil
}

// SetSpeaking updates the speaking state of the bot.
func (v *VC) SetSpeaking(ctx context.Context, flag voice.SpeakingFlags) error {
	conn, ok := v.liveConn()
	if !ok {
		return ErrNotConnected
	}

	ctx, cancel := withTimeout(ctx, opTimeout)
	defer cancel()

	return conn.SetSpeaking(ctx, flag)
}

func (v *VC) joinLocked(ctx context.Context, channelID snowflake.ID, manager voice.Manager) error {
	if manager == nil {
		return errors.New("voice: nil manager")
	}

	v.mu.RLock()
	oldConn := v.conn
	v.mu.RUnlock()

	if oldConn != nil {
		v.closeConn(oldConn)
	}

	conn := manager.CreateConn(v.guildID)

	v.mu.Lock()
	v.conn = conn
	v.ready = false
	v.lastChannel = channelID
	v.mu.Unlock()

	joinCtx, cancel := withTimeout(ctx, joinTimeout)
	defer cancel()

	if err := conn.Open(joinCtx, channelID, false, true); err != nil {
		v.closeConn(conn)

		v.mu.Lock()
		if v.conn == conn {
			v.conn = nil
			v.ready = false
		}
		v.mu.Unlock()

		return fmt.Errorf("voice: join channel %s: %w", channelID, err)
	}

	if conn.ChannelID() == nil {
		v.closeConn(conn)

		v.mu.Lock()
		if v.conn == conn {
			v.conn = nil
			v.ready = false
		}
		v.mu.Unlock()

		return errors.New("voice: connection did not report a channel")
	}

	v.mu.Lock()
	if v.conn == conn {
		v.ready = true
		v.lastChannel = channelID
	}
	v.mu.Unlock()

	return nil
}

func (v *VC) liveConn() (voice.Conn, bool) {
	v.mu.RLock()
	conn, ready := v.conn, v.ready
	v.mu.RUnlock()

	if !ready || conn == nil {
		return nil, false
	}

	if conn.ChannelID() == nil {
		return nil, false
	}

	switch conn.Gateway().Status() {
	case voice.StatusDisconnected, voice.StatusUnconnected:
		return nil, false
	}

	return conn, true
}

func (v *VC) closeConn(conn voice.Conn) {
	ctx, cancel := context.WithTimeout(context.Background(), closeTimeout)
	defer cancel()

	conn.Close(ctx)
}

func withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}

	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, timeout)
}
