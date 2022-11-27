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
	"testing"
	"time"

	"github.com/siyuan-note/logging"
	"github.com/siyuan-note/riff/fsrs"
	"github.com/vmihailenco/msgpack/v5"
)

func TestStoreLoadSave(t *testing.T) {
	const storePath = "../testdata/"
	store := NewFSRSStore(storePath)
	defer os.Remove(store.getMsgPackPath())

	p := fsrs.DefaultParam()
	start := time.Now()
	repeatTime := start
	for i := 0; i < 10000; i++ {
		card := fsrs.NewCard()
		store.cards = append(store.cards, &card)

		for j := 0; j < 10; j++ {
			schedulingCards := p.Repeat(&card, repeatTime)
			card = schedulingCards.Hard
			repeatTime = card.Due
		}
		repeatTime = start
	}
	cardsLen := len(store.cards)
	t.Logf("cards len [%d]", cardsLen)

	data, err := msgpack.Marshal(store.cards)
	if nil != err {
		logging.LogErrorf("marshal cards failed: %s", err)
		return
	}

	if err = store.Save(data); nil != err {
		t.Fatal(err)
	}
	t.Logf("data size [%.2fMB]", float64(len(data))/1024/1024)

	store.cards = nil
	if data, err = store.Load(); nil != err {
		t.Fatal(err)
	}

	if err = msgpack.Unmarshal(data, &store.cards); nil != err {
		t.Fatal(err)
	}
	t.Logf("loaded cards len [%d]", len(store.cards))

	if cardsLen != len(store.cards) {
		t.Fatal("cards len not equal")
	}
}
