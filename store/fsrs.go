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
	"time"

	"github.com/88250/gulu"
	"github.com/open-spaced-repetition/go-fsrs"
	"github.com/siyuan-note/logging"
	"github.com/vmihailenco/msgpack/v5"
)

type FSRSStore struct {
	*BaseStore

	cards  map[string]*FSRSCard
	params fsrs.Parameters
}

func NewFSRSStore(saveDir string) *FSRSStore {
	return &FSRSStore{
		BaseStore: NewBaseStore("fsrs", saveDir),
		cards:     map[string]*FSRSCard{},
		params:    fsrs.DefaultParam(),
	}
}

type FSRSCard struct {
	*BaseCard
	c *fsrs.Card
}

func (card *FSRSCard) Impl() interface{} {
	return card.c
}

func (card *FSRSCard) SetImpl(c interface{}) {
	card.c = c.(*fsrs.Card)
}

func (store *FSRSStore) AddCard(card Card) {
	store.lock.Lock()
	defer store.lock.Unlock()
	store.cards[card.ID()] = card.(*FSRSCard)
}

func (store *FSRSStore) GetCard(id string) Card {
	store.lock.Lock()
	defer store.lock.Unlock()
	return store.cards[id]
}

func (store *FSRSStore) RemoveCard(id string) {
	store.lock.Lock()
	defer store.lock.Unlock()
	delete(store.cards, id)
}

func (store *FSRSStore) Review(cardId string, rating Rating) {
	store.lock.Lock()
	defer store.lock.Unlock()

	now := time.Now()
	card := store.cards[cardId]
	if nil == card {
		logging.LogWarnf("not found card [id=%s] to review", cardId)
		return
	}

	schedulingInfo := store.params.Repeat(*card.c, now)
	updated := schedulingInfo[fsrs.Rating(rating)].Card
	card.SetImpl(&updated)
	store.cards[cardId] = card
	return
}

func (store *FSRSStore) Dues() (ret []Card) {
	store.lock.Lock()
	defer store.lock.Unlock()

	now := time.Now()
	for _, card := range store.cards {
		c := card.Impl().(*fsrs.Card)
		if now.After(c.Due) {
			ret = append(ret, card)
		}
	}
	return
}

func (store *FSRSStore) Load() (err error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	store.cards = map[string]*FSRSCard{}
	p := store.getMsgPackPath()
	data, err := os.ReadFile(p)
	if nil != err {
		logging.LogErrorf("load cards failed: %s", err)
	}
	if err = msgpack.Unmarshal(data, &store.cards); nil != err {
		logging.LogErrorf("load cards failed: %s", err)
		return
	}
	return
}

func (store *FSRSStore) Save() (err error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	p := store.getMsgPackPath()
	data, err := msgpack.Marshal(store.cards)
	if nil != err {
		logging.LogErrorf("save cards failed: %s", err)
		return
	}
	if err = gulu.File.WriteFileSafer(p, data, 0644); nil != err {
		logging.LogErrorf("save cards failed: %s", err)
		return
	}
	return
}
