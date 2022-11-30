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

	"github.com/open-spaced-repetition/go-fsrs"
)

func TestFSRSStore(t *testing.T) {
	const storePath = "testdata"
	os.MkdirAll(storePath, 0755)
	defer os.RemoveAll(storePath)

	store := NewFSRSStore(storePath)
	p := fsrs.DefaultParam()
	start := time.Now()
	repeatTime := start
	ids := map[string]bool{}
	for i := 0; i < 10000; i++ {
		id, blockID := newID(), newID()
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
}
