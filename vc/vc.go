package vc

import (
	"context"
	"sync"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
)

type VC struct {
	vc        voice.Conn
	guild     snowflake.ID
	connected bool
	rw        *sync.RWMutex
}

func NewVC(guild snowflake.ID) *VC {
	return &VC{
		guild:     guild,
		connected: false,
		rw:        &sync.RWMutex{},
	}
}

func (v *VC) GetChannelID() *snowflake.ID {
	v.rw.RLock()
	defer v.rw.RUnlock()

	return v.vc.ChannelID()
}

func (v *VC) Disconnect() {
	v.rw.Lock()
	defer v.rw.Unlock()

	if v.vc != nil {
		v.vc.Close(context.TODO())
		v.vc = nil
		v.connected = false
	}
}

func (v *VC) IsConnected() bool {
	v.rw.RLock()
	defer v.rw.RUnlock()

	return v.connected
}

func (v *VC) Join(channelID snowflake.ID, c *bot.Client) error {
	v.rw.Lock()
	defer v.rw.Unlock()

	if v.vc == nil {
		v.vc = c.VoiceManager.CreateConn(v.guild)
	}
	conn := v.vc

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- conn.Open(ctx, channelID, false, true)
	}()

	var err error
	select {
	case err = <-errCh:
	case <-ctx.Done():
		err = ctx.Err()
	}

	v.connected = err == nil

	return err
}

func (v *VC) GetUDP() voice.UDPConn {
	v.rw.RLock()
	defer v.rw.RUnlock()

	return v.vc.UDP()
}

func (v *VC) SetSpeaking(flag voice.SpeakingFlags) error {
	v.rw.Lock()
	defer v.rw.Unlock()

	return v.vc.SetSpeaking(context.TODO(), flag)
}
