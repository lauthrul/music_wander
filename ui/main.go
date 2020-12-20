package ui

import (
	"encoding/json"
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
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
	btnPrev           *walk.PushButton
	btnPlay           *walk.PushButton
	btnNext           *walk.PushButton

	// data
	playList  *PlaylistModel
	musicList *TrackModel

	// manager
	pm *model.PlayerManager
	ch chan model.PlayAction
}

func (mw *MyMainWindow) init() {
	go func() {
		for {
			select {
			case status := <-mw.ch:
				log.Debug(mw.pm.Current(), status)
				switch status {
				case model.PlayActionStop:
					// DO NOTHING
				case model.PlayActionPlay, model.PlayActionPause:
					text := textPause
					if status == model.PlayActionPause {
						text = textPlay
					}
					mw.btnPlay.SetText(text)
				case model.PlayActionNext:
					mw.onPlayNext()
				}
			case <-time.After(time.Second):
				if mw.pm.Current() != nil {
					if mw.pm.Current().Streamer != nil {
						name := fmt.Sprintf("%s - %s", mw.pm.Current().Name, mw.pm.Current().ArtistsName)
						pos := mw.pm.Pos()
						duration := mw.pm.Len()
						text := fmt.Sprintf(textCurrentPlaying, name+fmt.Sprintf(" [%v/%v]", pos, duration))
						log.Debug(text)
						mw.lblCurrentPlaying.SetText(text)
						if pos >= duration {
							mw.onPlayNext()
						}
					}
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
	mw.pm.Play(music)
}

func (mw *MyMainWindow) onPlayPrev() {
	mw.Synchronize(func() {
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
		idx := mw.lbTrackList.CurrentIndex() + 1
		max := len(mw.musicList.items) - 1
		if idx > max {
			idx = 0
		}
		mw.lbTrackList.SetCurrentIndex(idx)
		mw.play(idx)
	})
}

func Run() {

	walk.Resources.SetRootDirPath("cache")

	mw := &MyMainWindow{
		playList: NewPlaylist(),
		ch:       make(chan model.PlayAction),
	}
	mw.musicList = NewTrackList(mw)
	mw.pm = model.NewPlayerManager(mw.ch)

	mw.init()

	MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "wander",
		MinSize:  Size{Width: 500, Height: 300},
		MaxSize:  Size{Width: 500, Height: 300},
		Size:     Size{Width: 500, Height: 300},
		Layout:   HBox{},
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
