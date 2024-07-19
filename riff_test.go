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
)

func TestDeck(t *testing.T) {
	const saveDir = "testdata"
	const RequestRetention = 0.95
	const cardSourceNum = 100
	const preSourceCardNum = 10
	os.MkdirAll(saveDir, 0755)
	defer os.RemoveAll(saveDir)
	riff := NewBaseRiff()
	deck := DefaultBaseDeck()
	csList := []CardSource{}
	cardList := []Card{}
	for i := 0; i < cardSourceNum; i++ {
		cs := NewBaseCardSource(deck.DID)
		csList = append(csList, cs)
		for i := 0; i < preSourceCardNum; i++ {
			cardList = append(cardList, NewBaseCard(cs))
		}
	}

	riff.AddDeck(deck)
	riff.AddCardSource(csList)
	riff.AddCard(cardList)
	reviewCard := riff.Due()
	for _, card := range reviewCard {
		riff.Review(card, Easy, RequestRetention)
	}
	riff.Save(saveDir)
	riff.Load(saveDir)

}
