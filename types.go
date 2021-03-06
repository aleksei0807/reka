package reka

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"

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

	sync.RWMutex
	prevNode *node
}

type actionType uint8

const (
	actUndefined actionType = iota
	actStop
	actDelay
	actThrottle
	actDebounce
	actShard
)

type action struct {
	actionType actionType
	data       interface{}
}

type specificValue struct {
	action *action
	value  interface{}
}

type syncList struct {
	*sync.RWMutex
	*list.List
}

type waitData struct {
	wait   time.Duration
	isInit int32
	list   *syncList
}

type debounceData struct {
	wait    time.Duration
	isInit  int32
	list    *syncList
	expTime *atomic.Value
}
