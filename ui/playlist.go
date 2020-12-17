package ui

import "github.com/lxn/walk"

type PlaylistItem struct {
	ID   string
	Name string
}

type PlaylistModel struct {
	walk.ListModelBase
	items []PlaylistItem
}

func (m *PlaylistModel) ItemCount() int {
	return len(m.items)
}

func (m *PlaylistModel) Value(index int) interface{} {
	return m.items[index].Name
}

func NewPlaylist() *PlaylistModel {
	return &PlaylistModel{
		items: []PlaylistItem{
			{ID: "19723756", Name: "云音乐飙升榜"},
			{ID: "3779629", Name: "云音乐新歌榜"},
			{ID: "3778678", Name: "云音乐热歌榜"},
			{ID: "2250011882", Name: "抖音排行榜"},
		},
	}
}
