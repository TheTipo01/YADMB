package vc

import (
	"github.com/bwmarrin/discordgo"
	"sync"
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
	v.vc.RLock()
	defer v.vc.RUnlock()

	return v.vc.ChannelID
}

func (v *VC) Disconnect() {
	if !v.isConnectionNil() {
		v.rw.Lock()
		defer v.rw.Unlock()

		_ = v.vc.Disconnect()
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

	v.vc.RLock()
	defer v.vc.RUnlock()

	return v.vc.Ready
}

func (v *VC) Join(s *discordgo.Session, channelID string) error {
	v.rw.Lock()
	defer v.rw.Unlock()

	vc, err := s.ChannelVoiceJoin(v.guild, channelID, false, true)
	if err != nil {
		return err
	}

	v.vc = vc

	return nil
}

func (v *VC) Reconnect(s *discordgo.Session) error {
	channel := v.GetChannelID()
	if v.isConnectionNil() {
		return v.Join(s, channel)
	} else {
		v.rw.Lock()
		v.vc.Lock()
		defer v.rw.Unlock()
		defer v.vc.Unlock()

		return v.vc.ChangeChannel(channel, false, true)
	}
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

		_ = v.vc.Disconnect()
		v.vc, err = s.ChannelVoiceJoin(v.guild, id, false, true)
	}

	return err
}

func (v *VC) SetSpeaking(speaking bool) error {
	return v.vc.Speaking(speaking)
}
