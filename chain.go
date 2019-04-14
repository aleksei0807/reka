package reka

import (
	"time"
)

func (chain *Chain) addMethod(
	method func(interface{}) func(interface{}) (interface{}, *action),
	cb interface{},
) *Chain {
	var node = &node{method: method(cb)}

	chain.Lock()
	if chain.prevNode == nil {
		chain.stream.chains.childs = append(chain.stream.chains.childs, node)
	} else {
		node.prev = chain.prevNode
		chain.prevNode.childs = append(chain.prevNode.childs, node)
	}
	chain.Unlock()

	newChain := Chain{stream: chain.stream, prevNode: node}

	return &newChain
}

func defaultMethod(cb interface{}) func(interface{}) (interface{}, *action) {
	return func(value interface{}) (interface{}, *action) {
		return call(cb, value), &action{}
	}
}

func filterMethod(cb interface{}) func(interface{}) (interface{}, *action) {
	return func(value interface{}) (interface{}, *action) {
		action := &action{}
		if !call(cb, value).(bool) {
			action.actionType = stop
		}

		return value, action
	}
}

func delayMethod(cb interface{}) func(interface{}) (interface{}, *action) {
	return func(value interface{}) (interface{}, *action) {
		data := call(cb, value).(*specificValue)

		return data.value, data.action
	}
}

func (chain *Chain) Log() *Chain {
	logCallback := func(value interface{}) interface{} {
		chain.stream.Logger.WithField("value", value).Debug("Reka log")

		return value
	}

	return chain.addMethod(defaultMethod, logCallback)
}

func (chain *Chain) Logf(msg string, v ...interface{}) *Chain {
	logfCallback := func(value interface{}) interface{} {
		chain.stream.Logger.WithField("value", value).Debugf(msg, v...)

		return value
	}

	return chain.addMethod(defaultMethod, logfCallback)
}

func (chain *Chain) Map(cb interface{}) *Chain {
	return chain.addMethod(defaultMethod, cb)
}

func (chain *Chain) Filter(cb interface{}) *Chain {
	return chain.addMethod(filterMethod, cb)
}

func (chain *Chain) Diff(cb interface{}, seed interface{}) *Chain {
	prevValue := seed

	diffCallback := func(next interface{}) interface{} {
		value := call(cb, prevValue, next)
		prevValue = next

		return value
	}

	return chain.addMethod(defaultMethod, diffCallback)
}

func (chain *Chain) Scan(cb interface{}, seed interface{}) *Chain {
	prevValue := seed

	scanCallback := func(next interface{}) interface{} {
		value := call(cb, prevValue, next)
		prevValue = value

		return value
	}

	return chain.addMethod(defaultMethod, scanCallback)
}

func (chain *Chain) Delay(wait time.Duration) *Chain {
	delayCallback := func(value interface{}) interface{} {
		return &specificValue{
			action: &action{actionType: delay, data: wait},
			value:  value,
		}
	}

	return chain.addMethod(delayMethod, delayCallback)
}
