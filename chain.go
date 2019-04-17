package reka

import (
	"time"
)

func (chain *Chain) addMethod(
	method func(interface{}) func(interface{}) (interface{}, *action),
	cb interface{},
) *node {
	var node = &node{method: method(cb)}

	chain.Lock()
	if chain.prevNode == nil {
		chain.stream.chains.childs = append(chain.stream.chains.childs, node)
	} else {
		node.prev = chain.prevNode
		chain.prevNode.childs = append(chain.prevNode.childs, node)
	}
	chain.Unlock()

	return node
}

func (chain *Chain) createNewChain(node *node) *Chain {
	return &Chain{stream: chain.stream, prevNode: node}
}

func defaultMethod(cb interface{}) func(interface{}) (interface{}, *action) {
	return func(value interface{}) (interface{}, *action) {
		return call(cb, value), &action{}
	}
}

func actionMethod(cb interface{}) func(interface{}) (interface{}, *action) {
	return func(value interface{}) (interface{}, *action) {
		data := call(cb, value).(*specificValue)

		return data.value, data.action
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

func (chain *Chain) Log() *Chain {
	logCallback := func(value interface{}) interface{} {
		chain.stream.Logger.WithField("value", value).Debug("Reka log")

		return value
	}

	return chain.createNewChain(chain.addMethod(defaultMethod, logCallback))
}

func (chain *Chain) Logf(msg string, v ...interface{}) *Chain {
	logfCallback := func(value interface{}) interface{} {
		chain.stream.Logger.WithField("value", value).Debugf(msg, v...)

		return value
	}

	return chain.createNewChain(chain.addMethod(defaultMethod, logfCallback))
}

func (chain *Chain) Map(cb interface{}) *Chain {
	return chain.createNewChain(chain.addMethod(defaultMethod, cb))
}

func (chain *Chain) Filter(cb interface{}) *Chain {
	return chain.createNewChain(chain.addMethod(filterMethod, cb))
}

func (chain *Chain) Diff(cb interface{}, seed interface{}) *Chain {
	prevValue := seed

	diffCallback := func(next interface{}) interface{} {
		value := call(cb, prevValue, next)
		prevValue = next

		return value
	}

	return chain.createNewChain(chain.addMethod(defaultMethod, diffCallback))
}

func (chain *Chain) Scan(cb interface{}, seed interface{}) *Chain {
	prevValue := seed

	scanCallback := func(next interface{}) interface{} {
		value := call(cb, prevValue, next)
		prevValue = value

		return value
	}

	return chain.createNewChain(chain.addMethod(defaultMethod, scanCallback))
}

func (chain *Chain) Delay(wait time.Duration) *Chain {
	delayCallback := func(value interface{}) interface{} {
		return &specificValue{
			action: &action{actionType: delay, data: wait},
			value:  value,
		}
	}

	return chain.createNewChain(chain.addMethod(actionMethod, delayCallback))
}

func (chain *Chain) Shard(count uint64, rest ...interface{}) []*Chain {
	var iter uint64

	shardCallback := func(value interface{}) interface{} {
		var currentShard uint64
		if len(rest) != 0 {
			currentShard = call(rest[0], iter, value).(uint64)
		} else {
			currentShard = iter % count
		}

		iter++

		return &specificValue{
			action: &action{actionType: shard, data: currentShard},
			value:  value,
		}
	}

	node := chain.addMethod(actionMethod, shardCallback)

	chains := make([]*Chain, 0)
	intCount := int(count)
	for i := 0; i < intCount; i++ {
		chains = append(chains, &Chain{stream: chain.stream, prevNode: node})
	}

	return chains
}
