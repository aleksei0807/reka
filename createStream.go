package reka

func CreateStream() *Stream {
	return &Stream{chains: &tree{}}
}
