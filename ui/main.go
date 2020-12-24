package ui

import (
	"encoding/json"
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"time"
	"wander/log"
	"wander/model"
)

const (
	textPlay           = "▶"
	textPause          = "||"
	textPlayPrev       = "◀◀"
	textPlayNext       = "▶▶"
	textCurrentPlaying = "当前播放： <a>%s</a>"
)

type MyMainWindow struct {
	*walk.MainWindow

	// ui
	lbPlayList        *walk.ListBox
	lbTrackList       *walk.ListBox
	lblCurrentPlaying *walk.LinkLabel
	imgCover          *walk.ImageView
	lblName           *walk.Label
	slv               *walk.Slider
	btnPrev           *walk.PushButton
	btnPlay           *walk.PushButton
	btnNext           *walk.PushButton

	// data
	playList  *PlaylistModel
	musicList *TrackModel

	// manager
	pm         *model.PlayerManager
	chPlayback chan model.PlayCallback
}

func (mw *MyMainWindow) init() {
	go func() {
		for {
			select {
			case playback := <-mw.chPlayback:
				log.Debug(playback)
				switch playback.Action {
				case model.ActionStop:
					// DO NOTHING
				case model.ActionPlay, model.ActionPause:
					text := textPause
					if playback.Action == model.ActionPause {
						text = textPlay
					}
					mw.btnPlay.SetText(text)
				case model.ActionNext:
					mw.onPlayNext()
				}
			case <-time.After(time.Second):
				if !mw.pm.IsPlaying() {
					continue
				}

				current := mw.pm.Current()
				name := fmt.Sprintf("%s - %s", current.Name, current.ArtistsName)
				pos := mw.pm.Pos()
				duration := mw.pm.Len()
				text := fmt.Sprintf(textCurrentPlaying, name+fmt.Sprintf(" [%v/%v]", pos, duration))
				log.Debug(text)
				mw.lblCurrentPlaying.SetText(text)

				mw.slv.SetRange(0, current.Streamer.Len())
				mw.slv.SendMessage(win.TBM_SETPOS, 1, uintptr(current.Streamer.Position()))

				if pos >= duration {
					mw.onPlayNext()
				}
			}
		}
	}()
}

func (mw *MyMainWindow) updateControlPanel(music *model.MusicInfo) {
	img, err := walk.NewImageFromFile(music.MusicPicLocal)
	if err != nil {
		log.ErrorF("load music pic err:", err)
		return
	}
	mw.imgCover.SetImage(img)
	mw.lblName.SetText(music.Name + " - " + music.ArtistsName)

	if mw.pm.Current() != nil {
		if mw.pm.Current().MusicLocal == music.MusicLocal {
			mw.btnPlay.SetText(textPause)
		} else {
			mw.btnPlay.SetText(textPlay)
		}
	}
}

func (mw *MyMainWindow) onGotoTackList(link *walk.LinkLabelLink) {
	if mw.pm.Current() == nil {
		return
	}

	idx := -1
	for i, m := range mw.musicList.items {
		if m.ID == mw.pm.Current().ID {
			idx = i
			break
		}
	}
	mw.lbTrackList.SetCurrentIndex(idx)
}

func (mw *MyMainWindow) onPlaylistChanged() {
	mw.Synchronize(func() {
		idx := mw.lbPlayList.CurrentIndex()
		if idx < 0 || idx >= len(mw.playList.items) {
			return
		}
		item := mw.playList.items[idx]
		url := fmt.Sprintf(model.Playlist, item.ID)
		data, _, err := model.HttpDoTimeout(nil, "GET", url, nil, 30*time.Second)
		if err != nil {
			return
		}
		var playlist model.PlaylistResp
		err = json.Unmarshal(data, &playlist)
		if err != nil {
			return
		}
		if playlist.Code != 200 {
			log.Error(string(data))
			return
		}

		mw.musicList.items = model.WalkPlaylist(&playlist)
		mw.musicList.PublishItemsReset()
	})
}

func (mw *MyMainWindow) onTrackListChanged() {
	mw.Synchronize(func() {
		var err error
		idx := mw.lbTrackList.CurrentIndex()
		if idx < 0 || idx >= len(mw.musicList.items) {
			return
		}
		music := mw.musicList.items[idx]
		fileName := fmt.Sprintf("%s-%s", music.Name, music.ArtistsName)
		res, ok := model.CheckCaches("cache", fileName, model.CachePic)
		if ok {
			music.MusicPicLocal = res[model.CachePic]
		} else {
			// download music pic
			music.MusicPicLocal, err = model.Download(music.MusicPic, "/", fileName)
			if err != nil {
				return
			}
		}
		mw.updateControlPanel(music)
	})
}

func (mw *MyMainWindow) play(idx int) {
	if idx < 0 || idx > len(mw.musicList.items) {
		log.ErrorF("playlist idx err:", idx)
		return
	}
	music := mw.musicList.items[idx]
	if music.MusicLocal == "" {
		fileName := fmt.Sprintf("%s-%s", music.Name, music.ArtistsName)
		res, ok := model.CheckCaches("cache", fileName, model.CacheMusic)
		if ok {
			music.MusicLocal = res[model.CacheMusic]
		} else {
			// download music
			link := fmt.Sprintf(model.LinkUrl, music.ID)
			data, _, err := model.HttpDoTimeout(nil, "GET", link, nil, 2*time.Minute)
			if err != nil {
				return
			}
			var linkInfo model.LinkInfo
			err = json.Unmarshal(data, &linkInfo)
			if err != nil {
				log.Error(err)
				return
			}
			if linkInfo.Code != 200 {
				log.Error(string(data))
				return
			}
			music.MusicUrl = linkInfo.Data.Url
			music.MusicLocal, err = model.Download(music.MusicUrl, "/", fileName)
			if err != nil {
				return
			}
		}
	}
	// play music
	action := model.Action(model.ActionPlay)
	if mw.pm.IsPlaying() {
		action = model.ActionPause
	}
	mw.pm.Play(music, action, -1)
}

func (mw *MyMainWindow) onPlayPrev() {
	mw.Synchronize(func() {
		mw.pm.Stop()
		idx := mw.lbTrackList.CurrentIndex() - 1
		if idx < 0 {
			idx = len(mw.musicList.items) - 1
		}
		mw.lbTrackList.SetCurrentIndex(idx)
		mw.play(idx)
	})
}

func (mw *MyMainWindow) onPlay() {
	mw.Synchronize(func() {
		mw.play(mw.lbTrackList.CurrentIndex())
	})
}

func (mw *MyMainWindow) onPlayNext() {
	mw.Synchronize(func() {
		mw.pm.Stop()
		idx := mw.lbTrackList.CurrentIndex() + 1
		max := len(mw.musicList.items) - 1
		if idx > max {
			idx = 0
		}
		mw.lbTrackList.SetCurrentIndex(idx)
		mw.play(idx)
	})
}

func (mw *MyMainWindow) onPlayPos() {
	mw.pm.Play(mw.pm.Current(), model.ActionPlay, mw.slv.Value())
}

func Run() {

	walk.Resources.SetRootDirPath("cache")

	mw := &MyMainWindow{
		playList:   NewPlaylist(),
		chPlayback: make(chan model.PlayCallback),
	}
	mw.musicList = NewTrackList(mw)
	mw.pm = model.NewPlayerManager(mw.chPlayback)

	mw.init()

	MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "wander",
		MinSize:  Size{Width: 500, Height: 300},
		MaxSize:  Size{Width: 500, Height: 300},
		Size:     Size{Width: 500, Height: 300},
		Layout:   HBox{},
		//MenuItems: []MenuItem{
		//	Action{
		//		Shortcut:    Shortcut{walk.ModControl|walk.ModAlt, walk.KeyLeft},
		//		OnTriggered: mw.onPlayPrev,
		//	},
		//	Action{
		//		Shortcut:    Shortcut{walk.ModControl|walk.ModAlt, walk.KeyReturn},
		//		OnTriggered: mw.onPlay,
		//	},
		//	Action{
		//		Shortcut:    Shortcut{walk.ModControl|walk.ModAlt, walk.KeyRight},
		//		OnTriggered: mw.onPlayNext,
		//	},
		//},
		Children: []Widget{
			HSplitter{
				//MinSize: Size{Width: 300},
				//MaxSize: Size{Width: 300},
				Children: []Widget{
					// 播放列表
					ListBox{
						AssignTo:              &mw.lbPlayList,
						MinSize:               Size{Width: 100},
						MaxSize:               Size{Width: 100},
						Model:                 mw.playList,
						CurrentIndex:          0,
						OnCurrentIndexChanged: mw.onPlaylistChanged,
					},
					// 歌单
					ListBox{
						AssignTo: &mw.lbTrackList,
						MinSize:  Size{Width: 200, Height: 32},
						//MaxSize:  Size{Width: 200, Height: 32},
						Model:                 mw.musicList,
						OnCurrentIndexChanged: mw.onTrackListChanged,
					},
				},
			},
			Composite{
				Layout: VBox{},
				Children: []Widget{
					LinkLabel{
						AssignTo:        &mw.lblCurrentPlaying,
						Text:            "当前播放：",
						OnLinkActivated: mw.onGotoTackList,
					},
					ImageView{
						AssignTo: &mw.imgCover,
						//Background: SolidColorBrush{Color: walk.RGB(0, 0, 0)},
						Image: "img.jpg",
						//MaxSize: Size{Width: 200, Height: 200},
						MinSize: Size{Width: 200, Height: 200},
						//Margin:  10,
						Mode: ImageViewModeZoom,
					},
					Label{
						AssignTo:  &mw.lblName,
						Alignment: AlignHCenterVCenter,
						//Font:      Font{Family: "微软雅黑", Bold: true},
						Text: "音乐的力量",
					},
					Slider{
						AssignTo:       &mw.slv,
						Orientation:    Horizontal,
						OnValueChanged: mw.onPlayPos,
					},
					Composite{
						MaxSize: Size{0, 32},
						Layout:  Grid{Columns: 4},
						Children: []Widget{
							PushButton{
								AssignTo:  &mw.btnPrev,
								Text:      textPlayPrev,
								OnClicked: mw.onPlayPrev,
							},
							PushButton{
								AssignTo:  &mw.btnPlay,
								Text:      textPlay,
								OnClicked: mw.onPlay,
							},
							PushButton{
								AssignTo:  &mw.btnNext,
								Text:      textPlayNext,
								OnClicked: mw.onPlayNext,
							},
						},
					},
				},
			},
		},
	}.Run()
}
