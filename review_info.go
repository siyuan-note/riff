package riff

type ReviewInfo struct {
	BaseCard       `xorm:"extends"`
	BaseCardSource `xorm:"extends"`
}

func (ri *ReviewInfo) ToCard() (card Card) {
	card = &ri.BaseCard
	return
}
