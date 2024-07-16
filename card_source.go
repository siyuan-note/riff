package riff

type CardSource interface {
}

type BaseCardSource struct {
	CSID          string
	Hash          string
	BlockIDs      []string
	DID           string
	SourceContext map[string]interface{}
}
