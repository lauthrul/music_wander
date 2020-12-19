package ui

import (
	"fmt"
	"github.com/lxn/walk"
	"github.com/lxn/win"
	"syscall"
	"unsafe"
	"wander/model"
)

type TrackModel struct {
	walk.ListModelBase
	items []*model.MusicInfo
}

func (m *TrackModel) ItemCount() int {
	return len(m.items)
}

func (m *TrackModel) Value(index int) interface{} {
	return fmt.Sprintf("%s - %s", m.items[index].Name, m.items[index].ArtistsName)
}

func NewTrackList(mw *MyMainWindow) *TrackModel {
	m := &TrackModel{}
	m.ItemsReset().Attach(func() {
		mw.lbTrackList.SetSuspended(true)
		defer mw.lbTrackList.SetSuspended(false)

		mw.lbTrackList.SendMessage(win.LB_RESETCONTENT, 0, 0)

		for i := 0; i < len(m.items); i++ {
			str := m.Value(i)
			lp := uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(str.(string))))
			mw.lbTrackList.SendMessage(win.LB_INSERTSTRING, uintptr(i), lp)
		}
	})
	return m
}
