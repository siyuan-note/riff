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
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/open-spaced-repetition/go-fsrs"
)

func checkIDList(IDList []string, data interface{}, dataGetter func(item interface{}) string) (err error) {
	queryIDs := map[string]bool{}
	v := reflect.ValueOf(data)

	for i := 0; i < v.Len(); i++ {
		id := dataGetter(v.Index(i).Interface())
		queryIDs[id] = true
	}
	for _, ID := range IDList {
		if !queryIDs[ID] {
			err = errors.New(fmt.Sprintf("dont have id of %s", ID))
			return
		}
	}
	return
}

func TestPerformance(t *testing.T) {
	const saveDir = "testdata"
	const RequestRetention = 0.95
	const cardSourceNum = 1
	const preSourceCardNum = 3
	const totalCardNum = cardSourceNum * preSourceCardNum
	os.MkdirAll(saveDir, 0755)
	defer os.RemoveAll(saveDir)
	riff := NewBaseRiff()
	riff.SetParams(AlgoFSRS, fsrs.DefaultParam())
	deck := DefaultBaseDeck()
	csList := []CardSource{}
	cardList := []Card{}
	csIDList := []string{}
	cardIDList := []string{}
	for i := 0; i < cardSourceNum; i++ {
		cs := NewBaseCardSource(deck.DID)
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
	riff.(*BaseRiff).Db.Find(&queryCsList)
	if len(queryCsList) != cardSourceNum {
		t.Errorf("add CardSource err num %d:%d ", len(queryCsList), cardSourceNum)
	}
	if err := checkIDList(csIDList, queryCsList, func(item interface{}) string {
		return item.(BaseCardSource).CSID
	}); err != nil {
		t.Errorf("%s", err)
	}

	// 确保card完全插入数据库
	riff.(*BaseRiff).Db.Find(&queryCardList)
	if len(queryCardList) != totalCardNum {
		t.Errorf("add Card err num %d:%d ", len(queryCardList), totalCardNum)
	}

	if err := checkIDList(cardIDList, queryCardList, func(item interface{}) string {
		return item.(BaseCard).CID
	}); err != nil {
		t.Errorf("%s", err)
	}

	reviewCard := riff.Due()
	for _, card := range reviewCard {
		riff.Review(card, Easy, RequestRetention)
	}
	newreviewCard := riff.Due()
	_ = len(newreviewCard)
	riff.Save(saveDir)

	newRiff := NewBaseRiff()
	newRiff.SetParams(AlgoFSRS, fsrs.DefaultParam())
	newRiff.Load(saveDir)

	// 检查重新加载后 数据是否恢复

	queryCsList = []BaseCardSource{}
	queryCardList = []BaseCard{}

	// 确保cardsource完全插入数据库
	newRiff.(*BaseRiff).Db.Find(&queryCsList)
	if len(queryCsList) != cardSourceNum {
		t.Errorf("add CardSource err num %d:%d ", len(queryCsList), cardSourceNum)
	}
	if err := checkIDList(csIDList, queryCsList, func(item interface{}) string {
		return item.(BaseCardSource).CSID
	}); err != nil {
		t.Errorf("%s", err)
	}

	// 确保card完全插入数据库
	newRiff.(*BaseRiff).Db.Find(&queryCardList)
	if len(queryCardList) != totalCardNum {
		t.Errorf("add Card err num %d:%d ", len(queryCardList), totalCardNum)
	}

	if err := checkIDList(cardIDList, queryCardList, func(item interface{}) string {
		return item.(BaseCard).CID
	}); err != nil {
		t.Errorf("%s", err)
	}

}
