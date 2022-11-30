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
	"github.com/siyuan-note/riff/store"
)

// Deck 描述了一套闪卡包。
type Deck struct {
	Name    string                // 唯一名称
	Desc    string                // 描述
	Created int64                 // 创建时间
	Updated int64                 // 更新时间
	Cards   map[string]store.Card // 闪卡集合 <cardID, card>

	store store.Store // 底层存储
}

// NewDeck 创建一套命名为 name 的闪卡包，store 为底层数据存储和间隔复习算法的实现。
func NewDeck(name string, store store.Store) *Deck {
	return &Deck{Name: name, store: store}
}

// Review 复习一张闪卡，rating 为复习评分结果。
func (deck *Deck) Review(cardID string, rating store.Rating) {
	deck.store.Review(cardID, rating)
}

// Dues 返回所有到期的闪卡。
func (deck *Deck) Dues() (ret []store.Card) {
	return deck.store.Dues()
}
