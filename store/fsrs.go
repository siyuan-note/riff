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
	"time"

	"github.com/siyuan-note/riff/fsrs"
)

type FSRSStore struct {
	*BaseStore

	cards []*fsrs.Card
}

func NewFSRSStore(path string) *FSRSStore {
	return &FSRSStore{BaseStore: &BaseStore{algo: "fsrs", saveDir: path}}
}

func (store *FSRSStore) Dues() (ret []int64) {
	now := time.Now()
	for _, c := range store.cards {
		if now.After(c.Due) {
			ret = append(ret, c.Id)
		}
	}
	return
}
