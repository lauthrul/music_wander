package ui

import (
	"fmt"
	"github.com/lxn/walk"
	"github.com/lxn/win"
	"syscall"
	"unsafe"
	"wander/model"
)

type MusicListModel struct {
	walk.ListModelBase
	items []*model.Music
}

func (m *MusicListModel) ItemCount() int {
	return len(m.items)
}

func (m *MusicListModel) Value(index int) interface{} {
	return fmt.Sprintf("[%03d] %s - %s", index+1, m.items[index].Info.Name, m.items[index].Info.ArtistsName)
}

func NewTrackList(mw *MyMainWindow) *MusicListModel {
	m := &MusicListModel{}
	m.ItemsReset().Attach(func() {
		mw.lbMusicList.SetSuspended(true)
		defer mw.lbMusicList.SetSuspended(false)

		mw.lbMusicList.SendMessage(win.LB_RESETCONTENT, 0, 0)

		for i := 0; i < len(m.items); i++ {
			str := m.Value(i)
			lp := uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(str.(string))))
			mw.lbMusicList.SendMessage(win.LB_INSERTSTRING, uintptr(i), lp)
		}
	})
	return m
}
