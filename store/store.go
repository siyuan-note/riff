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

	// Dues 获取所有到期的闪卡 IDs。
	Dues() []int64
}

// BaseStore 描述了基础的闪卡存储。
type BaseStore struct {
	algo    string // 算法名称，如：fsrs
	saveDir string // 数据文件夹路径，如：F:\\SiYuan\\data\\storage\\riff\\
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
	if err = os.WriteFile(msgpackPath, data, 0644); nil != err {
		logging.LogErrorf("save cards [%s] failed: %s", msgpackPath, err)
		return
	}
	return
}
