package model

const (
	// API地址: https://cloud.tencent.com/developer/article/1543945

	// 歌单
	// id=19723756，云音乐飙升榜
	// id=3779629，云音乐新歌榜
	// id=3778678，云音乐热歌榜
	// id=2250011882，抖音排行榜
	Playlist = "https://music.163.com/api/playlist/detail?id=%s"

	// 评论
	Comment = "http://music.163.com/api/v1/resource/comments/R_SO_4_{歌曲ID}?limit=20&offset=0"

	// 歌词
	Lyrics = "https://music.163.com/api/song/lyric?id={歌曲ID}&lv=1&kv=1&tv=-1"

	// 随机歌曲
	RandomUrl  = "https://api.66mz8.com/api/rand.music.163.php?format=json"
	RandomUrl2 = "https://api.66mz8.com/api/music.163.php?format=json"

	// 歌曲搜索
	SearchUrl = "https://v1.alapi.cn/api/music/search?keyword=我爱你"

	// 歌曲真实地址
	LinkUrl = "https://v1.alapi.cn/api/music/url?format=json&id=%s"
)
