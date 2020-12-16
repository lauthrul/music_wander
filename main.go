package main

import (
	"encoding/json"
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

/*
{
    "code": 200,
    "name": "可乐",
    "artists_name": "赵紫骅",
    "music_url": "http:\/\/music.163.com\/song\/media\/outer\/url?id=29759733.mp3",
    "music_pic": "http:\/\/p4.music.126.net\/qOfVT6izV4mBe4IyQn489Q==\/18190320370401891.jpg",
    "avatarurl": "https:\/\/p1.music.126.net\/YB49c5avmPR0rzesWrdFOg==\/109951164057116045.jpg",
    "nickname": "為妳我受冷風吹i",
    "comments": "九寨沟地震的时候，她被埋在了木楼下面，用最后的力气给我发了一条短信，说如果我活着出来你会娶我吗？当时我就泪流满面。我会的我一定会的，后来被救了出来，只有微弱的呼吸，我跪着哭着给医生说一定要医好她啊，悲痛万分的是一直器官衰竭没有醒过来，到今天还是离开了我2018.2.2凌晨早晨9.30分。"
}
*/
type Response struct {
	Code          int    `json:"code"`
	Name          string `json:"name"`
	ArtistsName   string `json:"artists_name"`
	MusicUrl      string `json:"music_url"`
	MusicPic      string `json:"music_pic"`
	AvatarUrl     string `json:"avatarurl"`
	NickName      string `json:"nickname"`
	Comments      string `json:"comments"`
	MusicLocal    string `json:"music_local"`
	MusicPicLocal string `json:"music_pic_local"`
	AvatarLocal   string `json:"avatar_local"`
}

const (
	URL = "https://api.66mz8.com/api/music.163.php?format=json"
)

var (
	btnPlay, btnNext, btnStop *walk.PushButton
	imgCover                  *walk.ImageView
	lbName                    *walk.Label
	lbAuthor                  *walk.Label
)

func Download(uri, split string) (string, error) {
	data, _, err := HttpDoTimeout(nil, "GET", uri, nil, 30*time.Second)
	if err != nil {
		return "", err
	}
	name := "cache/" + uri[strings.LastIndex(uri, split)+1:]
	err = ioutil.WriteFile(name, data, os.ModePerm)
	return name, err
}

func RequestNext() (*Response, error) {
	var resp Response
	data, _, err := HttpDoTimeout(nil, "GET", URL, nil, 30*time.Second)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &resp)

	// download music pic
	resp.MusicPicLocal, err = Download(resp.MusicPic, "/")
	if err != nil {
		return nil, err
	}

	// download music
	resp.MusicLocal, err = Download(resp.MusicUrl, "=")
	if err != nil {
		return nil, err
	}

	// download avatar
	resp.AvatarLocal, err = Download(resp.AvatarUrl, "/")
	if err != nil {
		return nil, err
	}

	fmt.Printf("%+v, err: %v\n", resp, err)

	return &resp, err
}

func main() {
	walk.Resources.SetRootDirPath("cache")

	MainWindow{
		Title:   "wander",
		MinSize: Size{Width: 200, Height: 100},
		//MaxSize: Size{Width: 200, Height: 100},
		Size:   Size{Width: 200, Height: 100},
		Layout: VBox{},
		Children: []Widget{
			ImageView{
				AssignTo: &imgCover,
				//Background: SolidColorBrush{Color: walk.RGB(255, 191, 0)},
				Image:   "img.jpg",
				MaxSize: Size{200, 200},
				MinSize: Size{200, 200},
				//Margin:  10,
				Mode: ImageViewModeZoom,
			},
			Label{
				AssignTo:  &lbName,
				Alignment: AlignHCenterVCenter,
				//Font:      Font{Family: "微软雅黑", Bold: true},
			},
			Label{
				AssignTo:  &lbAuthor,
				Alignment: AlignHCenterVCenter,
				//Font:      Font{Family: "微软雅黑"},
			},
			Composite{
				Layout: Grid{Columns: 4},
				Children: []Widget{
					PushButton{
						AssignTo: &btnPlay,
						Text:     "▶",
						OnClicked: func() {
							resp, err := RequestNext()
							if err != nil {
								fmt.Println("request next err:", err)
								return
							}
							img, err := walk.NewImageFromFile(resp.MusicPicLocal)
							if err != nil {
								fmt.Println("load music pic err:", err)
								return
							}
							imgCover.SetImage(img)
							lbName.SetText(resp.Name + " - " + resp.ArtistsName)
							lbAuthor.SetText(resp.NickName + ":\n" + resp.Comments)
						},
					},
					//PushButton{
					//	Text: "▎▎",
					//},
					PushButton{
						AssignTo: &btnNext,
						Text:     "▶▶",
					},
					PushButton{
						AssignTo: &btnStop,
						Text:     "■",
					},
				},
			},
		},
	}.Run()
}
