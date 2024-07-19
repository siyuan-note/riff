package riff

type CardSource interface {
	GetCSID() string
	GetDID() string
}

type BaseCardSource struct {
	CSID          string
	Hash          string
	BlockIDs      []string
	DID           string
	CType         string
	SourceContext map[string]interface{}
}

func NewBaseCardSource(DID string) *BaseCardSource {
	ID := newID()
	cardSource := &BaseCardSource{
		CSID: ID,
		DID:  DID,
	}
	return cardSource
}

func (cs *BaseCardSource) GetCSID() string {
	return cs.CSID
}

func (cs *BaseCardSource) GetDID() string {
	return cs.DID
}
