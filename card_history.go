package riff

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/open-spaced-repetition/go-fsrs"
)

type History interface {
	ID() string

	// 对 Impl 进行 Marshal
	MarshalImpl()

	UnmarshalImpl() (err error)
}

type BaseHistory struct {
	HID          string `xorm:"pk index"`
	CID          string
	Update       time.Time `xorm:"index"`
	UpdateResult int
	State        State
	Tag          string
	Flag         string
	Suspend      bool
	Priority     float64
	Due          time.Time
	Algo         Algo
	AlgoImpl     interface{} `xorm:"-"`
	AlgoImplData []uint8     `json:"-"`
}

type ReviewLog struct {
	HID    string
	Rate   Rating
	Review time.Time
}

type ReviewHistory struct {
	BaseHistory `xorm:"extends"`
	ReviewLog   `xorm:"extends"`
}

func NewBaseHistory(c Card) (history *BaseHistory) {
	return &BaseHistory{
		HID:          newID(),
		CID:          c.ID(),
		Update:       time.Now(),
		UpdateResult: CreateHistory, // 这里可以根据需要设置具体的更新结果
		State:        c.GetState(),
		Tag:          c.GetTag(),
		Flag:         c.GetFlag(),
		Suspend:      c.GetSuspend(),
		Priority:     c.GetPriority(),
		Due:          c.GetDue(),
		Algo:         c.GetAlgo(),
		AlgoImpl:     c.Impl(),           // 假设Card接口有Impl方法
		AlgoImplData: c.GetMarshalImpl(), // 假设Card接口有GetMarshalImpl方法
	}
}

func (b *BaseHistory) ID() string {
	return b.HID
}

func (b *BaseHistory) UnmarshalImpl() (err error) {
	if b.AlgoImplData == nil || len(b.AlgoImplData) == 0 {
		err = errors.New("dont have AlgoImplData")
		return
	}
	switch b.Algo {
	case AlgoFSRS:
		impl := fsrs.Card{}
		json.Unmarshal(b.AlgoImplData, &impl)
		b.AlgoImpl = impl
	default:
		err = fmt.Errorf("un support Algo Type : %s", string(b.Algo))
		return
	}
	return
}

func (b *BaseHistory) MarshalImpl() {
	if len(b.AlgoImplData) != 0 && b.AlgoImpl == nil {
		return
	}
	data, _ := json.Marshal(b.AlgoImpl)
	b.AlgoImplData = data
}

func NewReviewLog(h History, rate Rating) (log *ReviewLog) {
	log = &ReviewLog{
		HID:    h.ID(),
		Rate:   rate,
		Review: time.Now(),
	}
	return
}

const (
	EditUpdate    int = -2
	CreateHistory int = -1
	ReviewUpdate  int = 1
)
