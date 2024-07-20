package riff

type CardSource interface {
	GetCSID() string
	GetDIDs() []string
	GetBlockIDs() []string
}

type BaseCardSource struct {
	CSID          string `xorm:"pk index"`
	Hash          string
	BlockIDs      []string
	DID           []string
	CType         string
	SourceContext map[string]interface{}
}

func NewBaseCardSource(DID string) *BaseCardSource {
	ID := newID()
	cardSource := &BaseCardSource{
		CSID:          ID,
		DID:           []string{DID},
		SourceContext: map[string]interface{}{},
	}
	return cardSource
}

func (cs *BaseCardSource) GetCSID() string {
	return cs.CSID
}

func (cs *BaseCardSource) GetDIDs() []string {
	return cs.DID
}
func (cs *BaseCardSource) GetBlockIDs() []string {
	return cs.BlockIDs
}
