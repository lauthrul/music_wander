package model

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"os"
	"time"
)

type PlayerManager struct {
	currentMusic   *MusicInfo
	ctrlCh         chan *MusicInfo // 内部播放控制chan
	cbPlayActionCh chan PlayAction // 播放控制回调chan
}

func NewPlayerManager(ch chan PlayAction) *PlayerManager {
	pm := &PlayerManager{ctrlCh: make(chan *MusicInfo), cbPlayActionCh: ch}
	pm.init()
	return pm
}

func (pm *PlayerManager) init() {
	go func() {
		for {
			select {
			case music := <-pm.ctrlCh:
				pm.play(music)
			}
		}
	}()
}

func (pm *PlayerManager) play(music *MusicInfo) {
	if pm.currentMusic != music {
		if pm.currentMusic != nil && pm.currentMusic.Streamer != nil {
			pm.currentMusic.Streamer.Close()
			pm.currentMusic.Ctrl.Streamer = nil
			pm.currentMusic.Ctrl = nil
			pm.currentMusic.Streamer = nil
			pm.currentMusic = nil
			pm.cbPlayActionCh <- PlayActionStop
		}
		pm.currentMusic = music
	}

	if music.Streamer == nil {
		f, err := os.Open(music.MusicLocal)
		if err != nil {
			fmt.Println(err)
		}

		streamer, format, err := mp3.Decode(f)
		if err != nil {
			fmt.Println(err)
		}

		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

		ctrl := &beep.Ctrl{Streamer: streamer, Paused: false}
		cb := beep.Seq(ctrl, beep.Callback(func() {
			//pm.cbPlayActionCh <- PlayActionNext
		}))
		speaker.Play(cb)

		music.Streamer = streamer
		music.Format = format
		music.Ctrl = ctrl

		pm.cbPlayActionCh <- PlayActionPlay

	} else {
		speaker.Lock()
		music.Ctrl.Paused = !music.Ctrl.Paused
		speaker.Unlock()

		action := PlayAction(PlayActionPlay)
		if music.Ctrl.Paused {
			action = PlayActionPause
		}
		pm.cbPlayActionCh <- action
	}
}

func (pm *PlayerManager) duration(pos int) time.Duration {
	speaker.Lock()
	d := pm.currentMusic.Format.SampleRate.D(pos).Round(time.Second)
	speaker.Unlock()
	return d
}

func (pm *PlayerManager) Current() *MusicInfo {
	return pm.currentMusic
}

func (pm *PlayerManager) Play(music *MusicInfo) {
	pm.ctrlCh <- music
}

func (pm *PlayerManager) Pos() time.Duration {
	return pm.duration(pm.currentMusic.Streamer.Position())
}

func (pm *PlayerManager) Len() time.Duration {
	return pm.duration(pm.currentMusic.Streamer.Len())
}
