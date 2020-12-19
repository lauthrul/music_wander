package ui

import (
	"encoding/json"
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"time"
	"wander/helper"
	"wander/model"
)

type MyMainWindow struct {
	// ui
	*walk.MainWindow
	btnPrev     *walk.PushButton
	btnPlay     *walk.PushButton
	btnNext     *walk.PushButton
	imgCover    *walk.ImageView
	lblName     *walk.Label
	lbPlayList  *walk.ListBox
	lbTrackList *walk.ListBox

	// data
	playList  *PlaylistModel
	musicList *TrackModel
}

func (mw *MyMainWindow) updateMusicUI(music *model.MusicInfo) {
	img, err := walk.NewImageFromFile(music.MusicPicLocal)
	if err != nil {
		fmt.Println("load music pic err:", err)
		return
	}
	mw.imgCover.SetImage(img)
	mw.lblName.SetText(music.Name + " - " + music.ArtistsName)
}

func (mw *MyMainWindow) onPlaylistChanged() {
	mw.Synchronize(func() {
		item := mw.playList.items[mw.lbPlayList.CurrentIndex()]
		url := fmt.Sprintf(model.Playlist, item.ID)
		data, _, err := helper.HttpDoTimeout(nil, "GET", url, nil, 30*time.Second)
		if err != nil {
			//fmt.Printf("get playList[%s] fail, err:%s", url, err.Error())
			return
		}
		var playlist model.PlaylistResp
		err = json.Unmarshal(data, &playlist)
		if err != nil {
			return
		}
		if playlist.Code != 200 {
			//fmt.Printf("get playList[%s] fail, code:%d", url, playlist.Code)
			return
		}

		mw.musicList.items = model.WalkPlaylist(&playlist)
		mw.musicList.PublishItemsReset()
	})
}

func (mw *MyMainWindow) onTrackListChanged() {
	mw.Synchronize(func() {
		var err error
		music := mw.musicList.items[mw.lbTrackList.CurrentIndex()]
		fileName := fmt.Sprintf("%s-%s", music.Name, music.ArtistsName)
		res, ok := model.CheckCaches("cache", fileName, model.CachePic)
		if ok {
			music.MusicPicLocal = res[model.CachePic]
		} else {
			// download music pic
			music.MusicPicLocal, err = model.Download(music.MusicPic, "/", fileName)
			if err != nil {
				fmt.Printf("cache music pic[%s : %s] fail:%s", fileName, music.MusicPic, err.Error())
				return
			}
		}
		mw.updateMusicUI(music)
	})
}

func (mw *MyMainWindow) onPlayPrev() {
}

func (mw *MyMainWindow) onPlay() {
	mw.Synchronize(func() {
		music := mw.musicList.items[mw.lbTrackList.CurrentIndex()]
		fileName := fmt.Sprintf("%s-%s", music.Name, music.ArtistsName)
		res, ok := model.CheckCaches("cache", fileName, model.CacheMusic)
		if ok {
			music.MusicLocal = res[model.CacheMusic]
		} else {
			// download music
			link := fmt.Sprintf(model.LinkUrl, music.ID)
			data, _, err := helper.HttpDoTimeout(nil, "GET", link, nil, 2*time.Minute)
			//fmt.Println(code, err, string(data))
			if err != nil {
				//fmt.Printf("get music real link[%s : %s] fail:%s", fileName, link, err.Error())
				return
			}
			var linkInfo model.LinkInfo
			err = json.Unmarshal(data, &linkInfo)
			if err != nil {
				//fmt.Printf("parse music real link[%s : %s] fail:%s", fileName, link, err.Error())
				return
			}
			if linkInfo.Code != 200 {
				//fmt.Printf("music real link[%s : %s] http code err:%d", fileName, link, linkInfo.Code)
				return
			}
			music.MusicUrl = linkInfo.Data.Url
			music.MusicLocal, err = model.Download(music.MusicUrl, "/", fileName)
			if err != nil {
				//fmt.Printf("cache music [%s : %s] http code err:%d", fileName, music.MusicUrl, linkInfo.Code)
				return
			}
		}
		// play music
	})
}

func (mw *MyMainWindow) onPlayNext() {
	resp, err := model.RequestNext()
	if err != nil {
		fmt.Println("request next err:", err)
		return
	}
	mw.updateMusicUI(resp)
}

func Run() {

	mw := &MyMainWindow{
		playList: NewPlaylist(),
	}
	mw.musicList = NewTrackList(mw)

	walk.Resources.SetRootDirPath("cache")

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
						AssignTo: &mw.lbPlayList,
						MinSize:  Size{Width: 100},
						MaxSize:  Size{Width: 100},
						Model:    mw.playList,
						//CurrentIndex:          0,
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
								Text:      "◀◀",
								OnClicked: mw.onPlayPrev,
							},
							PushButton{
								AssignTo:  &mw.btnPlay,
								Text:      "▶ / ||",
								OnClicked: mw.onPlay,
							},
							PushButton{
								AssignTo:  &mw.btnNext,
								Text:      "▶▶",
								OnClicked: mw.onPlayNext,
							},
						},
					},
				},
			},
		},
	}.Run()
}
