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
	"os"

	"github.com/siyuan-note/logging"
	"github.com/siyuan-note/riff/fsrs"
	"github.com/vmihailenco/msgpack/v5"
)

type Store struct {
	Path  string       // 数据库文件路径，如：F:\\SiYuan\\data\\storage\\riff.msgpack
	Cards []*fsrs.Card // 卡片列表
}

func NewStore(path string) *Store {
	return &Store{Path: path}
}

func (store *Store) Load() (err error) {
	data, err := os.ReadFile(store.Path)
	if nil != err {
		logging.LogErrorf("load store [%s] failed: %s", store.Path, err)
		return
	}

	if err = msgpack.Unmarshal(data, &store.Cards); nil != err {
		logging.LogErrorf("unmarshal store [%s] failed: %s", store.Path, err)
		return
	}
	return
}

func (store *Store) Save() (err error) {
	data, err := msgpack.Marshal(store.Cards)
	if nil != err {
		logging.LogErrorf("marshal store [%s] failed: %s", store.Path, err)
		return
	}
	if err = os.WriteFile(store.Path, data, 0644); nil != err {
		logging.LogErrorf("save store [%s] failed: %s", store.Path, err)
		return
	}
	return
}
