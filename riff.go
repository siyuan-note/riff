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
	//WaitLoad会等待至从磁盘加载完成
	WaitLoad() (err error)
	Save(path string) error
	Due() []ReviewInfo
	// Review(card Card, rating Rating, RequestRetention float64)
	GetCardsByBlockIDs(blockIDs []string) (ret []ReviewInfo)
	Review(cardID string, rating Rating)
	CountCards() int
	GetBlockIDs() (ret []string)

	//SetDeck设置Deck， 无锁
	SetDeck(deck Deck) (err error)

	//GetDeck获取Deck， 无锁
	GetDeck(did string) (deck Deck, err error)

	//SetCardSource设置CardSource， 无锁
	SetCardSource(cs CardSource) (err error)

	//GetCardSource获取CardSource， 无锁
	GetCardSource(csid string) (cs CardSource, err error)

	//SetCard设置Card， 无锁
	SetCard(card Card) (err error)

	//GetCard获取Card， 无锁
	GetCard(cardID string) (card Card, err error)
}

type BaseRiff struct {
	db                     *xorm.Engine
	GlobalRequestRetention float64
	MaxRequestRetention    float64
	MinRequestRetention    float64
	lock                   *sync.Mutex
	load                   *sync.Mutex
	startTime              time.Time
	ParamsMap              map[Algo]interface{}
	deckMap                map[string]Deck
	cardSourceMap          map[string]CardSource
	cardMap                map[string]Card
}

func NewBaseRiff() Riff {
	// orm, err := xorm.NewEngine("sqlite", ":memory:?_pragma=foreign_keys(1)")
	orm, err := xorm.NewEngine("sqlite3", ":memory:?mode=memory&cache=shared&loc=auto")
	orm.Exec(`
	create view block_id_to_card_source as 
	select c_s_i_d,value
	from 
	base_card_source , json_each(base_card_source.block_i_ds)`)
	if err != nil {
		return &BaseRiff{}
	}
	orm.Sync(new(BaseCard), new(BaseCardSource), new(BaseDeck), new(BaseHistory), new(ReviewLog))
	riff := BaseRiff{
		db:                     orm,
		GlobalRequestRetention: 0.900,
		MaxRequestRetention:    0.999,
		MinRequestRetention:    0.500,
		lock:                   &sync.Mutex{},
		load:                   &sync.Mutex{},
		startTime:              time.Now(),
		ParamsMap:              map[Algo]interface{}{},
		deckMap:                map[string]Deck{},
		cardSourceMap:          map[string]CardSource{},
		cardMap:                map[string]Card{},
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
	br.deckMap[deck.GetDID()] = deck
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
		err = br.db.Table(table).
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
	session := br.db.NewSession()
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
			// 添加到 cardSourceMap
			br.cardSourceMap[cardSource.GetCSID()] = cardSource
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
			br.cardMap[card.ID()] = card
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
	err := br.db.Find(beans, condiBeans...)
	return err
}

func (br *BaseRiff) Get(beans ...interface{}) error {
	br.lock.Lock()
	defer br.lock.Unlock()
	_, err := br.db.Get(beans...)
	return err
}

func (br *BaseRiff) Save(path string) (err error) {
	decks := make([]BaseDeck, 0)
	cardSources := make([]BaseCardSource, 0)
	cards := make([]BaseCard, 0)

	for _, deck := range br.deckMap {
		decks = append(decks, *(deck.(*BaseDeck)))
	}
	for _, cardSource := range br.cardSourceMap {
		cardSources = append(cardSources, *(cardSource.(*BaseCardSource)))
	}
	for _, card := range br.cardMap {
		cards = append(cards, *(card.(*BaseCard)))
	}

	if !gulu.File.IsDir(path) {
		if err = os.MkdirAll(path, 0755); nil != err {
			return
		}
	}
	err = saveData(decks, DeckExt, path)
	if err != nil {
		fmt.Printf("err in save riff data: %s \n", err)
	}
	err = saveData(cardSources, CardSourceExt, path)
	if err != nil {
		fmt.Printf("err in save riff data: %s \n", err)
	}
	err = saveData(cards, CardExt, path)
	if err != nil {
		fmt.Printf("err in save riff data: %s \n", err)
	}

	err = br.SaveHistory(path)

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
	br.load.Lock()
	defer br.load.Unlock()

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

func (br *BaseRiff) WaitLoad() (err error) {
	br.load.Lock()
	defer br.load.Unlock()
	return
}

func (br *BaseRiff) Due() (ret []ReviewInfo) {
	now := time.Now()

	qr := br.newReviewInfoQuery()
	qr, err := qr.ByDue(now)
	if err != nil {
		return
	}
	ret, err = qr.Query()
	if err != nil {
		ret = make([]ReviewInfo, 0)
		return
	}
	return
}

func (br *BaseRiff) GetCardsByBlockIDs(blockIDs []string) (ret []ReviewInfo) {

	qr := br.newReviewInfoQuery()
	qr, err := qr.ByBlockIDs(blockIDs)

	if err != nil {
		fmt.Println(err)
	}

	ret, err = qr.Query()

	if err != nil {
		fmt.Println(err)
	}

	return
}

func (br *BaseRiff) innerReview(card Card, rating Rating, RequestRetention float64) {
	// TODO review时更新card status
	br.lock.Lock()
	defer br.lock.Unlock()
	now := time.Now()

	history := NewBaseHistory(card)
	reviewlog := NewReviewLog(history, rating)
	_, err := br.db.Insert(history)
	if err != nil {
		logging.LogErrorf("error insert history %s \n", err)
	}
	_, err = br.db.Insert(reviewlog)
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
	br.SetCard(card)
}

func (br *BaseRiff) Review(cardID string, rating Rating) {
	br.lock.Lock()
	card, err := br.GetCard(cardID)
	br.lock.Unlock()
	if err != nil {
		return
	}

	RequestRetention := br.getRequestRetention((card))
	br.innerReview(card, rating, RequestRetention)

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

func (br *BaseRiff) SetDeck(deck Deck) (err error) {
	br.deckMap[deck.GetDID()] = deck
	_, err = br.db.Where("d_i_d = ?", deck.GetDID()).Update(deck)
	return
}

func (br *BaseRiff) GetDeck(did string) (deck Deck, err error) {
	deck, exist := br.deckMap[did]
	if !exist {
		err = fmt.Errorf("deck %s is not exist", did)
	}
	return
}

func (br *BaseRiff) SetCardSource(cs CardSource) (err error) {
	br.cardSourceMap[cs.GetCSID()] = cs
	_, err = br.db.Where("c_s_i_d = ?", cs.GetCSID()).Update(cs)
	return
}

func (br *BaseRiff) GetCardSource(csid string) (cs CardSource, err error) {
	cs, exist := br.cardSourceMap[csid]
	if !exist {
		err = fmt.Errorf("CardSource %s is not exist", csid)
	}
	return
}

func (br *BaseRiff) SetCard(card Card) (err error) {
	br.cardMap[card.ID()] = card
	card.MarshalImpl()
	_, err = br.db.Where("c_i_d = ?", card.ID()).Update(card)
	if err != nil {
		fmt.Printf("update card err:%s\n", err)
	}

	return
}

func (br *BaseRiff) GetCard(cardID string) (card Card, err error) {
	card, exist := br.cardMap[cardID]
	if !exist {
		err = fmt.Errorf("card %s is not exist", cardID)
	}
	return
}

// QueryReviewInfo 是对 ReviewInfo 搜索查询操作的抽象和封装
type QueryReviewInfo interface {
	ByBlockIDs(blockIDs []string) (ret QueryReviewInfo, err error)
	ByDue(dueTime time.Time) (ret QueryReviewInfo, err error)
	Query() (ret []ReviewInfo, err error)
	Conut() (ret int64, err error)
}
type baseQueryReviewInfo struct {
	br      *BaseRiff
	db      *xorm.Engine
	sission *xorm.Session
	lock    *sync.Mutex
}

func (br *BaseRiff) newReviewInfoQuery() (ret QueryReviewInfo) {
	br.lock.Lock()
	defer br.lock.Unlock()
	session := br.db.Table("base_card").
		Join("inner", "base_card_source", "base_card.c_s_i_d = base_card_source.c_s_i_d")
	ret = &baseQueryReviewInfo{
		br:      br,
		db:      br.db,
		sission: session,
		lock:    br.lock,
	}
	return
}

func (qr *baseQueryReviewInfo) ByBlockIDs(blockIDs []string) (ret QueryReviewInfo, err error) {
	qr.lock.Lock()
	defer qr.lock.Unlock()

	ret = qr

	csidList := make([]string, 0)
	queryLen := len(blockIDs)
	if queryLen > MAX_QUERY_PARAMS {
		queryLen = MAX_QUERY_PARAMS
	}
	queryBlockIDs := blockIDs[0:queryLen]

	err = qr.db.Table("base_card_source").
		Join("inner", "block_id_to_card_source", "base_card_source.c_s_i_d = block_id_to_card_source.c_s_i_d").
		In("block_id_to_card_source.value", queryBlockIDs).
		Distinct("base_card_source.c_s_i_d").Find(&csidList)

	if err != nil {
		return
	}

	queryCsidLen := len(csidList)
	if queryCsidLen > MAX_QUERY_PARAMS {
		queryCsidLen = MAX_QUERY_PARAMS
	}
	queryCsidList := csidList[0:queryCsidLen]

	qr.sission.In("base_card_source.c_s_i_d", queryCsidList)
	return
}

func (qr *baseQueryReviewInfo) ByDue(dueTime time.Time) (ret QueryReviewInfo, err error) {
	qr.lock.Lock()
	defer qr.lock.Unlock()

	ret = qr

	qr.sission.Where("base_card.due < ?", dueTime)
	return
}

func (qr *baseQueryReviewInfo) Query() (ret []ReviewInfo, err error) {
	qr.lock.Lock()
	defer qr.lock.Unlock()
	ret = make([]ReviewInfo, 0)
	reviewInfoIDList, err := qr.sission.
		Cols("base_card.c_i_d", "base_card_source.c_s_i_d").
		QueryString()
	if err != nil {
		return
	}
	for _, item := range reviewInfoIDList {
		card, err := qr.br.GetCard(item["c_i_d"])
		if err != nil {
			continue
		}
		cs, err := qr.br.GetCardSource(item["c_s_i_d"])
		if err != nil {
			continue
		}
		ret = append(ret, ReviewInfo{
			BaseCard:       *(card.(*BaseCard)),
			BaseCardSource: *(cs.(*BaseCardSource)),
		})
	}

	return
}

func (qr *baseQueryReviewInfo) Conut() (ret int64, err error) {
	qr.lock.Lock()
	defer qr.lock.Unlock()
	ri := new(ReviewInfo)
	ret, err = qr.sission.Count(ri)
	return
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

const MAX_QUERY_PARAMS = 3_0000

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
