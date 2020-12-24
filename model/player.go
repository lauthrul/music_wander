package model

import (
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
	Music
	Action
}

type playCtrl struct {
	music  *Music
	action Action
	pos    int
}

type PlayerManager struct {
	music          *Music
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
	if pm.music != playCtrl.music {
		if pm.music != nil {
			pm.music.Stop()
		}
		pm.music = playCtrl.music
	}

	pm.music.Play(playCtrl)

	pm.chPlayCallback <- PlayCallback{Music: *pm.music, Action: playCtrl.action}
}

func (pm *PlayerManager) Info() MusicInfo {
	if pm.music == nil {
		return MusicInfo{}
	}
	return pm.music.Info
}

func (pm *PlayerManager) IsPlaying() bool {
	return pm.music != nil && pm.music.IsPlaying()
}

func (pm *PlayerManager) Play(music *Music, action Action, pos int) {
	if music == nil {
		music = pm.music
	}
	pm.chPlayCtrl <- playCtrl{
		music:  music,
		action: action,
		pos:    pos,
	}
}

func (pm *PlayerManager) Stop() {
	if pm.music == nil {
		return
	}
	pm.music.Stop()
	pm.chPlayCallback <- PlayCallback{Music: *pm.music, Action: ActionStop}
	pm.music = nil
}

func (pm *PlayerManager) Pos() int {
	if pm.music == nil {
		return -1
	}
	return pm.music.Pos()
}

func (pm *PlayerManager) Len() int {
	if pm.music == nil {
		return -1
	}
	return pm.music.Len()
}

func (pm *PlayerManager) Duration(pos int) time.Duration {
	if pm.music == nil {
		return -1
	}
	return pm.music.Duration(pos)
}
