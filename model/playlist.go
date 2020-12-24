package model

import (
	"fmt"
	"strings"
)

type PlaylistResp struct {
	Code   int `json:"code"`
	Result struct {
		Tracks []struct {
			ID      int    `json:"id"`
			Name    string `json:"name"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
			Album struct {
				//Name   string `json:"name"`
				PicUrl string `json:"picUrl"`
			} `json:"album"`
		} `json:"tracks"`
	} `json:"result"`
}

func WalkPlaylist(playlist *PlaylistResp) []*Music {
	var musics []*Music
	for _, track := range playlist.Result.Tracks {
		music := &Music{
			Info: MusicInfo{
				ID:            fmt.Sprintf("%d", track.ID),
				Name:          track.Name,
				ArtistsName:   "",
				MusicUrl:      "",
				MusicPic:      track.Album.PicUrl,
				MusicLocal:    "",
				MusicPicLocal: "",
			},
		}
		for _, artist := range track.Artists {
			music.Info.ArtistsName += artist.Name + ","
		}
		music.Info.ArtistsName = strings.TrimRight(music.Info.ArtistsName, ",")
		musics = append(musics, music)
	}
	return musics
}
