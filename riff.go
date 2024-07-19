package riff

import (
	"math/rand"
	"time"

	"github.com/syndtr/goleveldb/leveldb/errors"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

type Riff interface {
	Load(savePath string)
	Query() []map[string]interface{}
	QueryCard() []Card
	AddDeck(deck Deck) (newDeck Deck, err error)
	AddCardSource(cardSources []CardSource) (cardSourceList []CardSource, err error)
	AddCard(cards []Card) (cardList []Card, err error)
	Save(path string) error
	Due() []Card
	Review(card Card, rating Rating, RequestRetention float64)
	CountCards() int
	GetBlockIDs() (ret []string)
}

type BaseRiff struct {
	db                  xorm.Interface
	MaxRequestRetention float64
	MinRequestRetention float64
}

func NewBaseRiff() Riff {
	orm, err := xorm.NewEngine("sqlite", ":memory:?_pragma=foreign_keys(1)")
	if err != nil {
		return &BaseRiff{}
	}
	orm.Sync(new(BaseCard), new(BaseCardSource), new(BaseDeck), new(BaseHistory))
	riff := BaseRiff{
		db:                  orm,
		MaxRequestRetention: 0.95,
		MinRequestRetention: 0.5,
	}
	return &riff
}

func (riff *BaseRiff) Load(savePath string) {
	// data, err := filelock.ReadFile(savePath)
}

func (br *BaseRiff) Query() []map[string]interface{} {
	// 空实现
	return nil
}

func (br *BaseRiff) QueryCard() []Card {
	// 空实现
	return nil
}

func (br *BaseRiff) AddDeck(deck Deck) (newDeck Deck, err error) {

	_, err = br.db.Insert(deck)
	newDeck = deck
	return
}

func checkExist(db xorm.Interface, data interface{}) error {
	exist, err := db.Exist(data)
	if !exist || err != nil {
		return errors.New("forgin no exist")
	}
	return nil
}

func (br *BaseRiff) AddCardSource(cardSources []CardSource) (cardSourceList []CardSource, err error) {

	for _, cardSource := range cardSources {
		err = checkExist(br.db, &BaseDeck{
			DID: cardSource.GetDID(),
		})
		if err != nil {
			return cardSourceList, err
		}
		_, err = br.db.Insert(cardSource)
		if err != nil {
			return cardSourceList, err
		}
		cardSourceList = append(cardSourceList, cardSource)
	}
	return
}

func (br *BaseRiff) AddCard(cards []Card) (cardList []Card, err error) {
	// 空实现
	for _, card := range cards {
		err = checkExist(br.db, &BaseCardSource{
			CSID: card.GetCSID(),
		})
		if err != nil {
			return cardList, err
		}
		card.MarshalImpl()
		_, err = br.db.Insert(card)
		if err != nil {
			return cardList, err
		}
		cardList = append(cardList, card)
	}
	return
}

func (br *BaseRiff) Save(path string) error {
	// 空实现
	return nil
}

func (br *BaseRiff) Due() []Card {
	// 空实现
	cards := make([]Card, 0)
	baseCards := make([]BaseCard, 0)
	now := time.Now()
	br.db.Where("due < ?", now).Find(&baseCards)
	for _, bc := range baseCards {
		UnmarshalImpl(&bc)
		cards = append(cards, &bc)
	}
	return cards
}

func (br *BaseRiff) Review(card Card, rating Rating, RequestRetention float64) {
	// 空实现
}

func (br *BaseRiff) CountCards() int {
	// 空实现
	return 0
}

func (br *BaseRiff) GetBlockIDs() (ret []string) {
	// 空实现
	return nil
}

// Rating 描述了闪卡复习的评分。
type Rating int8

const (
	Again Rating = iota + 1 // 完全不会，必须再复习一遍
	Hard                    // 有点难
	Good                    // 一般
	Easy                    // 很容易
)

// Algo 描述了闪卡复习算法的名称。
type Algo string

const (
	AlgoFSRS Algo = "fsrs"
	AlgoSM2  Algo = "sm2"
)

// State 描述了闪卡的状态。
type State int8

const (
	New State = iota
	Learning
	Review
	Relearning
)

const builtInDeck = "20240718214745-q7ocvvi"

func newID() string {
	now := time.Now()
	return now.Format("20060102150405") + "-" + randStr(7)
}

func randStr(length int) string {
	letter := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}
