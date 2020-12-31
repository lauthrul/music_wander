package model

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/lauthrul/goutil/log"
	"os"
	"time"
)

type MusicInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ArtistsName   string `json:"artists_name"`
	MusicUrl      string `json:"music_url"`
	MusicPic      string `json:"music_pic"`
	MusicLocal    string `json:"music_local"`
	MusicPicLocal string `json:"music_pic_local"`
}

type MusicController struct {
	Streamer beep.StreamSeekCloser
	Format   beep.Format
	Ctrl     *beep.Ctrl
}

type Music struct {
	Info       MusicInfo
	controller MusicController
}

func (m *Music) IsInit() bool {
	return m.controller.Streamer != nil && m.controller.Ctrl != nil
}

func (m *Music) Init(stream beep.StreamSeekCloser, format beep.Format, ctrl *beep.Ctrl) error {
	if m.IsInit() {
		return fmt.Errorf("already init")
	}
	m.controller = MusicController{
		Streamer: stream,
		Format:   format,
		Ctrl:     ctrl,
	}
	return nil
}

func (m *Music) IsPlaying() bool {
	return m.IsInit() && !m.controller.Ctrl.Paused
}

func (m *Music) Seek(pos int) error {
	if !m.IsInit() {
		return fmt.Errorf("not init")
	}
	return m.controller.Streamer.Seek(pos)
}

func (m *Music) SetPause(pause bool) {
	if !m.IsInit() {
		fmt.Errorf("not init")
	}
	m.controller.Ctrl.Paused = pause
}

func (m *Music) Play(playCtrl playCtrl) {
	if !m.IsInit() {
		f, err := os.Open(m.Info.MusicLocal)
		if err != nil {
			log.Error(err)
		}

		streamer, format, err := mp3.Decode(f)
		if err != nil {
			log.Error(err)
		}

		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		ctrl := &beep.Ctrl{Streamer: streamer, Paused: false}
		speaker.Play(ctrl)

		m.Init(streamer, format, ctrl)
	}

	speaker.Lock()
	if playCtrl.action == ActionPlay {
		if playCtrl.pos > 0 {
			m.Seek(playCtrl.pos)
		}
		m.SetPause(false)
	} else if playCtrl.action == ActionPause {
		m.SetPause(true)
	}
	speaker.Unlock()
}

func (m *Music) Stop() {
	if m.IsInit() {
		m.controller.Streamer.Close()
		m.controller.Streamer = nil
		m.controller.Ctrl.Streamer = nil
		m.controller.Ctrl = nil
		m.controller.Format = beep.Format{}
	}
}

func (m *Music) Pos() int {
	if !m.IsInit() {
		return -1
	}
	return m.controller.Streamer.Position()
}

func (m *Music) Len() int {
	if !m.IsInit() {
		return -1
	}
	return m.controller.Streamer.Len()
}

func (m *Music) Duration(pos int) time.Duration {
	if !m.IsInit() {
		return -1
	}
	return m.controller.Format.SampleRate.D(pos).Round(time.Second)
}
