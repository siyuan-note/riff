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

package store

// Card 描述了闪卡。
type Card interface {
	// ID 返回闪卡 ID。
	ID() string

	// BlockID 返回闪卡关联的内容块 ID。
	BlockID() string

	// Impl 返回具体的闪卡实现。
	Impl() interface{}

	// SetImpl 设置具体的闪卡实现。
	SetImpl(c interface{})
}

// BaseCard 描述了基础的闪卡实现。
type BaseCard struct {
	id      string
	blockID string
}

func (card *BaseCard) ID() string {
	return card.id
}

func (card *BaseCard) BlockID() string {
	return card.blockID
}
