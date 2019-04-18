package reka

import (
	"runtime"
	"time"
)

func (stream *Stream) OnValues() *Chain {
	chain := Chain{stream: stream}

	return &chain
}

func (stream *Stream) delayLoop(arr []*node, wait time.Duration, values syncList) {
	for {
		if runtime.GOMAXPROCS(0) == 1 {
			runtime.Gosched()
		}
		values.RLock()
		listLen := values.Len()
		values.RUnlock()
		if listLen > 0 {
			time.Sleep(wait)
			values.RLock()
			x := values.Front().Value
			values.RUnlock()
			for _, child := range arr {
				child.RLock()
				v, newAction := child.method(x)

				childs := child.childs
				child.RUnlock()

				if len(childs) > 0 {
					stream.forEach(childs, v, newAction)
				}
			}
			values.Lock()
			values.Remove(values.Front())
			values.Unlock()
		}
	}
}

func (stream *Stream) forEach(arr []*node, value interface{}, action *action) {
	if action.actionType != stop {
		switch action.actionType {
		case shard:
			currentShard := action.data.(uint64)

			if len(arr) > int(currentShard) {
				child := arr[currentShard]
				child.RLock()
				v, newAction := child.method(value)

				if len(child.childs) > 0 {
					stream.forEach(child.childs, v, newAction)
				}
				child.RUnlock()
			}

		case delay:
			actionData := action.data.(*delayData)
			if !actionData.isInit {
				go stream.delayLoop(arr, actionData.wait, actionData.list)
			}
			actionData.list.Lock()
			actionData.list.PushBack(value)
			actionData.list.Unlock()

		default:
			for _, child := range arr {
				child.RLock()
				v, newAction := child.method(value)

				if len(child.childs) > 0 {
					stream.forEach(child.childs, v, newAction)
				}
				child.RUnlock()
			}
		}
	}
}

func (stream *Stream) Push(value interface{}) {
	stream.chains.RLock()
	stream.forEach(stream.chains.childs, value, &action{})
	stream.chains.RUnlock()
}
