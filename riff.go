package riff

type Riff interface {
	Load(savePath string)
	Query() []map[string]interface{}
	Save() error
	Due() []Card
	Review(card Card, rating Rating)
	CountCards() int
	GetBlockIDs() (ret []string)
}

type BaseRiff struct {
	db interface{}
}

func NewRiff() Riff {
	return new(BaseRiff)
}

func (riff *BaseRiff) Load(savePath string) {
	// data, err := filelock.ReadFile(savePath)
}

func (riff *BaseRiff) Query() []map[string]interface{} {
	return nil
}
func (riff *BaseRiff) Save() error {
	return nil

}
func (riff *BaseRiff) Due() []Card {
	return nil

}
func (riff *BaseRiff) Review(card Card, rating Rating) {
	return

}
func (riff *BaseRiff) CountCards() int {
	return 0

}
func (riff *BaseRiff) GetBlockIDs() (ret []string) {
	return nil

}
