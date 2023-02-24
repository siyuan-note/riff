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
	"github.com/88250/gulu"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs"
)

func TestFSRSStore(t *testing.T) {
	const storePath = "testdata"
	os.MkdirAll(storePath, 0755)
	defer os.RemoveAll(storePath)

	store := NewFSRSStore("test-store", storePath)
	p := fsrs.DefaultParam()
	start := time.Now()
	repeatTime := start
	ids := map[string]bool{}
	var firstCardID, firstBlockID, lastCardID, lastBlockID string
	max := 10000
	for i := 0; i < max; i++ {
		id, blockID := newID(), newID()
		if 0 == i {
			firstCardID = id
			firstBlockID = blockID
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

	if err := store.Load(); nil != err {
		t.Fatal(err)
	}
	t.Logf("loaded cards len [%d]", len(store.cards))

	if cardsLen != len(store.cards) {
		t.Fatal("cards len not equal")
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
