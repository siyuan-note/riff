// Riff - Spaced repetition.
// Copyright (c) 2022-present, b3log.org
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package riff

import (
	"math/rand"
	"path/filepath"
	"sync"
	"time"
)

// Store 描述了闪卡存储。
type Store interface {

	// AddCard 添加一张卡片。
	AddCard(id, blockID string) Card

	// GetCard 获取一张卡片。
	GetCard(id string) Card

	// RemoveCard 移除一张卡片。
	RemoveCard(id string) Card

	// Review 闪卡复习。
	Review(id string, rating Rating)

	// Dues 获取所有到期的闪卡列表。
	Dues() []Card

	// Name 获取存储名称。
	Name() string

	// Algo 返回算法名称，如：fsrs。
	Algo() Algo

	// Load 从持久化存储中加载全部闪卡到内存。
	Load() (err error)

	// Save 将全部闪卡从内存保存到持久化存储中。
	Save() error

	// GetSaveDir 获取数据文件夹路径。
	GetSaveDir() string
}

// BaseStore 描述了基础的闪卡存储实现。
type BaseStore struct {
	name    string      // 存储名称，应该和卡包名称一致
	algo    Algo        // 算法名称，如：fsrs
	saveDir string      // 数据文件夹路径，如：F:\\SiYuan\\data\\storage\\riff\\
	lock    *sync.Mutex // 操作时需要用到的锁
}

func NewBaseStore(name string, algo Algo, saveDir string) *BaseStore {
	return &BaseStore{
		name:    name,
		algo:    algo,
		saveDir: saveDir,
		lock:    &sync.Mutex{},
	}
}

func (store *BaseStore) Name() string {
	return store.name
}

func (store *BaseStore) Algo() Algo {
	return store.algo
}

func (store *BaseStore) GetSaveDir() string {
	return store.saveDir
}

func (store *BaseStore) getMsgPackPath() string {
	return filepath.Join(store.saveDir, string(store.name)+"-"+string(store.algo)+"-cards.msgpack")
}

// Rating 描述了闪卡复习的评分。
type Rating int8

const (
	Again Rating = iota // 完全不会，必须再复习一遍
	Hard                // 有点难
	Good                // 一般
	Easy                // 很容易
)

// Algo 描述了闪卡复习算法的名称。
type Algo string

const (
	AlgoFSRS Algo = "fsrs"
	AlgoSM2  Algo = "sm2"
)

func newID() string {
	now := time.Now()
	return now.Format("20060102150405") + "-" + randStr(7)
}

func randStr(length int) string {
	letter := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}
