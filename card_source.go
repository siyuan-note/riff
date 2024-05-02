package riff

type CardSource interface {

	// CardSource的ID
	SourceID() string

	// 返回类型
	CardType() string

	// 返回卡源 Data
	GetSourceData() string

	// 更新卡源 Data
	UpdateSourceData(NewData string)

	// 返回关联的CardID list
	GetCardIDs() []string

	// 更新关联的CardID list
	SetCardIDs([]string)

	// 返回卡源的 Context
	GetContext() map[string]string

	// 设置卡源的 Context
	SetContext(key string, value string)
}

type BaseCardSource struct {
	SID     string
	CType   string
	CIDs    []string
	Data    string
	Context map[string]string
}

func (source *BaseCardSource) SourceID() string {
	return source.SID
}

func (source *BaseCardSource) CardType() string {
	return source.CType
}

func (source *BaseCardSource) GetSourceData() string {
	return source.Data
}

func (source *BaseCardSource) UpdateSourceData(NewData string) {
	source.Data = NewData
}

func (source *BaseCardSource) GetCardIDs() []string {
	return source.CIDs
}

func (source *BaseCardSource) SetCardIDs(NewCardIDs []string) {
	source.CIDs = NewCardIDs
}

func (source *BaseCardSource) GetContext() map[string]string {
	return source.Context
}

func (source *BaseCardSource) SetContext(key string, value string) {
	source.Context[key] = value
}
