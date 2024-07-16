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

type Deck interface {
	New(id string) Deck
	AddCard(cardID, blockID string)
	RemoveCard(cardID string)
	SetCard(card Card)
	GetCard(cardID string) Card
	GetCardsByBlockID(blockID string) (ret []Card)
	GetCardsByBlockIDs(blockIDs []string) (ret []Card)
	GetNewCardsByBlockIDs(blockIDs []string) (ret []Card)
	GetDueCardsByBlockIDs(blockIDs []string) (ret []Card)
	GetBlockIDs() (ret []string)
	CountCards() int
	Save() (err error)
	SaveLog(log *Log) (err error)
	Review(cardID string, rating Rating) (ret *Log)
	Dues() (ret []Card)
}

// BaseDeck 描述了一套闪卡包。
type BaseDeck struct {
	DID          string // ID
	Name         string // 名称
	Desc         string // 描述
	Created      int64  // 创建时间
	Updated      int64  // 更新时间
	ParentDeckID string
	DeckContext  map[string]interface{}
}
