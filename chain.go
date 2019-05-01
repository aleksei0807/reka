package reka

import (
	"container/list"
	"reflect"
	"sync"
	"sync/atomic"
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
		chain.prevNode.Lock()
		node.prev = chain.prevNode
		chain.prevNode.childs = append(chain.prevNode.childs, node)
		chain.prevNode.Unlock()
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
	list := syncList{
		RWMutex: &sync.RWMutex{},
		List:    list.New(),
	}
	chain.Lock()
	chain.delayValues = append(chain.delayValues, list)
	chain.Unlock()

	var isInit int32

	delayCallback := func(value interface{}) interface{} {
		v := &specificValue{
			action: &action{actionType: delay, data: &delayData{list: list, isInit: atomic.LoadInt32(&isInit), wait: wait}},
			value:  value,
		}

		atomic.AddInt32(&isInit, 1)

		return v
	}

	newNode := chain.addMethod(actionMethod, delayCallback)

	return chain.createNewChain(newNode)
}

func (chain *Chain) Shard(count uint64, shardFunc ...interface{}) []*Chain {
	var iter uint64

	detectShard := func(currIter uint64, value interface{}) uint64 {
		return currIter % count
	}

	if len(shardFunc) > 0 {
		restV := reflect.ValueOf(shardFunc[0])

		if restV.Kind() != reflect.Func {
			panic("shardFunc must be callable, but got " + restV.Kind().String())
		}

		restT := restV.Type()

		if restT.In(0).Kind() != reflect.Uint64 {
			panic("shardFunc must accept currentIteration (uint64) as a first argument")
		}

		if restT.Out(0).Kind() != reflect.Uint64 {
			panic("shardFunc must return shardNumber (uint64)")
		}

		switch restT.NumIn() {
		case 1:
			detectShard = func(currIter uint64, value interface{}) uint64 {
				return call(shardFunc[0], currIter).(uint64)
			}
		case 2:
			detectShard = func(currIter uint64, value interface{}) uint64 {
				return call(shardFunc[0], currIter, value).(uint64)
			}
		default:
			panic("shardFunc must accept 1 or 2 parameters: currentIteration (uint64) and value (optional, interface{})")
		}
	}

	shardCallback := func(value interface{}) interface{} {
		currentShard := detectShard(atomic.LoadUint64(&iter), value)

		atomic.AddUint64(&iter, 1)

		return &specificValue{
			action: &action{actionType: shard, data: currentShard},
			value:  value,
		}
	}

	node := chain.addMethod(actionMethod, shardCallback)

	chains := make([]*Chain, 0)
	for i := uint64(0); i < count; i++ {
		chains = append(chains, &Chain{stream: chain.stream, prevNode: node})
	}

	return chains
}
