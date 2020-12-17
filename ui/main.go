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
	btnPlay     *walk.PushButton
	btnNext     *walk.PushButton
	btnStop     *walk.PushButton
	imgCover    *walk.ImageView
	lblName     *walk.Label
	lbPlayList  *walk.ListBox
	lbTrackList *walk.ListBox

	// data
	playList  *PlaylistModel
	musicList *TrackModel
}

func (mw *MyMainWindow) onPlaylistChanged() {
	item := mw.playList.items[mw.lbPlayList.CurrentIndex()]
	url := fmt.Sprintf(model.Playlist, item.ID)
	data, code, err := helper.HttpDoTimeout(nil, "GET", url, nil, 30*time.Second)
	fmt.Println(code, err, string(data))
	if err != nil {
		fmt.Printf("get playList[%s] fail, err:%s", url, err.Error())
		return
	}
	var playlist model.PlaylistResp
	err = json.Unmarshal(data, &playlist)
	if err != nil {
		return
	}
	if playlist.Code != 200 {
		fmt.Printf("get playList[%s] fail, code:%d", url, playlist.Code)
		return
	}
	mw.musicList.items = model.WalkPlaylist(&playlist)
	mw.musicList.PublishItemsReset()
}

func (mw *MyMainWindow) onTrackListChanged() {

}

func (mw *MyMainWindow) onPlayNext() {
	resp, err := model.RequestNext()
	if err != nil {
		fmt.Println("request next err:", err)
		return
	}
	img, err := walk.NewImageFromFile(resp.MusicPicLocal)
	if err != nil {
		fmt.Println("load music pic err:", err)
		return
	}
	mw.imgCover.SetImage(img)
	mw.lblName.SetText(resp.Name + " - " + resp.ArtistsName)
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
			// 播放列表
			HSplitter{
				//MinSize: Size{Width: 300},
				//MaxSize: Size{Width: 300},
				Children: []Widget{
					ListBox{
						AssignTo:              &mw.lbPlayList,
						MinSize:               Size{Width: 100},
						MaxSize:               Size{Width: 100},
						Model:                 mw.playList,
						OnCurrentIndexChanged: mw.onPlaylistChanged,
					},
					ListBox{
						AssignTo: &mw.lbTrackList,
						MinSize:  Size{Width: 200, Height: 32},
						//MaxSize:  Size{Width: 200, Height: 32},
						Model:                 mw.musicList,
						OnCurrentIndexChanged: mw.onTrackListChanged,
					},
				},
			},
			// 歌单
			VSplitter{
				Children: []Widget{
					ImageView{
						AssignTo: &mw.imgCover,
						//Background: SolidColorBrush{Color: walk.RGB(0, 0, 0)},
						Image:   "img.jpg",
						MaxSize: Size{Width: 200, Height: 200},
						MinSize: Size{Width: 200, Height: 200},
						//Margin:  10,
						Mode: ImageViewModeZoom,
					},
					Label{
						AssignTo:  &mw.lblName,
						Alignment: AlignHCenterVCenter,
						//Font:      Font{Family: "微软雅黑", Bold: true},
					},
					Composite{
						Layout: Grid{Columns: 4},
						Children: []Widget{
							PushButton{
								AssignTo:  &mw.btnPlay,
								Text:      "▶",
								OnClicked: mw.onPlayNext,
							},
							//PushButton{
							//	Text: "▎▎",
							//},
							PushButton{
								AssignTo: &mw.btnNext,
								Text:     "▶▶",
							},
							PushButton{
								AssignTo: &mw.btnStop,
								Text:     "■",
							},
						},
					},
				},
			},
		},
	}.Run()
}
