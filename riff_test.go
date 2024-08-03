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
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs"
)

type timeTicker struct {
	totalTicker map[string]time.Time
}

func newTimeTicker() *timeTicker {
	return &timeTicker{
		totalTicker: map[string]time.Time{},
	}
}

func (tt *timeTicker) start(task string) {
	tt.totalTicker[task] = time.Now()
}
func (tt *timeTicker) log(task string) {
	since := time.Since(tt.totalTicker[task])
	fmt.Printf("task %s use time :%s\n", task, since)
}

func checkIDList(IDList []string, data interface{}, dataGetter func(item interface{}) string) (err error) {
	queryIDs := map[string]bool{}
	v := reflect.ValueOf(data)

	for i := 0; i < v.Len(); i++ {
		id := dataGetter(v.Index(i).Interface())
		queryIDs[id] = true
	}
	for _, ID := range IDList {
		if !queryIDs[ID] {
			err = fmt.Errorf("dont have id of %s", ID)
			return
		}
	}
	return
}

func randomSubset[T any](slice []T, size int) ([]T, error) {
	if size > len(slice) {
		return nil, fmt.Errorf("requested subset size is larger than the slice size")
	}

	// Create a copy of the original slice to avoid modifying it

	subset := make([]T, size)
	// copy(subset, slice)
	sourceLen := len(slice)

	for i := range subset {
		subset[i] = slice[rand.Intn(sourceLen)]
	}

	// Return the first `size` elements
	return subset[:size], nil
}

// 检查功能是否正确实现
func TestFunction(t *testing.T) {
	const saveDir = "testdata"
	const RequestRetention = 0.95
	const cardSourceNum = 200
	const blocksNum = 40
	const preSourceCardNum = 3
	const totalCardNum = cardSourceNum * preSourceCardNum

	os.MkdirAll(saveDir, 0755)
	defer os.RemoveAll(saveDir)

	blocksIDs := []string{}
	blockIDsCIDMap := make(map[string][]string, 0)

	for i := 0; i < blocksNum; i++ {
		blocksIDs = append(blocksIDs, newID())
	}

	riff := NewBaseRiff()
	riff.SetParams(AlgoFSRS, fsrs.DefaultParam())
	deck := DefaultBaseDeck()
	csList := []CardSource{}
	cardList := []Card{}
	csIDList := []string{}
	cardIDList := []string{}

	for i := 0; i < cardSourceNum; i++ {
		cs := NewBaseCardSource(deck.DID)

		// 放入blockIDs
		subBlockIDs, _ := randomSubset(blocksIDs, preSourceCardNum)
		cs.BlockIDs = subBlockIDs
		// 准备一个临时数组来储存当前blocksIDs对应的cardSourceID
		for _, blockID := range subBlockIDs {
			blockIDsCIDMap[blockID] = append(blockIDsCIDMap[blockID], cs.CSID)
		}

		csList = append(csList, cs)
		csIDList = append(csIDList, cs.CSID)
		for i := 0; i < preSourceCardNum; i++ {
			card := NewBaseCard(cs)
			card.UseAlgo(AlgoFSRS)
			cardList = append(cardList, card)
			cardIDList = append(cardIDList, card.CID)
		}
	}

	//检查待插入卡片数量
	if len(csIDList) != cardSourceNum {
		t.Errorf("add card source error")
	}

	if len(cardIDList) != totalCardNum {
		t.Errorf("add card  error")
	}

	riff.AddDeck(deck)
	riff.AddCardSource(csList)
	riff.AddCard(cardList)

	queryCsList := []BaseCardSource{}
	queryCardList := []BaseCard{}

	// 确保cardsource完全插入数据库
	riff.Find(&queryCsList)
	if len(queryCsList) != cardSourceNum {
		t.Errorf("add CardSource err num %d:%d ", len(queryCsList), cardSourceNum)
	}
	if err := checkIDList(csIDList, queryCsList, func(item interface{}) string {
		return item.(BaseCardSource).CSID
	}); err != nil {
		t.Errorf("%s", err)
	}

	// 确保card完全插入数据库
	riff.Find(&queryCardList)
	if len(queryCardList) != totalCardNum {
		t.Errorf("add Card err num %d:%d ", len(queryCardList), totalCardNum)
	}

	if err := checkIDList(cardIDList, queryCardList, func(item interface{}) string {
		return item.(BaseCard).CID
	}); err != nil {
		t.Errorf("%s", err)
	}

	reviewInfoList := riff.Due()
	for _, reviewInfo := range reviewInfoList {
		riff.Review(reviewInfo.BaseCard.ID(), Again)
	}

	reviewInfoList = riff.Due()
	for _, reviewInfo := range reviewInfoList {
		riff.Review(reviewInfo.BaseCard.ID(), Easy)
	}

	newreviewCard := riff.Due()

	if len(newreviewCard) != 0 {
		t.Errorf("review error with un review card num :%d\n", len(newreviewCard))
	}

	// 抽取一半的blockIDs检查
	testBlockIDs, _ := randomSubset(blocksIDs, blocksNum/2)
	reviewInfoByblocks := riff.GetCardsByBlockIDs(testBlockIDs)
	existMap := map[string]bool{}
	for _, ri := range reviewInfoByblocks {
		existMap[ri.BaseCardSource.CSID] = true
	}
	for _, blockID := range testBlockIDs {
		mapCSIDs := blockIDsCIDMap[blockID]
		for _, csid := range mapCSIDs {
			if !existMap[csid] {
				t.Errorf("blockID %s cardSource %s no call back", blockID, csid)
			}
		}
	}

	riff.Save(saveDir)

	newRiff := NewBaseRiff()
	newRiff.SetParams(AlgoFSRS, fsrs.DefaultParam())
	newRiff.Load(saveDir)

	// 检查重新加载后 数据是否恢复

	queryCsList = []BaseCardSource{}
	queryCardList = []BaseCard{}

	// 确保cardsource完全插入数据库
	newRiff.Find(&queryCsList)
	if len(queryCsList) != cardSourceNum {
		t.Errorf("add CardSource err num %d:%d ", len(queryCsList), cardSourceNum)
	}
	if err := checkIDList(csIDList, queryCsList, func(item interface{}) string {
		return item.(BaseCardSource).CSID
	}); err != nil {
		t.Errorf("%s", err)
	}

	// 确保card完全插入数据库
	newRiff.Find(&queryCardList)
	if len(queryCardList) != totalCardNum {
		t.Errorf("add Card err num %d:%d ", len(queryCardList), totalCardNum)
	}

	if err := checkIDList(cardIDList, queryCardList, func(item interface{}) string {
		return item.(BaseCard).CID
	}); err != nil {
		t.Errorf("%s", err)
	}

}

// 检查性能参数
// 10000 cardSOurce,
// 30 card
func TestPerformance(t *testing.T) {
	ticker := newTimeTicker()
	ticker.start("TestPerformance")

	const saveDir = "testdata"
	const RequestRetention = 0.95
	const cardSourceNum = 20000
	const blocksNum = 20000
	const preSourceCardNum = 5
	const totalCardNum = cardSourceNum * preSourceCardNum

	os.MkdirAll(saveDir, 0755)
	defer os.RemoveAll(saveDir)

	blocksIDs := []string{}
	blockIDsCIDMap := make(map[string][]string, 0)

	for i := 0; i < blocksNum; i++ {
		blocksIDs = append(blocksIDs, newID())
	}

	riff := NewBaseRiff()
	riff.SetParams(AlgoFSRS, fsrs.DefaultParam())
	deck := DefaultBaseDeck()

	csList := []CardSource{}
	cardList := []Card{}
	csIDList := []string{}
	cardIDList := []string{}
	ticker.start("init card and cardsource")
	for i := 0; i < cardSourceNum; i++ {
		cs := NewBaseCardSource(deck.DID)

		// 放入blockIDs
		subBlockIDs, _ := randomSubset(blocksIDs, preSourceCardNum)
		cs.BlockIDs = subBlockIDs
		// 准备一个临时数组来储存当前blocksIDs对应的cardSourceID
		for _, blockID := range subBlockIDs {
			blockIDsCIDMap[blockID] = append(blockIDsCIDMap[blockID], cs.CSID)
		}

		csList = append(csList, cs)
		csIDList = append(csIDList, cs.CSID)

		if i%10 == 0 {
			time.Sleep(1 * time.Microsecond)
		}
		for i := 0; i < preSourceCardNum; i++ {
			card := NewBaseCard(cs)
			card.UseAlgo(AlgoFSRS)
			cardList = append(cardList, card)
			cardIDList = append(cardIDList, card.CID)

		}
	}
	ticker.log("init card and cardsource")

	ticker.start("adddeck")
	riff.AddDeck(deck)
	ticker.log("adddeck")

	ticker.start("add cardSource")
	riff.AddCardSource(csList)
	ticker.log("add cardSource")

	ticker.start("add card")
	riff.AddCard(cardList)
	ticker.log("add card")

	queryCsList := []BaseCardSource{}
	queryCardList := []BaseCard{}

	// 确保cardsource完全插入数据库
	riff.Find(&queryCsList)
	if len(queryCsList) != cardSourceNum {
		t.Errorf("add CardSource err num %d:%d ", len(queryCsList), cardSourceNum)
	}
	if err := checkIDList(csIDList, queryCsList, func(item interface{}) string {
		return item.(BaseCardSource).CSID
	}); err != nil {
		t.Errorf("%s", err)
	}

	// 确保card完全插入数据库
	riff.Find(&queryCardList)
	if len(queryCardList) != totalCardNum {
		t.Errorf("add Card err num %d:%d ", len(queryCardList), totalCardNum)
	}

	if err := checkIDList(cardIDList, queryCardList, func(item interface{}) string {
		return item.(BaseCard).CID
	}); err != nil {
		t.Errorf("%s", err)
	}

	ticker.start("query due")
	reviewInfoList := riff.Due()
	ticker.log("query due")

	fmt.Printf("due card len :%d \n", len(reviewInfoList))

	ticker.start("review")
	for _, reviewInfo := range reviewInfoList {
		riff.Review(reviewInfo.BaseCard.ID(), Again)
	}
	ticker.log("review")

	ticker.start("query due again")
	reviewInfoList = riff.Due()
	ticker.log("query due again")

	ticker.start("review again")
	for _, reviewInfo := range reviewInfoList {
		riff.Review(reviewInfo.BaseCard.ID(), Easy)
	}
	ticker.log("review again")

	ticker.start("query due after review all")
	newreviewCard := riff.Due()
	ticker.log("query due after review all")

	if len(newreviewCard) != 0 {
		t.Errorf("review error with un review card num :%d\n", len(newreviewCard))
	}

	// 抽取一半的blockIDs检查
	ticker.start("get card by blocks")
	testBlockIDs, _ := randomSubset(blocksIDs, blocksNum)
	reviewInfoByblocks := riff.GetCardsByBlockIDs(testBlockIDs)
	fmt.Printf("reviewInfoByblocks len :%d \n", len(reviewInfoByblocks))
	existMap := map[string]bool{}
	for _, ri := range reviewInfoByblocks {
		existMap[ri.BaseCardSource.CSID] = true
	}
	for _, blockID := range testBlockIDs {
		mapCSIDs := blockIDsCIDMap[blockID]
		for _, csid := range mapCSIDs {
			if !existMap[csid] {
				t.Errorf("blockID %s cardSource %s no call back", blockID, csid)
			}
		}
	}
	ticker.log("get card by blocks")

	ticker.start("save")
	riff.Save(saveDir)
	ticker.log("save")

	ticker.start("load")
	newRiff := NewBaseRiff()
	newRiff.SetParams(AlgoFSRS, fsrs.DefaultParam())
	go newRiff.Load(saveDir)

	ticker.log("load")

	// 检查重新加载后 数据是否恢复

	queryCsList = []BaseCardSource{}
	queryCardList = []BaseCard{}

	// 确保cardsource完全插入数据库
	ticker.start("query total cardSource")
	newRiff.WaitLoad()
	newRiff.Find(&queryCsList)
	ticker.log("query total cardSource")

	if len(queryCsList) != cardSourceNum {
		t.Errorf("add CardSource err num %d:%d ", len(queryCsList), cardSourceNum)
	}
	if err := checkIDList(csIDList, queryCsList, func(item interface{}) string {
		return item.(BaseCardSource).CSID
	}); err != nil {
		t.Errorf("%s", err)
	}

	// 确保card完全插入数据库
	ticker.start("query total card")
	newRiff.Find(&queryCardList)
	ticker.log("query total card")

	if len(queryCardList) != totalCardNum {
		t.Errorf("add Card err num %d:%d ", len(queryCardList), totalCardNum)
	}

	if err := checkIDList(cardIDList, queryCardList, func(item interface{}) string {
		return item.(BaseCard).CID
	}); err != nil {
		t.Errorf("%s", err)
	}
	ticker.log("TestPerformance")
	time.Sleep(5 * time.Second)
}
