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
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/88250/gulu"
	"github.com/open-spaced-repetition/go-fsrs"
	"github.com/siyuan-note/filelock"
	"github.com/siyuan-note/logging"
	"github.com/vmihailenco/msgpack/v5"
)

type FSRSStore struct {
	*BaseStore

	cardSources map[string]*BaseCardSource
	cards       map[string]*FSRSCard
	params      fsrs.Parameters
}

func NewFSRSStore(id, saveDir string, requestRetention float64, maximumInterval int, weights string) *FSRSStore {
	params := fsrs.DefaultParam()
	params.RequestRetention = requestRetention
	params.MaximumInterval = float64(maximumInterval)
	params.W = [17]float64{}
	for i, w := range strings.Split(weights, ",") {
		w = strings.TrimSpace(w)
		params.W[i], _ = strconv.ParseFloat(w, 64)
	}

	return &FSRSStore{
		BaseStore:   NewBaseStore(id, "fsrs", saveDir),
		cardSources: map[string]*BaseCardSource{},
		cards:       map[string]*FSRSCard{},
		params:      params,
	}
}

func (store *FSRSStore) AddCard(id, blockID string) Card {
	store.lock.Lock()
	defer store.lock.Unlock()

	cardSourceID := newID()
	c := fsrs.NewCard()
	card := &FSRSCard{BaseCard: &BaseCard{CID: id, SID: cardSourceID}, C: &c}
	store.cards[id] = card

	cardSource := &BaseCardSource{
		SID:     cardSourceID,
		CType:   builtInCardType,
		CIDMap:  map[string]string{builtInCardIDMapKey: id},
		Context: map[string]string{builtInContext: blockID}}
	store.cardSources[cardSourceID] = cardSource
	return card
}

func (store *FSRSStore) GetCard(id string) Card {
	store.lock.Lock()
	defer store.lock.Unlock()

	ret := store.cards[id]
	if nil == ret {
		return nil
	}
	return ret
}

func (store *FSRSStore) SetCard(card Card) {
	store.lock.Lock()
	defer store.lock.Unlock()

	store.cards[card.ID()] = card.(*FSRSCard)
}

func (store *FSRSStore) RemoveCard(id string) Card {
	store.lock.Lock()
	defer store.lock.Unlock()

	card := store.cards[id]
	if nil == card {
		return nil
	}
	cardSourceID := card.CardSourceID()
	cardSource := store.cardSources[cardSourceID]
	cardSource.GetCardIDMap()
	if nil != cardSource {
		cardSource.RemoveCardID(id)
	}
	cardMap := cardSource.GetCardIDMap()
	if 0 == len(cardMap) {
		delete(store.cardSources, cardSourceID)
	}
	delete(store.cards, id)
	return card
}

func getCardSourceRelatedBlockID(cardSource CardSource) (ret []string) {
	blockIDsStr := cardSource.GetContext()[builtInContext]
	blockIDsStr = strings.Replace(blockIDsStr, " ", "", -1)
	ret = strings.Split(blockIDsStr, ",")
	ret = gulu.Str.RemoveDuplicatedElem(ret)
	return
}

// 根据 blockID 从 cardSource Context["blockIDs"] 里查询，返回一个不重复的CardIDs
// 这个操作没有加锁，调用者必须自己加锁
func (store *FSRSStore) getCardIDsByBlockID(blockID string) (ret []string) {
	for _, cardSource := range store.cardSources {
		cardSourceBlockIDs := getCardSourceRelatedBlockID(cardSource)
		if gulu.Str.Contains(blockID, cardSourceBlockIDs) {
			for _, cardID := range cardSource.GetCardIDs() {
				if _, ok := store.cards[cardID]; ok {
					ret = append(ret, cardID)
				}
			}
		}
	}
	ret = gulu.Str.RemoveDuplicatedElem(ret)

	return
}
func (store *FSRSStore) GetCardsByBlockID(blockID string) (ret []Card) {
	store.lock.Lock()
	defer store.lock.Unlock()

	cardIDs := store.getCardIDsByBlockID(blockID)
	for _, cardID := range cardIDs {
		ret = append(ret, store.cards[cardID])
	}
	return
}

func (store *FSRSStore) GetCardsByBlockIDs(blockIDs []string) (ret []Card) {
	store.lock.Lock()
	defer store.lock.Unlock()

	var cardIDs []string

	blockIDs = gulu.Str.RemoveDuplicatedElem(blockIDs)

	for _, blockID := range blockIDs {
		cardIDs = append(cardIDs, store.getCardIDsByBlockID(blockID)...)
	}

	cardIDs = gulu.Str.RemoveDuplicatedElem(cardIDs)
	for _, cardID := range cardIDs {
		ret = append(ret, store.cards[cardID])
	}
	return
}

func (store *FSRSStore) GetNewCardsByBlockIDs(blockIDs []string) (ret []Card) {
	store.lock.Lock()
	defer store.lock.Unlock()

	cards := store.GetCardsByBlockIDs(blockIDs)
	for _, card := range cards {
		c := card.Impl().(*fsrs.Card)
		if !c.LastReview.IsZero() {
			continue
		}
		ret = append(ret, card)
	}
	return
}

func (store *FSRSStore) GetDueCardsByBlockIDs(blockIDs []string) (ret []Card) {
	store.lock.Lock()
	defer store.lock.Unlock()

	now := time.Now()

	cards := store.GetCardsByBlockIDs(blockIDs)
	for _, card := range cards {
		c := card.Impl().(*fsrs.Card)
		if now.Before(c.Due) {
			continue
		}
		ret = append(ret, card)
	}
	return
}

func (store *FSRSStore) GetBlockIDs() (ret []string) {
	store.lock.Lock()
	defer store.lock.Unlock()

	ret = []string{}
	for _, cardSource := range store.cardSources {
		blockIDs := getCardSourceRelatedBlockID(cardSource)
		ret = append(ret, blockIDs...)
	}
	ret = gulu.Str.RemoveDuplicatedElem(ret)
	sort.Strings(ret)
	return
}

func (store *FSRSStore) CountCards() int {
	store.lock.Lock()
	defer store.lock.Unlock()

	return len(store.cards)
}

func (store *FSRSStore) Review(cardId string, rating Rating) (ret *Log) {
	store.lock.Lock()
	defer store.lock.Unlock()

	now := time.Now()
	card := store.cards[cardId]
	if nil == card {
		logging.LogWarnf("not found card [id=%s] to review", cardId)
		return
	}

	schedulingInfo := store.params.Repeat(*card.C, now)
	updated := schedulingInfo[fsrs.Rating(rating)].Card
	card.SetImpl(&updated)
	store.cards[cardId] = card

	reviewLog := schedulingInfo[fsrs.Rating(rating)].ReviewLog
	ret = &Log{
		ID:            newID(),
		CardID:        cardId,
		Rating:        rating,
		ScheduledDays: reviewLog.ScheduledDays,
		ElapsedDays:   reviewLog.ElapsedDays,
		Reviewed:      reviewLog.Review.Unix(),
		State:         State(reviewLog.State),
	}
	return
}

func (store *FSRSStore) Dues() (ret []Card) {
	store.lock.Lock()
	defer store.lock.Unlock()

	now := time.Now()
	for _, card := range store.cards {
		c := card.Impl().(*fsrs.Card)
		if now.Before(c.Due) {
			continue
		}

		schedulingInfos := store.params.Repeat(*c, now)
		nextDues := map[Rating]time.Time{}
		for rating, schedulingInfo := range schedulingInfos {
			nextDues[Rating(rating)] = schedulingInfo.Card.Due
		}
		card.SetNextDues(nextDues)
		ret = append(ret, card)
	}
	return
}

func (store *FSRSStore) Load() (err error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	store.cards = map[string]*FSRSCard{}
	p := store.getMsgPackPath()
	if !filelock.IsExist(p) {
		return
	}

	data, err := filelock.ReadFile(p)
	if nil != err {
		logging.LogErrorf("load cards failed: %s", err)
	}
	if err = msgpack.Unmarshal(data, &store.cards); nil != err {
		logging.LogErrorf("load cards failed: %s", err)
		return
	}

	p = store.getCardSourcesMsgPackPath()
	if !filelock.IsExist(p) {
		return
	}
	data, err = filelock.ReadFile(p)
	if nil != err {
		logging.LogErrorf("load cardsources failed: %s", err)
	}
	if err = msgpack.Unmarshal(data, &store.cardSources); nil != err {
		logging.LogErrorf("load cardsources failed: %s", err)
		return
	}
	return
}

func (store *FSRSStore) Save() (err error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	saveDir := store.GetSaveDir()
	if !gulu.File.IsDir(saveDir) {
		if err = os.MkdirAll(saveDir, 0755); nil != err {
			return
		}
	}

	p := store.getMsgPackPath()
	data, err := msgpack.Marshal(store.cards)
	if nil != err {
		logging.LogErrorf("save cards failed: %s", err)
		return
	}
	if err = filelock.WriteFile(p, data); nil != err {
		logging.LogErrorf("save cards failed: %s", err)
		return
	}

	p = store.getCardSourcesMsgPackPath()
	data, err = msgpack.Marshal(store.cardSources)
	if nil != err {
		logging.LogErrorf("save cards failed: %s", err)
		return
	}
	if err = filelock.WriteFile(p, data); nil != err {
		logging.LogErrorf("save cards failed: %s", err)
		return
	}
	return
}

func (store *FSRSStore) SaveLog(log *Log) (err error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	saveDir := store.GetSaveDir()
	saveDir = filepath.Join(saveDir, "logs")
	if !gulu.File.IsDir(saveDir) {
		if err = os.MkdirAll(saveDir, 0755); nil != err {
			return
		}
	}

	yyyyMM := time.Now().Format("200601")
	p := filepath.Join(saveDir, yyyyMM+".msgpack")
	logs := []*Log{}
	var data []byte
	if filelock.IsExist(p) {
		data, err = filelock.ReadFile(p)
		if nil != err {
			logging.LogErrorf("load logs failed: %s", err)
			return
		}

		if err = msgpack.Unmarshal(data, &logs); nil != err {
			logging.LogErrorf("unmarshal logs failed: %s", err)
			return
		}
	}
	logs = append(logs, log)

	if data, err = msgpack.Marshal(logs); nil != err {
		logging.LogErrorf("marshal logs failed: %s", err)
		return
	}
	if err = filelock.WriteFile(p, data); nil != err {
		logging.LogErrorf("write logs failed: %s", err)
		return
	}
	return
}

func (store *FSRSStore) AddCardSource(id string, CType CardType, cardIDMap map[string]string) CardSource {
	store.lock.Lock()
	defer store.lock.Unlock()
	cardSource := &BaseCardSource{
		SID:    id,
		CType:  CType,
		CIDMap: cardIDMap}
	for _, cardID := range cardIDMap {
		c := fsrs.NewCard()
		card := &FSRSCard{BaseCard: &BaseCard{CID: cardID}, C: &c}
		store.cards[id] = card
	}
	store.cardSources[id] = cardSource
	return cardSource
}

type FSRSCard struct {
	*BaseCard
	C *fsrs.Card
}

func (card *FSRSCard) Impl() interface{} {
	return card.C
}

func (card *FSRSCard) SetImpl(c interface{}) {
	card.C = c.(*fsrs.Card)
}

func (card *FSRSCard) GetLapses() int {
	return int(card.C.Lapses)
}

func (card *FSRSCard) GetReps() int {
	return int(card.C.Reps)
}

func (card *FSRSCard) GetState() State {
	return State(card.C.State)
}

func (card *FSRSCard) GetLastReview() time.Time {
	return card.C.LastReview
}

func (card *FSRSCard) Clone() Card {
	data, err := gulu.JSON.MarshalJSON(card)
	if nil != err {
		logging.LogErrorf("marshal card failed: %s", err)
		return nil
	}
	ret := &FSRSCard{}
	if err = gulu.JSON.UnmarshalJSON(data, ret); nil != err {
		logging.LogErrorf("unmarshal card failed: %s", err)
		return nil
	}
	return ret
}

func (card *FSRSCard) SetDue(due time.Time) {
	card.C.Due = due
}
