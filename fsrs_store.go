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

func NewFSRSStore(name, saveDir string) *FSRSStore {
	return &FSRSStore{
		BaseStore: NewBaseStore(name, "fsrs", saveDir),
		cards:     map[string]*FSRSCard{},
		params:    fsrs.DefaultParam(),
	}
}

func (store *FSRSStore) AddCard(id, blockID string) Card {
	store.lock.Lock()
	defer store.lock.Unlock()

	c := fsrs.NewCard()
	card := &FSRSCard{BaseCard: &BaseCard{id, blockID}, C: &c}
	store.cards[id] = card
	return card
}

func (store *FSRSStore) GetCard(id string) Card {
	store.lock.Lock()
	defer store.lock.Unlock()

	return store.cards[id]
}

func (store *FSRSStore) RemoveCard(id string) Card {
	store.lock.Lock()
	defer store.lock.Unlock()
	card := store.cards[id]
	if nil == card {
		return nil
	}
	delete(store.cards, id)
	return card
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

	schedulingInfo := store.params.Repeat(*card.C, now)
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
	if !gulu.File.IsExist(p) {
		return
	}

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

type FSRSCard struct {
	*BaseCard
	C *fsrs.Card
}

func (card *FSRSCard) Impl() interface{} {
	return card.C
}

func (card *FSRSCard) SetImpl(c interface{}) {
	card.C = c.(*fsrs.Card)
}
