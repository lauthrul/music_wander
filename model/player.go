package model

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"os"
	"time"
)

var (
	streamer     beep.StreamSeekCloser
	currentMusic MusicInfo
)

func CurrentMusic() MusicInfo {
	return currentMusic
}

func Play(music *MusicInfo) {
	if currentMusic.MusicLocal != music.MusicLocal {
		if currentMusic.MusicLocal != "" {
			//if music.Ctrl != nil {
			//	music.Ctrl.Streamer = nil
			//	music.Ctrl = nil
			//}
			streamer.Close()
			music.Ctrl = nil
		}
		currentMusic = *music
	}
	if music.Ctrl == nil {
		f, err := os.Open(music.MusicLocal)
		if err != nil {
			fmt.Println(err)
		}

		s, format, err := mp3.Decode(f)
		if err != nil {
			fmt.Println(err)
		}
		streamer = s

		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

		ctrl := &beep.Ctrl{Streamer: s, Paused: false}
		cb := beep.Seq(ctrl, beep.Callback(func() {
			music.PlayStatus <- PlayStatusFinished
		}))
		speaker.Play(cb)

		music.Ctrl = ctrl
	} else {
		speaker.Lock()
		music.Ctrl.Paused = !music.Ctrl.Paused
		speaker.Unlock()

		status := PlayStatus(PlayStatusPlaying)
		if music.Ctrl.Paused {
			status = PlayStatusPaused
		}
		music.PlayStatus <- status
	}
}
