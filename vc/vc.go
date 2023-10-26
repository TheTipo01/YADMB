package vc

import (
	"github.com/bwmarrin/discordgo"
	"sync"
)

type VC struct {
	vc    *discordgo.VoiceConnection
	guild string
	s     *discordgo.Session
	rw    *sync.RWMutex
}

func NewVC(s *discordgo.Session, guild string) *VC {
	return &VC{
		s:     s,
		guild: guild,
		rw:    &sync.RWMutex{},
	}
}

func (v *VC) GetChannelID() string {
	v.vc.RLock()
	defer v.vc.RUnlock()

	return v.vc.ChannelID
}

func (v *VC) Disconnect() {
	if v.isConnectionNil() {
		_ = v.vc.Disconnect()

		v.rw.Lock()
		v.vc = nil
		v.rw.Unlock()
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

	v.vc.RLock()
	defer v.vc.RUnlock()

	return v.vc.Ready
}

func (v *VC) Join(channelID string) error {
	v.rw.Lock()
	defer v.rw.Unlock()

	vc, err := v.s.ChannelVoiceJoin(v.guild, channelID, false, true)
	if err != nil {
		return err
	}

	v.vc = vc

	return nil
}

func (v *VC) Reconnect() error {
	v.rw.Lock()
	defer v.rw.Unlock()

	vc, err := v.s.ChannelVoiceJoin(v.guild, v.vc.ChannelID, false, true)
	if err != nil {
		return err
	}

	v.vc = vc

	return nil
}

func (v *VC) SendAudioPacket(packet []byte) {
	if !v.isConnectionNil() {
		v.vc.OpusSend <- packet
	}
}

func (v *VC) ChangeChannel(id string) error {
	var err error

	v.vc.RLock()
	if v.vc.ChannelID != id {
		v.vc.RUnlock()

		_ = v.vc.Disconnect()
		v.vc, err = v.s.ChannelVoiceJoin(v.guild, id, false, true)
	} else {
		v.vc.RUnlock()
	}

	return err
}

func (v *VC) SetSpeaking(speaking bool) error {
	return v.vc.Speaking(speaking)
}
