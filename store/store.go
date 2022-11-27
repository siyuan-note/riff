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

package store

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/88250/gulu"
	"github.com/siyuan-note/logging"
)

// Store 描述了闪卡存储。
type Store interface {

	// Algo 返回算法名称，如：fsrs。
	Algo() string

	// Load 加载全部闪卡数据。
	Load() (data []byte, err error)

	// Save 保存全部闪卡数据。
	Save(data []byte) error

	// Review 闪卡复习。
	Review(id int64, rating Rating) error

	// Dues 获取所有到期的闪卡 IDs。
	Dues() []int64
}

// Rating 描述了闪卡复习的评分。
type Rating int8

const (
	Again Rating = iota // 完全不会，必须再复习一遍
	Hard                // 有点难
	Good                // 一般
	Easy                // 很容易
)

// BaseStore 描述了基础的闪卡存储。
type BaseStore struct {
	algo    string // 算法名称，如：fsrs
	saveDir string // 数据文件夹路径，如：F:\\SiYuan\\data\\storage\\riff\\

	lock *sync.Mutex
}

func NewBaseStore(algo, saveDir string) *BaseStore {
	return &BaseStore{
		algo:    algo,
		saveDir: saveDir,
		lock:    &sync.Mutex{},
	}
}

func (store *BaseStore) Algo() string {
	return store.algo
}

func (store *BaseStore) getMsgPackPath() string {
	return filepath.Join(store.saveDir, store.algo+".msgpack")
}

func (store *BaseStore) Load() (data []byte, err error) {
	msgpackPath := store.getMsgPackPath()
	data, err = os.ReadFile(msgpackPath)
	if nil != err {
		logging.LogErrorf("load cards [%s] failed: %s", msgpackPath, err)
		return
	}
	return
}

func (store *BaseStore) Save(data []byte) (err error) {
	msgpackPath := store.getMsgPackPath()
	if err = gulu.File.WriteFileSafer(msgpackPath, data, 0644); nil != err {
		logging.LogErrorf("save cards [%s] failed: %s", msgpackPath, err)
		return
	}
	return
}
