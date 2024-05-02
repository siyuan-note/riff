package riff

type CardSourceStore interface {
	// 添加 CardSource 并添加对应卡片
	AddCardSource(id string, CType CardType, cardIDMap map[string]string) CardSource

	// 更新CardSource CIDMap 对应卡片，使其与 cardIDMap 一致
	// 不存在则新建，已存在则不操作，在 cardIDMap 不存在则删除
	UpdateCardSource(id string, cardIDMap map[string]string) error

	// 通过 CardSourceID 获得 CardSource
	GetCardSourceByID(id string) CardSource

	// 通过 Card获取
	GetCardSourceByCard(card Card) CardSource

	SetCardSource(cardSource CardSource)

	RemoveCardSource(id string)

	Load()

	Save()
}
