package vc

import (
	"bytes"
	"context"
	"sync"

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

	v.vc.Close(context.TODO())
	v.connected = false
}

func (v *VC) IsConnected() bool {
	v.rw.RLock()
	defer v.rw.RUnlock()

	return v.connected
}

func (v *VC) Join(c bot.Client, channelID snowflake.ID) error {
	v.rw.Lock()
	defer v.rw.Unlock()

	v.vc = c.VoiceManager().CreateConn(v.guild)
	err := v.vc.Open(context.TODO(), channelID, false, false)
	v.connected = true

	return err
}

func (v *VC) SetBuffer(buffer *bytes.Buffer) {
	v.rw.Lock()
	defer v.rw.Unlock()

	if v.connected {
		v.vc.SetOpusFrameProvider(voice.NewOpusReader(buffer))
	}
}

func (v *VC) SetSpeaking(flag voice.SpeakingFlags) error {
	v.rw.Lock()
	defer v.rw.Unlock()

	return v.vc.SetSpeaking(context.TODO(), flag)
}
