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

import "github.com/siyuan-note/riff/store"

// Deck 描述了一套闪卡。
type Deck struct {
	Name    string // 唯一名称
	Desc    string // 描述
	Created int64  // 创建时间
	Updated int64  // 更新时间

	store *store.Store // 闪卡存储
}

func NewDeck(name string, store *store.Store) *Deck {
	return &Deck{Name: name, store: store}
}

// Card 描述了一张闪卡。
type Card struct {
	CardID  int64  // 卡片 ID，对应 fsrs.Card.Id
	BlockID string // 内容块 ID，对应 siyuan.Block.ID
}
