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
	"encoding/json"
	"time"

	"github.com/open-spaced-repetition/go-fsrs"
	"github.com/siyuan-note/logging"
	_ "modernc.org/sqlite"
)

// Card 描述了闪卡。
type Card interface {
	// ID 返回闪卡 ID。
	ID() string

	// BlockID 返回闪卡关联的内容块 ID。
	// BlockID() string

	// CSID 获取卡片的 CSID
	GetCSID() string

	// NextDues 返回每种评分对应的下次到期时间。
	NextDues() map[Rating]time.Time

	// SetNextDues 设置每种评分对应的下次到期时间。
	SetNextDues(map[Rating]time.Time)

	// GetUpdate 返回闪卡的更新时间。
	GetUpdate() time.Time

	// SetUpdate 设置闪卡的更新时间。
	SetUpdate(time.Time)

	// GetState 返回闪卡状态。
	GetState() State

	// SetState 设置闪卡状态。
	SetState(State)

	// GetLapses 返回闪卡的遗忘次数。
	GetLapses() int

	// SetLapses 设置闪卡的遗忘次数。
	SetLapses(int)

	// GetReps 返回闪卡的复习次数。
	GetReps() int

	// SetReps 设置闪卡的复习次数。
	SetReps(int)

	// GetSuspend 返回闪卡是否被暂停。
	GetSuspend() bool

	// SetSuspend 设置闪卡是否被暂停。
	SetSuspend(bool)

	// GetTag 返回闪卡的标签。
	GetTag() string

	// SetTag 设置闪卡的标签。
	SetTag(string)

	// GetFlag 返回闪卡的标志。
	GetFlag() string

	// SetFlag 设置闪卡的标志。
	SetFlag(string)

	// GetPriority 返回闪卡的优先级。
	GetPriority() float64

	// SetPriority 设置闪卡的优先级。
	SetPriority(float64)

	// GetDue 返回闪卡的到期时间。
	GetDue() time.Time

	// SetDue 设置闪卡的到期时间。
	SetDue(time.Time)

	// 返回 Algo
	GetAlgo() Algo

	UseAlgo(algo Algo)

	// 返回 MarshalImpl
	GetMarshalImpl() []uint8

	// 对 Impl 进行 Marshal
	MarshalImpl()

	UnmarshalImpl()

	// Impl 返回具体的闪卡实现。
	Impl() interface{}

	// SetImpl 设置具体的闪卡实现。
	SetImpl(c interface{})
}

func UnmarshalImpl(card Card) {
	switch card.GetAlgo() {
	case AlgoFSRS:
		impl := fsrs.Card{}
		json.Unmarshal(card.GetMarshalImpl(), &impl)
		card.SetImpl(impl)
	default:
		return
	}
}

// BaseCard 描述了基础的闪卡实现。
type BaseCard struct {
	CID          string `xorm:"pk index"`
	CSID         string
	Update       time.Time
	State        State //State 返回闪卡状态。
	Lapses       int   //Lapses 返回闪卡的遗忘次数。
	Reps         int   //Reps 返回闪卡的复习次数。
	Suspend      bool  `xorm:"index"`
	Tag          string
	Flag         string
	Priority     float64
	Due          time.Time            `xorm:"index"`
	NDues        map[Rating]time.Time `xorm:"-"`
	Algo         Algo
	AlgoImpl     interface{} `xorm:"-"`
	AlgoImplData []uint8     `json:"-"`
}

func NewBaseCard(cs CardSource) (card *BaseCard) {
	CSID := cs.GetCSID()
	card = &BaseCard{
		CSID:         CSID,
		CID:          newID(),
		State:        New,
		Update:       time.Now(),
		Due:          time.Now(),
		NDues:        map[Rating]time.Time{},
		Priority:     0.5,
		AlgoImplData: []uint8{},
	}
	return
}

func (card *BaseCard) ID() string {
	return card.CID
}

func (card *BaseCard) GetCSID() string {
	return card.CSID
}

func (c *BaseCard) GetUpdate() time.Time {
	return c.Update
}

func (c *BaseCard) SetUpdate(update time.Time) {
	c.Update = update
}

func (c *BaseCard) GetState() State {
	return c.State
}

func (c *BaseCard) SetState(state State) {
	c.State = state
}

func (c *BaseCard) GetLapses() int {
	return c.Lapses
}

func (c *BaseCard) SetLapses(lapses int) {
	c.Lapses = lapses
}

func (c *BaseCard) GetReps() int {
	return c.Reps
}

func (c *BaseCard) SetReps(reps int) {
	c.Reps = reps
}

func (c *BaseCard) GetSuspend() bool {
	return c.Suspend
}

func (c *BaseCard) SetSuspend(suspend bool) {
	c.Suspend = suspend
}

func (c *BaseCard) GetTag() string {
	return c.Tag
}

func (c *BaseCard) SetTag(tag string) {
	c.Tag = tag
}

func (c *BaseCard) GetFlag() string {
	return c.Flag
}

func (c *BaseCard) SetFlag(flag string) {
	c.Flag = flag
}

func (c *BaseCard) GetPriority() float64 {
	return c.Priority
}

func (c *BaseCard) SetPriority(priority float64) {
	c.Priority = priority
}

func (c *BaseCard) GetDue() time.Time {
	return c.Due
}

func (c *BaseCard) SetDue(due time.Time) {
	c.Due = due
}

func (card *BaseCard) NextDues() map[Rating]time.Time {
	return card.NDues
}

func (card *BaseCard) SetNextDues(dues map[Rating]time.Time) {
	card.NDues = dues
}

func (card *BaseCard) Impl() interface{} {
	return card.AlgoImpl
}
func (card *BaseCard) SetImpl(c interface{}) {
	card.AlgoImpl = c
}

func (card *BaseCard) UseAlgo(algo Algo) {
	switch algo {
	case AlgoFSRS:
		AlgoImpl := fsrs.NewCard()
		card.AlgoImpl = AlgoImpl
		card.Algo = AlgoFSRS
		// card.Due = AlgoImpl.Due
	default:
		logging.LogErrorf("unsupported Algo: %s", algo)
	}
}

func (card *BaseCard) GetMarshalImpl() []uint8 {
	if card.AlgoImplData == nil || len(card.AlgoImplData) == 0 {
		card.MarshalImpl()
	}

	return card.AlgoImplData
}

func (card *BaseCard) MarshalImpl() {
	data, _ := json.Marshal(card.AlgoImpl)
	card.AlgoImplData = data
}

func (card *BaseCard) UnmarshalImpl() {
	UnmarshalImpl(card)
}

func (card *BaseCard) GetAlgo() Algo {
	return card.Algo
}
