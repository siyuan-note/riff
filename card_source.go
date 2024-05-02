package riff

type CardSource interface {

	// CardSource的ID
	SourceID() string

	// 返回类型
	CardType() CardType

	// 返回卡源 Data
	GetSourceData() string

	// 更新卡源 Data
	SetSourceData(NewData string)

	// 返回关联的CardID list
	GetCardIDs() []string

	RemoveCardID(CardID string)

	GetCardIDMap() map[string]string

	// 更新关联的CardID list
	SetCardIDMap(key string, CardID string)

	// 返回卡源的 Context
	GetContext() map[string]string

	// 设置卡源的 Context
	SetContext(key string, value string)
}

type CardType string

// BaseCardSource 描述了卡源的基础实现
type BaseCardSource struct {
	SID     string
	CType   CardType
	CIDMap  map[string]string
	Data    string
	Context map[string]string
}

const (
	builtInCardType     CardType = "siyuan_busic_card"
	builtInCardIDMapKey string   = "basic_card"
	builtInContext      string   = "blockIDs"
)

func (source *BaseCardSource) SourceID() string {
	return source.SID
}

func (source *BaseCardSource) CardType() CardType {
	return source.CType
}

func (source *BaseCardSource) GetSourceData() string {
	return source.Data
}

func (source *BaseCardSource) SetSourceData(NewData string) {
	source.Data = NewData
}

func (source *BaseCardSource) GetCardIDs() []string {
	var CIDs []string
	for _, CID := range source.CIDMap {
		CIDs = append(CIDs, CID)
	}
	return CIDs
}

func (source *BaseCardSource) RemoveCardID(CardID string) {

	for key, CID := range source.CIDMap {
		if CID == CardID {
			delete(source.CIDMap, key)
		}
	}
	return
}

func (source *BaseCardSource) GetCardIDMap() map[string]string {
	return source.CIDMap
}

func (source *BaseCardSource) SetCardIDMap(key string, CardID string) {
	source.CIDMap[key] = CardID
}

func (source *BaseCardSource) GetContext() map[string]string {
	return source.Context
}

func (source *BaseCardSource) SetContext(key string, value string) {
	source.Context[key] = value
}
