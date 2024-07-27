package riff

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/88250/gulu"
	_ "github.com/mattn/go-sqlite3"
	"github.com/open-spaced-repetition/go-fsrs"

	"github.com/siyuan-note/filelock"
	"github.com/siyuan-note/logging"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"xorm.io/xorm"
)

type Riff interface {
	Query() []map[string]interface{}
	QueryCard() []Card

	// 底层 xorm DB的 Find 加锁包装
	Find(beans interface{}, condiBeans ...interface{}) error
	SetParams(algo Algo, params interface{})
	AddDeck(deck Deck) (newDeck Deck, err error)
	AddCardSource(cardSources []CardSource) (cardSourceList []CardSource, err error)
	AddCard(cards []Card) (cardList []Card, err error)
	Load(savePath string) (err error)
	Save(path string) error
	Due() []ReviewInfo
	// Review(card Card, rating Rating, RequestRetention float64)
	GetCardsByBlockIDs(blockIDs []string) (ret []ReviewInfo)
	Review(cardID string, rating Rating)
	CountCards() int
	GetBlockIDs() (ret []string)
}

type BaseRiff struct {
	Db                     *xorm.Engine
	GlobalRequestRetention float64
	MaxRequestRetention    float64
	MinRequestRetention    float64
	lock                   *sync.Mutex
	startTime              time.Time
	ParamsMap              map[Algo]interface{}
}

func NewBaseRiff() Riff {
	// orm, err := xorm.NewEngine("sqlite", ":memory:?_pragma=foreign_keys(1)")
	orm, err := xorm.NewEngine("sqlite3", ":memory:?mode=memory&cache=shared&loc=auto")
	if err != nil {
		return &BaseRiff{}
	}
	orm.Sync(new(BaseCard), new(BaseCardSource), new(BaseDeck), new(BaseHistory), new(ReviewLog))
	riff := BaseRiff{
		Db:                     orm,
		GlobalRequestRetention: 0.900,
		MaxRequestRetention:    0.999,
		MinRequestRetention:    0.500,
		lock:                   &sync.Mutex{},
		startTime:              time.Now(),
		ParamsMap:              map[Algo]interface{}{},
	}
	return &riff
}

func (br *BaseRiff) SetParams(algo Algo, params interface{}) {
	br.ParamsMap[algo] = params
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
	br.lock.Lock()
	defer br.lock.Unlock()
	_, err = br.Db.Insert(deck)
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

func (br *BaseRiff) batchCheck(table, field string, IDs []string) (existMap map[string]bool, err error) {

	br.lock.Lock()
	defer br.lock.Unlock()

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
		err = br.Db.Table(table).
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

func (br *BaseRiff) batchInsert(rowSlice interface{}) (err error) {
	// 获取数据类型

	br.lock.Lock()
	defer br.lock.Unlock()

	t := reflect.TypeOf(rowSlice)

	// 检查是否是切片类型
	if t.Kind() != reflect.Slice {
		fmt.Println("Not a slice")
		return
	}
	sliceValue := reflect.Indirect(reflect.ValueOf(rowSlice))
	Len := sliceValue.Len()
	session := br.Db.NewSession()
	defer session.Close()
	session.Begin()
	for i := 0; i < Len; i++ {
		_, err = session.Insert(sliceValue.Index(i).Interface())
		if err != nil {
			fmt.Printf("error on insert CardSource %s \n", err)
			continue
		}
	}
	err = session.Commit()
	return
}

func (br *BaseRiff) AddCardSource(cardSources []CardSource) (cardSourceList []CardSource, err error) {

	DIDs := make([]string, 0)
	existsCardSourceList := make([]CardSource, 0)
	for index := range cardSources {
		DIDs = append(DIDs, cardSources[index].GetDIDs()...)
	}

	existCSIDMap, err := br.batchCheck(
		"base_deck",
		"d_i_d",
		DIDs,
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

	br.batchInsert(existsCardSourceList)

	return
}

func (br *BaseRiff) AddCard(cards []Card) (cardList []Card, err error) {
	// 空实现
	// start := time.Now()

	CSIDs := make([]string, 0)
	existsCardList := make([]Card, 0)
	for index := range cards {
		cards[index].MarshalImpl()
		CSIDs = append(CSIDs, cards[index].GetCSID())
	}

	existCSIDMap, err := br.batchCheck(
		"base_card_source",
		"c_s_i_d",
		CSIDs,
	)

	for _, card := range cards {
		if existCSIDMap[card.GetCSID()] {
			existsCardList = append(existsCardList, card)
		}
	}

	br.batchInsert(existsCardList)

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

func (br *BaseRiff) Find(beans interface{}, condiBeans ...interface{}) error {
	br.lock.Lock()
	defer br.lock.Unlock()
	err := br.Db.Find(beans, condiBeans...)
	return err
}

func (br *BaseRiff) Get(beans ...interface{}) error {
	br.lock.Lock()
	defer br.lock.Unlock()
	_, err := br.Db.Get(beans...)
	return err
}

func (br *BaseRiff) Save(path string) (err error) {
	// 空实现

	decks := make([]BaseDeck, 0)
	cardSources := make([]BaseCardSource, 0)
	cards := make([]BaseCard, 0)
	err = br.Find(&decks)

	if err != nil {
		return
	}

	err = br.Find(&cardSources)

	if err != nil {
		return
	}

	err = br.Find(&cards)

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

func saveHistoryData(data interface{}, suffix SaveExt, saveDirPath string, time time.Time) (err error) {
	byteData, err := json.Marshal(data)
	if err != nil {
		logging.LogErrorf("marshal logs failed: %s", err)
		return
	}
	yyyyMMddHHmmss := time.Format("2006-01-02-15_04_05")
	savePath := path.Join(saveDirPath, yyyyMMddHHmmss+string(suffix))
	err = filelock.WriteFile(savePath, byteData)
	if err != nil {
		return
	}
	return
}

func (br *BaseRiff) SaveHistory(path string) (err error) {
	historys := make([]BaseHistory, 0)
	reviewLogs := make([]ReviewLog, 0)
	err = br.Find(&historys)
	if err != nil {
		return
	}

	for i := range historys {
		historys[i].UnmarshalImpl()
	}

	err = br.Find(&reviewLogs)
	if err != nil {
		return
	}
	err = saveHistoryData(historys, HistoryExt, path, br.startTime)
	if err != nil {
		logging.LogErrorf("write history file failed: %s", err)
	}
	err = saveHistoryData(reviewLogs, reviewLogExt, path, br.startTime)
	if err != nil {
		logging.LogErrorf("write log file failed: %s", err)
	}

	if err != nil {
		return
	}
	return
}

func (br *BaseRiff) LoadHistory(savePath string) (err error) {
	if !gulu.File.IsDir(savePath) {
		return errors.New("no a save path")
	}

	totalHistory := make([]History, 0)
	totalReviewLog := make([]ReviewLog, 0)

	filelock.Walk(savePath, func(walkPath string, info fs.FileInfo, err error) (reErr error) {
		if info.IsDir() {
			return
		}
		ext := filepath.Ext(walkPath)
		data, reErr := filelock.ReadFile(walkPath)
		switch SaveExt(ext) {

		case HistoryExt:
			historys := make([]BaseHistory, 0)
			json.Unmarshal(data, &historys)
			for _, history := range historys {
				copy := history
				totalHistory = append(totalHistory, &copy)
			}
		case reviewLogExt:
			reviewLog := make([]ReviewLog, 0)
			json.Unmarshal(data, &reviewLog)
			totalReviewLog = append(totalReviewLog, reviewLog...)
		}

		return
	})
	br.batchInsert(totalHistory)
	br.batchInsert(totalReviewLog)
	return
}

func (br *BaseRiff) Load(savePath string) (err error) {

	if !gulu.File.IsDir(savePath) {
		return errors.New("no a save path")
	}
	totalDecks := make([]Deck, 0)
	totalCards := make([]Card, 0)
	totalCardSources := make([]CardSource, 0)

	filelock.Walk(savePath, func(walkPath string, info fs.FileInfo, err error) (reErr error) {
		if info.IsDir() {
			return
		}
		ext := filepath.Ext(walkPath)
		// 后期性能改进点：把读文件位置后移
		data, reErr := filelock.ReadFile(walkPath)
		switch SaveExt(ext) {

		case DeckExt:
			decks := make([]BaseDeck, 0)
			json.Unmarshal(data, &decks)
			for _, deck := range decks {
				copy := deck
				totalDecks = append(totalDecks, &copy)
			}

		case CardExt:
			cards := make([]BaseCard, 0)
			json.Unmarshal(data, &cards)
			for _, card := range cards {
				copy := card
				totalCards = append(totalCards, &copy)
			}

		case CardSourceExt:
			cardSources := make([]BaseCardSource, 0)
			json.Unmarshal(data, &cardSources)
			for _, cardSource := range cardSources {
				copy := cardSource
				totalCardSources = append(totalCardSources, &copy)
			}
		}

		return
	})
	for _, deck := range totalDecks {
		br.AddDeck(deck)
	}

	br.AddCardSource(totalCardSources)
	br.AddCard(totalCards)

	go br.LoadHistory(savePath)
	return

}

func (br *BaseRiff) Due() []ReviewInfo {
	br.lock.Lock()
	defer br.lock.Unlock()

	ris := make([]ReviewInfo, 0)
	now := time.Now()

	err := br.Db.Table("base_card").
		Select("base_card_source.*, base_card.*").
		Join("INNER", "base_card_source", "base_card_source.c_s_i_d = base_card.c_s_i_d").
		Where("base_card.due < ?", now).
		Find(&ris)

	if err != nil {
		fmt.Printf("%s", err)
	}
	for i := range ris {
		ris[i].UnmarshalImpl()
	}
	return ris
}

func (br *BaseRiff) GetCardsByBlockIDs(blockIDs []string) (ret []ReviewInfo) {
	br.lock.Lock()
	defer br.lock.Unlock()
	ret = make([]ReviewInfo, 0)
	// bscs := make([]BaseCardSource, 0)

	reviewInfo := new(ReviewInfo)
	rows, err := br.Db.Table("base_card").
		Join("inner", "base_card_source", "base_card.c_s_i_d = base_card_source.c_s_i_d").
		Rows(reviewInfo)
	if err != nil {
		return
	}
	queryBlockIDs := make(map[string]bool, 0)
	for _, blockID := range blockIDs {
		queryBlockIDs[blockID] = true
	}
	for rows.Next() {
		reviewInfo := new(ReviewInfo)
		existsNum := 0
		rows.Scan(reviewInfo)
		for _, blockID := range reviewInfo.BlockIDs {
			if queryBlockIDs[blockID] {
				existsNum += 1
			}
		}
		if existsNum > 0 {
			ret = append(ret, *reviewInfo)
		}
	}
	if err != nil {
		fmt.Printf("%s", err)
	}
	for i := range ret {
		ret[i].UnmarshalImpl()
	}
	return
}

func (br *BaseRiff) innerReview(card Card, rating Rating, RequestRetention float64) {
	br.lock.Lock()
	defer br.lock.Unlock()
	now := time.Now()

	history := NewBaseHistory(card)
	reviewlog := NewReviewLog(history, rating)
	_, err := br.Db.Insert(history)
	if err != nil {
		logging.LogErrorf("error insert history %s \n", err)
	}
	_, err = br.Db.Insert(reviewlog)
	if err != nil {
		logging.LogErrorf("error insert reviewLog %s \n", err)
	}

	switch card.GetAlgo() {
	case AlgoFSRS:
		fsrsCard := (card.Impl()).(fsrs.Card)
		params := br.ParamsMap[AlgoFSRS].(fsrs.Parameters)
		params.RequestRetention = RequestRetention
		schedulingInfo := params.Repeat(fsrsCard, now)[rating.ToFsrs()]
		newCard := schedulingInfo.Card
		card.SetDue(newCard.Due)
		card.SetImpl(newCard)
	}
	switch rating {
	case Again:
		card.SetLapses(card.GetLapses() + 1)
		card.SetReps(card.GetReps() + 1)
	default:
		card.SetReps(card.GetReps() + 1)
	}
	card.MarshalImpl()
	_, err = br.Db.Where("c_i_d = ?", card.ID()).Update(card)
	if err != nil {
		fmt.Printf("update card err:%s\n", err)
	}
}

func (br *BaseRiff) Review(cardID string, rating Rating) {
	card := BaseCard{
		CID: cardID,
	}
	br.Get(&card)
	card.UnmarshalImpl()
	RequestRetention := br.getRequestRetention(&card)
	br.innerReview(&card, rating, RequestRetention)

}
func (br *BaseRiff) getRequestRetention(card Card) float64 {
	priority := card.GetPriority()
	requestRetention := br.GlobalRequestRetention
	switch {
	case priority >= 0 && priority <= 0.5:
		requestRetention = (br.GlobalRequestRetention - br.MinRequestRetention) / (0.5 - 0) * (priority - 0)
	case priority > 0.5 && priority <= 1:
		requestRetention = (br.MaxRequestRetention - br.GlobalRequestRetention) / (1 - 0.5) * (priority - 0.5)
	}
	return requestRetention
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

var RatingToFsrs = map[Rating]fsrs.Rating{
	Again: fsrs.Again,
	Hard:  fsrs.Hard,
	Good:  fsrs.Good,
	Easy:  fsrs.Easy,
}

func (rate Rating) ToFsrs() fsrs.Rating {
	return RatingToFsrs[rate]
}

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
	reviewLogExt  = ".revlog"
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
