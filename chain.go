package reka

import (
	"time"

	"github.com/apex/log"
)

func (chain *Chain) addMethod(
	method func(func(interface{}) interface{}) func(interface{}) (interface{}, *action),
	cb func(interface{}) interface{},
) *Chain {
	var node = &node{method: method(cb)}

	if chain.prevNode == nil {
		chain.stream.chains.childs = append(chain.stream.chains.childs, node)
	} else {
		node.prev = chain.prevNode
		chain.prevNode.childs = append(chain.prevNode.childs, node)
	}

	newChain := Chain{stream: chain.stream, prevNode: node}

	return &newChain
}

func defaultMethod(cb func(interface{}) interface{}) func(interface{}) (interface{}, *action) {
	return func(value interface{}) (interface{}, *action) {
		return cb(value), &action{}
	}
}

func filterMethod(cb func(interface{}) interface{}) func(interface{}) (interface{}, *action) {
	return func(value interface{}) (interface{}, *action) {
		action := &action{}
		if !cb(value).(bool) {
			action.actionType = "stop"
		}

		return value, action
	}
}

func delayMethod(cb func(interface{}) interface{}) func(interface{}) (interface{}, *action) {
	return func(value interface{}) (interface{}, *action) {
		data := cb(value).(*specificValue)

		return data.value, data.action
	}
}

func (chain *Chain) Log() *Chain {
	logCallback := func(value interface{}) interface{} {
		log.WithField("value", value).Info("Reka log")

		return value
	}

	return chain.addMethod(defaultMethod, logCallback)
}

func (chain *Chain) Logf(msg string, v ...interface{}) *Chain {
	logfCallback := func(value interface{}) interface{} {
		log.WithField("value", value).Infof(msg, v...)

		return value
	}

	return chain.addMethod(defaultMethod, logfCallback)
}

func (chain *Chain) Map(cb func(interface{}) interface{}) *Chain {
	return chain.addMethod(defaultMethod, cb)
}

func (chain *Chain) Filter(cb func(interface{}) interface{}) *Chain {
	return chain.addMethod(filterMethod, cb)
}

func (chain *Chain) Diff(cb func(interface{}, interface{}) interface{}, seed interface{}) *Chain {
	prevValue := seed

	diffCallback := func(next interface{}) interface{} {
		value := cb(prevValue, next)
		prevValue = next

		return value
	}

	return chain.addMethod(defaultMethod, diffCallback)
}

func (chain *Chain) Scan(cb func(interface{}, interface{}) interface{}, seed interface{}) *Chain {
	prevValue := seed

	scanCallback := func(next interface{}) interface{} {
		value := cb(prevValue, next)
		prevValue = value

		return value
	}

	return chain.addMethod(defaultMethod, scanCallback)
}

func (chain *Chain) Delay(wait time.Duration) *Chain {
	delayCallback := func(value interface{}) interface{} {
		return &specificValue{
			action: &action{actionType: "delay", data: wait},
			value:  value,
		}
	}

	return chain.addMethod(delayMethod, delayCallback)
}
