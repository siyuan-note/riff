package riff

type CardSource interface {

	// CardSource的ID
	SourceID() string

	// 返回类型
	CardType() string

	// 返回数据
	SourceData() map[string]string

	// 更新数据
	UpdateData(key string, value string)

	// 返回关联的CardID list
	CardIDs() []string
}

type BaseCardSource struct {
	SID   string
	CType string
	CIDs  []string
	Data  map[string]string
}

func (source *BaseCardSource) SourceID() string {
	return source.SID
}

func (source *BaseCardSource) CardType() string {
	return source.CType
}

func (source *BaseCardSource) SourceData() map[string]string {
	return source.Data
}

func (source *BaseCardSource) UpdateData(key string, value string) {
	source.Data[key] = value
}
func (source *BaseCardSource) CardIDs() []string {
	return source.CIDs
}
