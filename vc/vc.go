package vc

import (
	"context"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type VC struct {
	vc    *discordgo.VoiceConnection
	guild string
	rw    *sync.RWMutex
}

func NewVC(guild string) *VC {
	return &VC{
		guild: guild,
		rw:    &sync.RWMutex{},
	}
}

func (v *VC) GetChannelID() string {
	// TODO: actually return the channel ID
	return ""
}

func (v *VC) Disconnect() {
	if !v.isConnectionNil() {
		v.rw.Lock()
		defer v.rw.Unlock()

		_ = v.vc.Disconnect(context.Background())

		v.vc = nil
	}
}

func (v *VC) isConnectionNil() bool {
	v.rw.RLock()
	defer v.rw.RUnlock()

	return v.vc == nil
}

func (v *VC) IsConnected() bool {
	if v.isConnectionNil() {
		return false
	}

	return v.vc.Status == discordgo.VoiceConnectionStatusReady
}

func (v *VC) GetIsDeadChannel() <-chan struct{} {
	return v.vc.Dead
}

func (v *VC) GetCond() *sync.Cond {
	return v.vc.Cond
}

func (v *VC) Join(s *discordgo.Session, channelID string) error {
	v.rw.Lock()
	defer v.rw.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	vc, err := s.ChannelVoiceJoin(ctx, v.guild, channelID, false, true)
	cancel()
	if err != nil {
		return err
	}

	v.vc = vc

	return nil
}

func (v *VC) Reconnect(s *discordgo.Session) error {
	channel := v.GetChannelID()
	return v.Join(s, channel)
}

func (v *VC) GetAudioChannel() chan []byte {
	if !v.isConnectionNil() {
		return v.vc.OpusSend
	} else {
		return nil
	}
}

func (v *VC) ChangeChannel(s *discordgo.Session, id string) error {
	var err error

	if v.GetChannelID() != id {
		v.rw.Lock()
		defer v.rw.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		v.vc.Disconnect(ctx)
		cancel()

		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		v.vc, err = s.ChannelVoiceJoin(ctx, v.guild, id, false, true)
		cancel()
	}

	return err
}

func (v *VC) GetDeadChannel() <-chan struct{} {
	return v.vc.Dead
}

func (v *VC) SetSpeaking(speaking bool) error {
	return v.vc.Speaking(speaking)
}
