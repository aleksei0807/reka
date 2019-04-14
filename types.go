package reka

import "sync"

type node struct {
	prev *node
	sync.RWMutex
	childs []*node

	method func(interface{}) (interface{}, *action)
}

type tree struct {
	sync.RWMutex
	childs []*node
}

type Stream struct {
	chains *tree
}

type Chain struct {
	stream *Stream
	sync.Mutex
	prevNode *node
}

type action struct {
	actionType string
	data       interface{}
}

type specificValue struct {
	action *action
	value  interface{}
}
