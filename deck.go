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

import "time"

type Deck interface {
	AddCard(cardID, blockID string)                       // FINISH
	RemoveCard(cardID string)                             // TODO
	SetCard(card Card)                                    // TODO
	GetCard(cardID string) Card                           // TODO
	GetCardsByBlockID(blockID string) (ret []Card)        // UNKNOW
	GetCardsByBlockIDs(blockIDs []string) (ret []Card)    // INSTEAD OF BETTER
	GetNewCardsByBlockIDs(blockIDs []string) (ret []Card) // INSTEAD OF BETTER
	GetDueCardsByBlockIDs(blockIDs []string) (ret []Card) // INSTEAD OF BETTER
	GetBlockIDs() (ret []string)                          // TODO
	CountCards() int                                      // TODO
	Save() (err error)                                    // FINISH
	SaveLog(log *Log) (err error)                         // FINISH
	Review(cardID string, rating Rating) (ret *Log)       // FINISH
	Dues() (ret []Card)                                   // FINISH
}

// BaseDeck 描述了一套闪卡包。
type BaseDeck struct {
	DID          string // DeckID
	Name         string // 名称
	Desc         string // 描述
	DeckType     DeckType
	Created      time.Time // 创建时间
	Updated      time.Time `xorm:"updated"` // 更新时间
	ParentDeckID string
	DeckContext  map[string]interface{}
	riff         *Riff `json:"-"`
}

type DeckType string

const (
	Set        DeckType = "Set"
	Collection DeckType = "Collection"
)

func DefaultBaseDeck() *BaseDeck {
	Created := time.Now()
	deck := &BaseDeck{
		DID:         builtInDeck,
		Name:        "builtInDeck",
		Desc:        "built in Deck",
		Created:     Created,
		DeckType:    Collection,
		DeckContext: map[string]interface{}{},
	}
	return deck
}

func NewBaseDeck() (deck *BaseDeck) {
	deck = &BaseDeck{
		DID:      newID(),
		Created:  time.Now(),
		DeckType: Collection,
	}
	return
}

func (bd *BaseDeck) AddCard(cardID, blockID string) {
	// 空实现
}

func (bd *BaseDeck) RemoveCard(cardID string) {
	// 空实现
}

func (bd *BaseDeck) SetCard(card Card) {
	// 空实现
}

func (bd *BaseDeck) GetCard(cardID string) Card {
	// 空实现
	return nil
}

func (bd *BaseDeck) GetCardsByBlockID(blockID string) (ret []Card) {
	// 空实现
	return nil
}

func (bd *BaseDeck) GetCardsByBlockIDs(blockIDs []string) (ret []Card) {
	// 空实现
	return nil
}

func (bd *BaseDeck) GetNewCardsByBlockIDs(blockIDs []string) (ret []Card) {
	// 空实现
	return nil
}

func (bd *BaseDeck) GetDueCardsByBlockIDs(blockIDs []string) (ret []Card) {
	// 空实现
	return nil
}

func (bd *BaseDeck) GetBlockIDs() (ret []string) {
	// 空实现
	return nil
}

func (bd *BaseDeck) CountCards() int {
	// 空实现
	return 0
}

func (bd *BaseDeck) Save() (err error) {
	// 空实现
	return nil
}

func (bd *BaseDeck) SaveLog(log *Log) (err error) {
	// 空实现
	return nil
}

func (bd *BaseDeck) Review(cardID string, rating Rating) (ret *Log) {
	// 空实现
	return nil
}

func (bd *BaseDeck) Dues() (ret []Card) {
	// 空实现
	return nil
}
