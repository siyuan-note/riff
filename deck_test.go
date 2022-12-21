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

	"github.com/open-spaced-repetition/go-fsrs"
)

func TestDeck(t *testing.T) {
	const saveDir = "testdata"
	os.MkdirAll(saveDir, 0755)
	defer os.RemoveAll(saveDir)
	deckName := "deck0"
	deck, err := LoadDeck(saveDir, deckName, AlgoFSRS)
	if nil != err {
		t.Fatal(err)
	}

	cardID, blockID := newID(), newID()
	deck.AddCard(cardID, blockID)
	card := deck.GetCard(blockID)
	if card.ID() != cardID {
		t.Fatalf("card id [%s] != [%s]", card.ID(), cardID)
	}

	deck.Review(blockID, Good)
	due := card.Impl().(*fsrs.Card).Due.UnixMilli()
	card = deck.GetCard(blockID)
	due2 := card.Impl().(*fsrs.Card).Due.UnixMilli()
	if due2 != due {
		t.Fatalf("card due [%v] != [%v]", due2, due)
	}

	err = deck.Save()
	if nil != err {
		t.Fatal(err)
	}
	deck = nil

	deck, err = LoadDeck(saveDir, deckName, AlgoFSRS)
	if nil != err {
		t.Fatal(err)
	}

	if deckName != deck.Name {
		t.Fatalf("deck name [%s] != [%s]", deck.Name, deckName)
	}

	card = deck.GetCard(blockID)
	if card.ID() != cardID {
		t.Fatalf("card id [%s] != [%s]", card.ID(), cardID)
	}
	due3 := card.Impl().(*fsrs.Card).Due.UnixMilli()
	if due2 != due3 {
		t.Fatalf("card due [%v] != [%v]", due2, due3)
	}
}
