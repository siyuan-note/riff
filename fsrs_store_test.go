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
	"strings"
	"testing"
	"time"

	"github.com/88250/gulu"

	"github.com/open-spaced-repetition/go-fsrs"
)

const (
	requestRetention = 0.9
	maximumInterval  = 36500
	weights          = "0.40, 0.60, 2.40, 5.80, 4.93, 0.94, 0.86, 0.01, 1.49, 0.14, 0.94, 2.18, 0.05, 0.34, 1.26, 0.29, 2.61"
)

type BaseStruct struct {
	CID string
}

func TestFSRSStore(t *testing.T) {

	newCard := BaseStruct{
		CID: "1111",
	}
	t.Log(newCard.CID)

	const storePath = "testdata"
	os.MkdirAll(storePath, 0755)
	defer os.RemoveAll(storePath)

	store := NewFSRSStore("test-store", storePath, requestRetention, maximumInterval, weights)
	// 判断是否实现全部必须接口
	// var _ CardSourceStore = store
	var _ Store = store
	p := fsrs.DefaultParam()
	start := time.Now()
	repeatTime := start
	ids := map[string]bool{}
	var firstCardID, secondCardID, secondCardSourceID, firstBlockID, lastCardID, lastBlockID string
	max := 10
	for i := 0; i < max; i++ {
		id, blockID := newID(), newID()
		if 0 == i {
			firstCardID = id
			firstBlockID = blockID
		} else if 1 == i {
			secondCardID = id
		} else if max-1 == i {
			lastCardID = id
			lastBlockID = blockID
		}
		store.AddCard(id, blockID)
		card := store.GetCard(id)
		c := *card.Impl().(*fsrs.Card)
		ids[id] = true

		for j := 0; j < 10; j++ {
			schedulingInfo := p.Repeat(c, repeatTime)
			c = schedulingInfo[fsrs.Hard].Card
			repeatTime = c.Due
		}
		repeatTime = start
	}
	secondCard := store.cards[secondCardID]
	secondCardSourceID = secondCard.CardSourceID()
	secondCardSource, ok := store.cardSources[secondCardSourceID]
	if !ok {
		t.Fatal("CardSource no add successful")
	}

	if secondCardSource.CIDMap == nil {
		t.Fatal("CardSource CIDMap field init fail")
	}
	if !gulu.Str.Contains(secondCardID, secondCardSource.GetCardIDs()) {
		t.Fatal("add card no successful add to cardsource")
	}
	cardsLen := len(store.cards)
	t.Logf("cards len [%d]", cardsLen)
	if len(ids) != len(store.cards) {
		t.Fatalf("cards len [%d] != ids len [%d]", len(store.cards), len(ids))
	}

	count := store.CountCards()
	if cardsLen != count {
		t.Fatalf("cards len [%d] != count [%d]", cardsLen, count)
	}

	if err := store.Save(); nil != err {
		t.Fatal(err)
	}
	t.Logf("saved cards [len=%d]", len(store.cards))

	store = NewFSRSStore("test-store", storePath, requestRetention, maximumInterval, weights)
	if err := store.Load(); nil != err {
		t.Fatal(err)
	}
	t.Logf("loaded cards len [%d]", len(store.cards))

	if cardsLen != len(store.cards) {
		t.Fatal("cards len not equal")
	}

	secondCardSourceID = store.cards[secondCardID].CardSourceID()

	store.RemoveCard(secondCardID)
	if cardsLen-1 != len(store.cards) {
		t.Fatalf("remove cards len [%d] != [%d]", len(store.cards), cardsLen-1)
	}
	if cardsLen-1 != len(store.cardSources) {
		t.Fatalf("remove cardSources len [%d] != [%d]", len(store.cardSources), cardsLen-1)
	}
	if _, ok := store.cardSources[secondCardSourceID]; ok {
		t.Fatal("remove card related cardSources fail")
	}
	cards := store.GetCardsByBlockID(firstBlockID)
	if 1 != len(cards) {
		t.Fatalf("cards by block id [len=%d]", len(cards))
	}
	if firstCardID != cards[0].ID() {
		t.Fatalf("cards by block id [cardID=%s]", cards[0].ID())
	}

	cards = store.GetCardsByBlockID(lastBlockID)
	if 1 != len(cards) {
		t.Fatalf("cards by block id [len=%d]", len(cards))
	}
	if lastCardID != cards[0].ID() {
		t.Fatalf("cards by block id [cardID=%s]", cards[0].ID())
	}

	cards = store.GetCardsByBlockIDs([]string{firstBlockID, lastBlockID})
	if 2 != len(cards) {
		t.Fatalf("cards by block ids [len=%d]", len(cards))
	}
	cardIDs := []string{cards[0].ID(), cards[1].ID()}
	if !gulu.Str.Contains(firstCardID, cardIDs) {
		t.Fatalf("cards by block ids [cardIDs=%v]", cardIDs)
	}
	if !gulu.Str.Contains(lastCardID, cardIDs) {
		t.Fatalf("cards by block ids [cardIDs=%v]", cardIDs)
	}
	t.Logf("cards by block ids [len=%d], card ids [%s]", len(cards), strings.Join(cardIDs, ", "))
}
