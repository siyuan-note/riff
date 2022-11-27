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
	"testing"
	"time"

	"github.com/88250/gulu"
	"github.com/siyuan-note/riff/fsrs"
	"github.com/vmihailenco/msgpack/v5"
)

func TestStoreLoadSave(t *testing.T) {
	const storePath = "testdata/"
	store := NewStore(storePath)

	defer os.Remove(store.getMsgPackPath())
	p := fsrs.DefaultParam()
	start := time.Now()
	repeatTime := start
	for i := 0; i < 10000; i++ {
		card := fsrs.NewCard()
		store.Cards = append(store.Cards, &card)

		for j := 0; j < 10; j++ {
			schedulingCards := p.Repeat(&card, repeatTime)
			card = schedulingCards.Hard
			repeatTime = card.Due
		}
		repeatTime = start
	}
	cardsLen := len(store.Cards)
	t.Logf("cards len [%d]", cardsLen)

	if err := store.Save(); nil != err {
		t.Fatal(err)
	}

	store.Cards = nil
	if err := store.Load(); nil != err {
		t.Fatal(err)
	}
	if cardsLen != len(store.Cards) {
		t.Fatal("cards len not equal")
	}
}

func TestJSONMsgPack(t *testing.T) {
	const cardsJSON = "testdata/cards.json"
	defer os.Remove(cardsJSON)
	const cardsMsgPack = "testdata/cards.msgpack"
	defer os.Remove(cardsMsgPack)

	p := fsrs.DefaultParam()
	start := time.Now()
	repeatTime := start
	var cards []*fsrs.Card
	for i := 0; i < 100000; i++ {
		card := fsrs.NewCard()
		cards = append(cards, &card)

		for j := 0; j < 10; j++ {
			schedulingCards := p.Repeat(&card, repeatTime)
			card = schedulingCards.Hard
			repeatTime = card.Due
		}
		repeatTime = start
	}
	cardsLen := len(cards)
	t.Logf("cards len [%d]", cardsLen)

	now := time.Now()
	data, err := gulu.JSON.MarshalJSON(cards)
	if nil != err {
		t.Fatal(err)
	}
	if err = os.WriteFile(cardsJSON, data, 0644); nil != err {
		t.Fatal(err)
	}
	t.Logf("save by [json] elapsed [%dms], size [%.2fMB]", time.Since(now).Milliseconds(), float64(len(data))/1024/1024)

	now = time.Now()
	data, err = msgpack.Marshal(cards)
	if nil != err {
		t.Fatal(err)
	}
	if err = os.WriteFile(cardsMsgPack, data, 0644); nil != err {
		t.Fatal(err)
	}
	t.Logf("save by [masgpack] elapsed [%dms], size [%.2fMB]", time.Since(now).Milliseconds(), float64(len(data))/1024/1024)

	now = time.Now()
	data, err = os.ReadFile(cardsJSON)
	if nil != err {
		t.Fatal(err)
	}
	if err = gulu.JSON.UnmarshalJSON(data, &cards); nil != err {
		t.Fatal(err)
	}
	if cardsLen != len(cards) {
		t.Fatal("cards len not equal")
	}
	t.Logf("load by [json] elapsed [%dms], size [%.2fMB]", time.Since(now).Milliseconds(), float64(len(data))/1024/1024)

	now = time.Now()
	data, err = os.ReadFile(cardsMsgPack)
	if nil != err {
		t.Fatal(err)
	}
	if err = msgpack.Unmarshal(data, &cards); nil != err {
		t.Fatal(err)
	}
	if cardsLen != len(cards) {
		t.Fatal("cards len not equal")
	}
	t.Logf("load by [masgpack] elapsed [%dms], size [%.2fMB]", time.Since(now).Milliseconds(), float64(len(data))/1024/1024)
}
