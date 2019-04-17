package reka

import (
	"sync"

	"github.com/apex/log"
)

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

	Logger *log.Logger
}

type Chain struct {
	stream *Stream
	sync.Mutex
	prevNode *node
}

type actionType uint8

const (
	undefined actionType = iota
	stop
	delay
	shard
)

type action struct {
	actionType actionType
	data       interface{}
}

type specificValue struct {
	action *action
	value  interface{}
}
