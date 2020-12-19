package model

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
	"wander/helper"
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
type RandomInfo struct {
	Code        int    `json:"code"`
	Name        string `json:"name"`
	ArtistsName string `json:"artists_name"`
	MusicUrl    string `json:"music_url"`
	MusicPic    string `json:"music_pic"`
	AvatarUrl   string `json:"avatarurl"`
	NickName    string `json:"nickname"`
	Comments    string `json:"comments"`
}

type LinkInfo struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ID  int    `json:"id"`
		Url string `json:"url"`
	} `json:"data"`
}

type MusicInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ArtistsName   string `json:"artists_name"`
	MusicUrl      string `json:"music_url"`
	MusicPic      string `json:"music_pic"`
	MusicLocal    string `json:"music_local"`
	MusicPicLocal string `json:"music_pic_local"`
}

type CacheType uint

const (
	CachePic   CacheType = 1 << 0
	CacheMusic           = 1 << 1
	CacheAll             = CachePic | CacheMusic
)

var (
	PicExts = map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".bmp":  true,
		".gif":  true,
	}
	MusicExts = map[string]bool{
		".mp3": true,
		".wma": true,
		".wav": true,
	}
)

func CheckCaches(path, name string, typ CacheType) (map[CacheType]string, bool) {
	var (
		res      = map[CacheType]string{}
		fitTypes CacheType
	)
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if strings.Index(info.Name(), name) >= 0 {
			ext := filepath.Ext(info.Name())
			if typ&CachePic == CachePic && PicExts[ext] {
				res[CachePic] = path
				fitTypes |= CachePic
			}
			if typ&CacheMusic == CacheMusic && MusicExts[ext] {
				res[CacheMusic] = path
				fitTypes |= CacheMusic
			}
		}
		return nil
	})

	return res, fitTypes == typ
}

func Download(uri, split, fileName string) (string, error) {
	data, _, err := helper.HttpDoTimeout(nil, "GET", uri, nil, 30*time.Second)
	if err != nil {
		return "", err
	}
	name := uri[strings.LastIndex(uri, split)+1:]
	if fileName != "" {
		name = "cache/" + fileName + filepath.Ext(name)
	}
	err = ioutil.WriteFile(name, data, os.ModePerm)
	return name, err
}

func RequestNext() (*MusicInfo, error) {
	var (
		randomInfo = new(RandomInfo)
		linkInfo   = new(LinkInfo)
		musicInfo  = new(MusicInfo)
	)

	// get random randomInfo info
	data, _, err := helper.HttpDoTimeout(nil, "GET", RandomUrl, nil, 30*time.Second)
	//fmt.Println(code, err, string(data))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &randomInfo)
	if err != nil {
		return nil, err
	}
	if randomInfo.Code != 200 {
		return nil, fmt.Errorf("randomInfo http code err[%d]", randomInfo.Code)
	}

	musicInfo.ID = randomInfo.MusicUrl[strings.LastIndex(randomInfo.MusicUrl, "=")+1 : strings.LastIndex(randomInfo.MusicUrl, ".mp3")]
	musicInfo.Name = randomInfo.Name
	musicInfo.ArtistsName = randomInfo.ArtistsName
	musicInfo.MusicPic = randomInfo.MusicPic

	fileName := fmt.Sprintf("%s-%s", musicInfo.Name, musicInfo.ArtistsName)

	if res, ok := CheckCaches("cache", fileName, CacheAll); ok {
		musicInfo.MusicPicLocal = res[CachePic]
		musicInfo.MusicLocal = res[CacheMusic]
		return musicInfo, nil
	}

	// download music pic
	musicInfo.MusicPicLocal, err = Download(musicInfo.MusicPic, "/", fileName)
	if err != nil {
		return nil, err
	}

	// download music
	data, _, err = helper.HttpDoTimeout(nil, "GET", fmt.Sprintf(LinkUrl, musicInfo.ID), nil, 30*time.Second)
	//fmt.Println(code, err, string(data))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &linkInfo)
	if err != nil {
		return nil, err
	}
	if linkInfo.Code != 200 {
		return nil, fmt.Errorf("link http code err[%d]", linkInfo.Code)
	}
	musicInfo.MusicUrl = linkInfo.Data.Url
	musicInfo.MusicLocal, err = Download(musicInfo.MusicUrl, "/", fileName)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%+v, err: %v\n", musicInfo, err)

	return musicInfo, err
}
