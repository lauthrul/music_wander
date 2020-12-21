package model

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"os"
	"time"
	"wander/log"
)

type Action uint

const (
	ActionStop  Action = 0
	ActionPlay         = 1
	ActionPause        = 2
	ActionNext         = 3
)

type PlayAction struct {
	MusicInfo
	Action
}

type PlayCtrl struct {
	*MusicInfo
	pos   int
}

type PlayerManager struct {
	currentMusic   *MusicInfo
	ctrlCh         chan PlayCtrl   // 内部播放控制chan
	cbPlayActionCh chan PlayAction // 播放控制回调chan
}

func NewPlayerManager(ch chan PlayAction) *PlayerManager {
	pm := &PlayerManager{ctrlCh: make(chan PlayCtrl), cbPlayActionCh: ch}
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
			pm.cbPlayActionCh <- PlayAction{MusicInfo: *pm.currentMusic, Action: ActionStop}
			pm.currentMusic = nil
		}
		pm.currentMusic = music
	}

	if music.Streamer == nil {
		f, err := os.Open(music.MusicLocal)
		if err != nil {
			log.Error(err)
		}

		streamer, format, err := mp3.Decode(f)
		if err != nil {
			log.Error(err)
		}

		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

		ctrl := &beep.Ctrl{Streamer: streamer, Paused: false}
		cb := beep.Seq(ctrl, beep.Callback(func() {
			//pm.cbPlayActionCh <- ActionNext
		}))
		speaker.Play(cb)

		music.Streamer = streamer
		music.Format = format
		music.Ctrl = ctrl

		pm.cbPlayActionCh <- PlayAction{MusicInfo: *music, Action: ActionPlay}

	} else {
		speaker.Lock()
		music.Ctrl.Paused = !music.Ctrl.Paused
		speaker.Unlock()

		action := Action(ActionPlay)
		if music.Ctrl.Paused {
			action = ActionPause
		}
		pm.cbPlayActionCh <- PlayAction{MusicInfo: *music, Action: action}
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

func (pm *PlayerManager) PlayPos() {
	pm.ctrlCh <- music
}

func (pm *PlayerManager) Pos() time.Duration {
	return pm.duration(pm.currentMusic.Streamer.Position())
}

func (pm *PlayerManager) Len() time.Duration {
	return pm.duration(pm.currentMusic.Streamer.Len())
}
