package riff

import (
	"fmt"
	"testing"
)

func TestCardSource(t *testing.T) {
	sid := newID()
	cid := newID()
	cid2 := newID()
	data := "1111"
	context := map[string]string{
		"aaa": "bbb",
	}
	basecardSource := &BaseCardSource{
		SID:   sid,
		CType: builtInCardType,
		CIDMap: map[string]string{
			"card": cid,
		},
		Data: data,
	}
	var cardSource CardSource = basecardSource
	if cardSource.SourceID() != sid {
		t.Fatalf("cardSource id [%s] != [%s]", cardSource.SourceID(), sid)
	}
	if cardSource.CardType() != builtInCardType {
		t.Fatalf("cardSource cardType [%s] != [%s]", cardSource.CardType(), builtInCardType)
	}
	if cardSource.GetSourceData() != data {
		t.Fatalf("cardSource SourceData [%s] != [%s]", cardSource.GetSourceData(), data)
	}
	for key, value := range cardSource.GetContext() {
		if context[key] != value {
			t.Fatalf("cardSource context key [%s] value [%s] != [%s]", key, value, context[key])
		}
	}
	if cardSource.GetCardIDMap()["card"] != cid {
		t.Fatalf("cardSource CardIDMap key card [%s] != [%s]", cardSource.GetCardIDMap()["card"], cid)
	}
	cardSource.SetCardIDMap("card", cid2)
	if cardSource.GetCardIDMap()["card"] != cid2 {
		t.Fatalf("cardSource SetCardIDMap key card [%s] != [%s]", cardSource.GetCardIDMap()["card"], cid)
	}
	if len(cardSource.GetCardIDs()) != 1 {
		t.Fatalf("cardSource GetCardIDs cardIDs len [%s] != 1", fmt.Sprint(len(cardSource.GetCardIDs())))
	}
	if cardSource.GetCardIDs()[0] != cid2 {
		t.Fatalf("cardSource GetCardIDs cardIDs first [%s] != [%s] ", cardSource.GetCardIDs()[0], cid2)
	}

}
