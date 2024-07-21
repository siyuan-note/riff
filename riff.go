package riff

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/88250/gulu"

	"github.com/siyuan-note/filelock"
	"github.com/siyuan-note/logging"
	"github.com/syndtr/goleveldb/leveldb/errors"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

type Riff interface {
	Query() []map[string]interface{}
	QueryCard() []Card
	AddDeck(deck Deck) (newDeck Deck, err error)
	AddCardSource(cardSources []CardSource) (cardSourceList []CardSource, err error)
	AddCard(cards []Card) (cardList []Card, err error)
	Load(savePath string) (err error)
	Save(path string) error
	Due() []Card
	Review(card Card, rating Rating, RequestRetention float64)
	CountCards() int
	GetBlockIDs() (ret []string)
}

type BaseRiff struct {
	db                  *xorm.Engine
	MaxRequestRetention float64
	MinRequestRetention float64
}

func NewBaseRiff() Riff {
	// orm, err := xorm.NewEngine("sqlite", ":memory:?_pragma=foreign_keys(1)")
	orm, err := xorm.NewEngine("sqlite", "file::memory:?mode=memory&cache=shared&loc=auto")
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

func batchCheck(table, field string, IDs []string, db xorm.Interface) (existMap map[string]bool, err error) {

	const MAX_BATCH = 5000

	existMap = map[string]bool{}
	existsIDs := make([]string, 0)
	IDs = gulu.Str.RemoveDuplicatedElem(IDs)
	IDsLength := len(IDs)

	for i := 0; i < IDsLength; i += MAX_BATCH {
		end := i + MAX_BATCH
		if end > IDsLength {
			end = IDsLength
		}
		subIDs := IDs[i:end]
		err = db.Table(table).
			In(field, subIDs).
			Cols(field).
			Find(&existsIDs)
	}

	for _, existsID := range existsIDs {
		existMap[existsID] = true
	}
	for _, ID := range IDs {
		if !existMap[ID] {
			err = errors.New(fmt.Sprintf("no exit field in %s : %s = %s", table, field, ID))
			return
		}
	}

	return
}

// func (br *BaseRiff) AddCardSource(cardSources []CardSource) (cardSourceList []CardSource, err error) {

// 	for _, cardSource := range cardSources {
// 		DIDs := cardSource.GetDIDs()
// 		for _, DID := range DIDs {
// 			err = checkExist(br.db, &BaseDeck{
// 				DID: DID,
// 			})
// 		}

// 		if err != nil {
// 			return cardSourceList, err
// 		}
// 		_, err = br.db.Insert(cardSource)
// 		if err != nil {
// 			return cardSourceList, err
// 		}
// 		cardSourceList = append(cardSourceList, cardSource)
// 	}
// 	return
// }

func (br *BaseRiff) AddCardSource(cardSources []CardSource) (cardSourceList []CardSource, err error) {

	DIDs := make([]string, 0)
	existsCardSourceList := make([]CardSource, 0)
	for index := range cardSources {
		DIDs = append(DIDs, cardSources[index].GetDIDs()...)
	}

	existCSIDMap, err := batchCheck(
		"base_deck",
		"d_i_d",
		DIDs,
		br.db,
	)

	for _, cardSource := range cardSources {
		DIDs := cardSource.GetDIDs()
		unExist := 0
		for _, DID := range DIDs {
			if !existCSIDMap[DID] {
				unExist += 1
			}
		}
		if unExist == 0 {
			existsCardSourceList = append(existsCardSourceList, cardSource)
		}
	}

	session := br.db.NewSession()
	defer session.Close()
	session.Begin()

	for _, cardSource := range existsCardSourceList {
		// session.Prepare()
		_, err = session.Insert(cardSource)
		if err != nil {
			return
		}
	}

	err = session.Commit()

	testList := make([]BaseCardSource, 0)
	br.db.Find(&testList)

	return
}

func (br *BaseRiff) AddCard(cards []Card) (cardList []Card, err error) {
	// 空实现
	start := time.Now()
	CSIDs := make([]string, 0)
	existsCardList := make([]Card, 0)
	for index := range cards {
		cards[index].MarshalImpl()
		CSIDs = append(CSIDs, cards[index].GetCSID())
	}

	existCSIDMap, err := batchCheck(
		"base_card_source",
		"c_s_i_d",
		CSIDs,
		br.db,
	)

	for _, card := range cards {
		if existCSIDMap[card.GetCSID()] {
			existsCardList = append(existsCardList, card)
		}
	}

	session := br.db.NewSession()
	defer session.Close()
	session.Begin()

	for _, card := range existsCardList {
		// session.Prepare()
		_, err = session.Insert(card)
		if err != nil {
			fmt.Printf("error on insert Card %s \n", err)
			return
		}
	}

	err = session.Commit()
	test := make([]BaseCard, 0)
	br.db.Find(&test)
	fmt.Printf("add card taken %s \n", time.Since(start))
	return
}

func saveData(data interface{}, suffix SaveExt, saveDirPath string) (err error) {
	byteData, err := json.Marshal(data)
	if err != nil {
		logging.LogErrorf("marshal logs failed: %s", err)
		return
	}
	savePath := path.Join(saveDirPath, "siyuan"+string(suffix))
	err = filelock.WriteFile(savePath, byteData)
	if err != nil {
		logging.LogErrorf("write riff file failed: %s", err)
		return
	}
	return
}

func (br *BaseRiff) Save(path string) (err error) {
	// 空实现

	decks := make([]BaseDeck, 0)
	cardSources := make([]BaseCardSource, 0)
	cards := make([]BaseCard, 0)
	err = br.db.Find(&decks)

	if err != nil {
		return
	}

	err = br.db.Find(&cardSources)

	if err != nil {
		return
	}

	err = br.db.Find(&cards)

	if err != nil {
		return
	}

	err = br.SaveHistory(path)

	if err != nil {
		return
	}

	if !gulu.File.IsDir(path) {
		if err = os.MkdirAll(path, 0755); nil != err {
			return
		}
	}
	err = saveData(decks, DeckExt, path)
	if err != nil {
		return
	}
	err = saveData(cardSources, CardSourceExt, path)
	if err != nil {
		return
	}

	// card需要先反序列化Impl
	for index := range cards {
		cards[index].UnmarshalImpl()
	}
	err = saveData(cards, CardExt, path)
	if err != nil {
		return
	}

	return
}

func (br *BaseRiff) SaveHistory(path string) (err error) {
	historys := make([]BaseHistory, 0)
	err = br.db.Find(&historys)
	if err != nil {
		return
	}
	err = saveData(historys, HistoryExt, path)
	if err != nil {
		return
	}
	return
}

func (br *BaseRiff) Load(savePath string) (err error) {
	// data, err := filelock.ReadFile(savePath)
	if !gulu.File.IsDir(savePath) {
		return errors.New("no a save path")
	}
	totalDecks := make([]Deck, 0)
	totalCards := make([]Card, 0)
	totalCardSources := make([]CardSource, 0)
	totalHistory := make([]BaseHistory, 0)
	filelock.Walk(savePath, func(walkPath string, info fs.FileInfo, err error) (reErr error) {
		if info.IsDir() {
			return
		}
		ext := filepath.Ext(walkPath)
		data, reErr := filelock.ReadFile(walkPath)
		switch SaveExt(ext) {

		case DeckExt:
			decks := make([]BaseDeck, 0)
			json.Unmarshal(data, &decks)
			for _, deck := range decks {
				totalDecks = append(totalDecks, &deck)
			}

		case CardExt:
			cards := make([]BaseCard, 0)
			json.Unmarshal(data, &cards)
			for _, card := range cards {
				totalCards = append(totalCards, &card)
			}

		case CardSourceExt:
			cardSources := make([]BaseCardSource, 0)
			json.Unmarshal(data, &cardSources)
			for _, cardSource := range cardSources {
				totalCardSources = append(totalCardSources, &cardSource)
			}

		case HistoryExt:
			history := make([]BaseHistory, 0)
			json.Unmarshal(data, &history)
			totalHistory = append(totalHistory, history...)
		}

		return
	})
	for _, deck := range totalDecks {
		br.AddDeck(deck)
	}

	br.AddCardSource(totalCardSources)
	br.AddCard(totalCards)

	return
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

type SaveExt string

const (
	CardExt       = ".cards"
	CardSourceExt = ".cardSources"
	DeckExt       = ".decks"
	HistoryExt    = ".history"
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
