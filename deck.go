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
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/88250/gulu"
	"github.com/siyuan-note/logging"
	"github.com/vmihailenco/msgpack/v5"
)

// Deck 描述了一套闪卡包。
type Deck struct {
	ID        string            // ID
	Name      string            // 名称
	Algo      Algo              // 间隔重复算法
	Desc      string            // 描述
	Created   int64             // 创建时间
	Updated   int64             // 更新时间
	BlockCard map[string]string // 内容块 ID 到闪卡 ID 的映射

	store Store // 底层存储
	lock  *sync.Mutex
}

// LoadDeck 从文件夹 saveDir 路径上加载 id 闪卡包。
func LoadDeck(saveDir, id string) (deck *Deck, err error) {
	deck = &Deck{
		ID:        id,
		Name:      id,
		Algo:      AlgoFSRS,
		Created:   time.Now().UnixMilli(),
		BlockCard: map[string]string{},
		lock:      &sync.Mutex{},
	}

	dataPath := getDeckMsgpackPath(saveDir, id)
	if gulu.File.IsExist(dataPath) {
		var data []byte
		data, err = os.ReadFile(dataPath)
		if nil != err {
			logging.LogErrorf("load deck [%s] failed: %s", deck.Name, err)
			return
		}

		err = msgpack.Unmarshal(data, deck)
		if nil != err {
			logging.LogErrorf("load deck [%s] failed: %s", deck.Name, err)
			return
		}
	}

	var store Store
	switch deck.Algo {
	case AlgoFSRS:
		store = NewFSRSStore(deck.ID, saveDir)
		err = store.Load()
	default:
		err = errors.New("not supported yet")
		return
	}
	if nil != err {
		return
	}
	deck.store = store
	return
}

// AddCard 新建一张闪卡。
func (deck *Deck) AddCard(cardID, blockID string) {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	if "" != deck.BlockCard[blockID] {
		return
	}

	deck.store.AddCard(cardID, blockID)
	deck.BlockCard[blockID] = cardID
}

// RemoveCard 删除一张闪卡。
func (deck *Deck) RemoveCard(blockID string) {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	cardID := deck.BlockCard[blockID]
	delete(deck.BlockCard, blockID)
	if "" == cardID {
		return
	}

	deck.store.RemoveCard(cardID)
}

// GetCard 根据内容块 ID 获取对应的闪卡。
func (deck *Deck) GetCard(blockID string) Card {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	cardID := deck.BlockCard[blockID]
	if "" == cardID {
		return nil
	}
	return deck.store.GetCard(cardID)
}

// Save 保存闪卡包。
func (deck *Deck) Save() (err error) {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	err = deck.store.Save()
	if nil != err {
		logging.LogErrorf("save deck [%s] failed: %s", deck.Name, err)
		return
	}

	saveDir := deck.store.GetSaveDir()
	if !gulu.File.IsDir(saveDir) {
		if err = os.MkdirAll(saveDir, 0755); nil != err {
			return
		}
	}

	dataPath := getDeckMsgpackPath(saveDir, deck.ID)
	data, err := msgpack.Marshal(deck)
	if nil != err {
		logging.LogErrorf("save deck failed: %s", err)
		return
	}
	if err = gulu.File.WriteFileSafer(dataPath, data, 0644); nil != err {
		logging.LogErrorf("save deck failed: %s", err)
		return
	}
	return
}

// Review 复习一张闪卡，rating 为复习评分结果。
func (deck *Deck) Review(blockID string, rating Rating) {
	deck.lock.Lock()
	defer deck.lock.Unlock()

	cardID := deck.BlockCard[blockID]
	if "" == cardID {
		return
	}
	deck.store.Review(cardID, rating)
}

// Dues 返回所有到期的闪卡。
func (deck *Deck) Dues() (ret []Card) {
	deck.lock.Lock()
	defer deck.lock.Unlock()
	return deck.store.Dues()
}

func getDeckMsgpackPath(saveDir, id string) string {
	return filepath.Join(saveDir, id+".deck")
}
