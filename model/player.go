package model

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"os"
	"sync"
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

type PlayCallback struct {
	Music  MusicInfo
	Action Action
}

type playCtrl struct {
	music  *MusicInfo
	action Action
	pos    int
}

type PlayerManager struct {
	lock           sync.Mutex
	currentMusic   *MusicInfo
	chPlayCtrl     chan playCtrl     // 内部播放控制chan
	chPlayCallback chan PlayCallback // 播放控制回调chan
}

func NewPlayerManager(ch chan PlayCallback) *PlayerManager {
	pm := &PlayerManager{chPlayCtrl: make(chan playCtrl), chPlayCallback: ch}
	pm.init()
	return pm
}

func (pm *PlayerManager) init() {
	go func() {
		for {
			select {
			case ctrl := <-pm.chPlayCtrl:
				pm.play(ctrl)
			}
		}
	}()
}

func (pm *PlayerManager) play(playCtrl playCtrl) {
	log.Debug("play:", playCtrl)
	if pm.currentMusic != playCtrl.music {
		pm.Stop()
		pm.currentMusic = playCtrl.music
	}

	if pm.currentMusic.Streamer == nil {
		f, err := os.Open(pm.currentMusic.MusicLocal)
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
			//pm.chPlayCallback <- ActionNext
		}))
		speaker.Play(cb)

		pm.currentMusic.Streamer = streamer
		pm.currentMusic.Format = format
		pm.currentMusic.Ctrl = ctrl
	}

	speaker.Lock()
	if playCtrl.action == ActionPlay {
		if playCtrl.pos > 0 {
			pm.currentMusic.Streamer.Seek(playCtrl.pos)
		}
		pm.currentMusic.Ctrl.Paused = false
	} else if playCtrl.action == ActionPause {
		pm.currentMusic.Ctrl.Paused = true
	}
	speaker.Unlock()

	pm.chPlayCallback <- PlayCallback{Music: *pm.currentMusic, Action: playCtrl.action}
}

func (pm *PlayerManager) duration(pos int) time.Duration {
	speaker.Lock()
	defer speaker.Unlock()
	return pm.currentMusic.Format.SampleRate.D(pos).Round(time.Second)
}

func (pm *PlayerManager) Current() *MusicInfo {
	return pm.currentMusic
}

func (pm *PlayerManager) IsPlaying() bool {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	if pm.currentMusic != nil && pm.currentMusic.Ctrl != nil && !pm.currentMusic.Ctrl.Paused {
		return true
	}
	return false
}

func (pm *PlayerManager) Play(music *MusicInfo, action Action, pos int) {
	log.Debug("Play:", music, action, pos)
	pm.chPlayCtrl <- playCtrl{
		music:  music,
		action: action,
		pos:    pos,
	}
}

func (pm *PlayerManager) Stop() {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	if pm.currentMusic != nil && pm.currentMusic.Streamer != nil {
		pm.currentMusic.Streamer.Close()
		pm.currentMusic.Ctrl.Streamer = nil
		pm.currentMusic.Ctrl = nil
		pm.currentMusic.Streamer = nil
		pm.chPlayCallback <- PlayCallback{Music: *pm.currentMusic, Action: ActionStop}
		pm.currentMusic = nil
	}
}

func (pm *PlayerManager) Pos() time.Duration {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	return pm.duration(pm.currentMusic.Streamer.Position())
}

func (pm *PlayerManager) Len() time.Duration {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	return pm.duration(pm.currentMusic.Streamer.Len())
}
