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

	"github.com/siyuan-note/logging"
	"github.com/siyuan-note/riff/fsrs"
)

type FSRSStore struct {
	*BaseStore

	cards  []*fsrs.Card
	params fsrs.Parameters
}

func NewFSRSStore(saveDir string) *FSRSStore {
	return &FSRSStore{BaseStore: NewBaseStore("fsrs", saveDir), params: fsrs.DefaultParam()}
}

func (store *FSRSStore) Review(cardId int64, rating Rating) {
	store.lock.Lock()
	defer store.lock.Unlock()

	now := time.Now()
	for i, c := range store.cards {
		if c.Id == cardId {
			schedulingCards := store.params.Repeat(c, now)
			switch rating {
			case Again:
				c = &schedulingCards.Again
			case Hard:
				c = &schedulingCards.Hard
			case Good:
				c = &schedulingCards.Good
			case Easy:
				c = &schedulingCards.Easy
			}
			store.cards[i] = c
			return
		}
	}
	logging.LogWarnf("not found card [id=%d] to review", cardId)
}

func (store *FSRSStore) Dues() (ret []int64) {
	store.lock.Lock()
	defer store.lock.Unlock()

	now := time.Now()
	for _, c := range store.cards {
		if now.After(c.Due) {
			ret = append(ret, c.Id)
		}
	}
	return
}
