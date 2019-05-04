package reka

import (
	"runtime"
	"sync/atomic"
	"time"
)

func getListLen(list *syncList) int {
	list.RLock()
	listLen := list.Len()
	list.RUnlock()
	return listLen
}

func clearAndPush(list *syncList, value interface{}) {
	list.Lock()
	list.Init()
	list.PushFront(value)
	list.Unlock()
}

func (stream *Stream) OnValues() *Chain {
	chain := Chain{stream: stream}

	return &chain
}

func (stream *Stream) runMethods(child *node, value interface{}) {
	child.RLock()
	v, newAction := child.method(value)

	childs := child.childs
	child.RUnlock()

	if len(childs) > 0 {
		stream.forEach(childs, v, newAction)
	}
}

func (stream *Stream) runMethodsWithList(arr []*node, list *syncList) {
	list.RLock()
	value := list.Front().Value
	list.RUnlock()
	for _, child := range arr {
		stream.runMethods(child, value)
	}
	list.Lock()
	list.Remove(list.Front())
	list.Unlock()
}

func (stream *Stream) delayLoop(arr []*node, wait time.Duration, values *syncList) {
	for {
		if runtime.GOMAXPROCS(0) == 1 {
			runtime.Gosched()
		}
		if getListLen(values) > 0 {
			time.Sleep(wait)
			stream.runMethodsWithList(arr, values)
		}
	}
}

func (stream *Stream) throttleLoop(arr []*node, wait time.Duration, values *syncList) {
	for {
		if runtime.GOMAXPROCS(0) == 1 {
			runtime.Gosched()
		}
		time.Sleep(wait)
		if getListLen(values) > 0 {
			stream.runMethodsWithList(arr, values)
		}
	}
}

func (stream *Stream) debounceLoop(arr []*node, data *debounceData) {
	var prevExpTime time.Time
	for {
		if runtime.GOMAXPROCS(0) == 1 {
			runtime.Gosched()
		}

		expTime := data.expTime.Load().(time.Time)
		if prevExpTime == expTime {
			continue
		}
		isExp := time.Now().After(expTime)

		if isExp {
			prevExpTime = expTime
			if getListLen(data.list) > 0 {
				stream.runMethodsWithList(arr, data.list)
			}
		} else {
			time.Sleep(expTime.Sub(time.Now()))
		}
	}
}

func (stream *Stream) forEach(arr []*node, value interface{}, action *action) {
	if action.actionType != actStop {
		switch action.actionType {
		case actShard:
			currentShard := action.data.(uint64)

			if len(arr) > int(currentShard) {
				child := arr[currentShard]
				stream.runMethods(child, value)
			}

		case actDelay:
			actionData := action.data.(*waitData)
			if atomic.LoadInt32(&actionData.isInit) == 0 {
				go stream.delayLoop(arr, actionData.wait, actionData.list)
				atomic.SwapInt32(&actionData.isInit, 1)
			}
			actionData.list.Lock()
			actionData.list.PushBack(value)
			actionData.list.Unlock()

		case actThrottle:
			actionData := action.data.(*waitData)
			if atomic.LoadInt32(&actionData.isInit) == 0 {
				go stream.throttleLoop(arr, actionData.wait, actionData.list)
				atomic.SwapInt32(&actionData.isInit, 1)
			}
			clearAndPush(actionData.list, value)

		case actDebounce:
			actionData := action.data.(*debounceData)
			actionData.expTime.Store(time.Now().Add(actionData.wait))
			if atomic.LoadInt32(&actionData.isInit) == 0 {
				go stream.debounceLoop(arr, actionData)
				atomic.SwapInt32(&actionData.isInit, 1)
			}
			clearAndPush(actionData.list, value)

		default:
			for _, child := range arr {
				stream.runMethods(child, value)
			}
		}
	}
}

func (stream *Stream) Push(value interface{}) {
	stream.chains.RLock()
	stream.forEach(stream.chains.childs, value, &action{})
	stream.chains.RUnlock()
}
