package reka

import "time"

func (stream *Stream) OnValues() *Chain {
	chain := Chain{stream: stream}

	return &chain
}

func forEach(arr []*node, firstValue interface{}, action *action) {
	if action.actionType != "stop" {
		if action.actionType == "delay" {
			time.Sleep(action.data.(time.Duration))
		}

		value := firstValue

		for _, child := range arr {
			child.RLock()
			v, newAction := child.method(firstValue)
			value = v

			if len(child.childs) != 0 {
				forEach(child.childs, value, newAction)
			}
			child.RUnlock()
		}
	}
}

func (stream *Stream) Push(value interface{}) {
	stream.chains.RLock()
	forEach(stream.chains.childs, value, &action{})
	stream.chains.RUnlock()
}
