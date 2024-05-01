package riff

import "sync"

type CardSourceStore interface {
}

type BaseCardSourceStore struct {
	id      string // 存储 ID，应该和卡包 ID 一致
	saveDir string // 数据文件夹路径，如：F:\\SiYuan\\data\\storage\\riff\\
	lock    *sync.Mutex
}

func NewBaseCardSourceStore(id string, saveDir string) *BaseCardSourceStore {
	return &BaseCardSourceStore{
		id:      id,
		saveDir: saveDir,
		lock:    &sync.Mutex{},
	}
}
