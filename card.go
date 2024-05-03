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

// Card 描述了闪卡。
type Card interface {
	// ID 返回闪卡 ID。
	ID() string

	// CardSourceID 返回关联的cardsource ID
	CardSourceID() string

	// GetGroup 返回当前卡片的Group
	GetGroup() string

	// SetGroup 设置当前卡片的Group
	SetGroup(newGroup string)

	// GetTag 返回当前卡片的Tag
	GetTag() string

	// SetTag 设置当前卡片的Tag
	SetTag(newTag string)

	// GetSuspend 返回当前卡片的暂停状态
	GetSuspend() bool

	// SwtichSuspend 切换当前卡片的暂停状态
	SwtichSuspend()

	// GetContext 返回当前卡片的Context
	GetContext() map[string]string

	// SetContext 使用 value 设置当前卡片的 key
	SetContext(key string, value string)

	// NextDues 返回每种评分对应的下次到期时间。
	NextDues() map[Rating]time.Time

	// SetNextDues 设置每种评分对应的下次到期时间。
	SetNextDues(map[Rating]time.Time)

	// SetDue 设置到期时间。
	SetDue(time.Time)

	// GetLapses 返回闪卡的遗忘次数。
	GetLapses() int

	// GetReps 返回闪卡的复习次数。
	GetReps() int

	// GetState 返回闪卡状态。
	GetState() State

	// GetLastReview 返回闪卡的最后复习时间。
	GetLastReview() time.Time

	// Clone 返回闪卡的克隆。
	Clone() Card

	// Impl 返回具体的闪卡实现。
	Impl() interface{}

	// SetImpl 设置具体的闪卡实现。
	SetImpl(c interface{})
}

// BaseCard 描述了基础的闪卡实现。
type BaseCard struct {
	CID     string
	SID     string
	Group   string
	Tag     string
	Suspend bool
	Context map[string]string
	NDues   map[Rating]time.Time
}

func (card *BaseCard) NextDues() map[Rating]time.Time {
	return card.NDues
}

func (card *BaseCard) SetNextDues(dues map[Rating]time.Time) {
	card.NDues = dues
}

func (card *BaseCard) ID() string {
	return card.CID
}

func (card *BaseCard) CardSourceID() string {
	return card.SID
}

func (card *BaseCard) GetGroup() string {
	return card.Group
}

func (card *BaseCard) SetGroup(newGroup string) {
	card.Group = newGroup
}

func (card *BaseCard) GetTag() string {
	return card.Tag
}

func (card *BaseCard) SetTag(newTag string) {
	card.Tag = newTag
}

func (card *BaseCard) GetSuspend() bool {
	return card.Suspend
}

func (card *BaseCard) SwtichSuspend() {
	card.Suspend = !card.Suspend
}

func (card *BaseCard) GetContext() map[string]string {
	return card.Context
}

func (card *BaseCard) SetContext(key string, value string) {
	card.Context[key] = value
}
