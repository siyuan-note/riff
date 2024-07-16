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

	// BlockID 返回闪卡关联的内容块 ID。
	BlockID() string

	// CSID 获取卡片的 CSID
	CSID() string

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
	CID      string
	CSID     string
	Update   time.Time
	State    State
	Lapses   int
	Reps     int
	Suspend  bool
	Tag      string
	Flag     string
	Priority float64
	Due      time.Time
	NDues    map[Rating]time.Time
	algo     string
	AlgoImpl interface{}
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

func (card *BaseCard) Impl() interface{} {
	return card.AlgoImpl
}
func (card *BaseCard) SetImpl(c interface{}) {
	card.AlgoImpl = c
}
