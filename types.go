package reka

type node struct {
	prev   *node
	childs []*node

	method func(interface{}) (interface{}, *action)
}

type tree struct {
	childs []*node
}

type Stream struct {
	chains *tree
}

type Chain struct {
	stream   *Stream
	prevNode *node

	prevValue interface{}
}

type action struct {
	actionType string
	data       interface{}
}

type specificValue struct {
	action *action
	value  interface{}
}
