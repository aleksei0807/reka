package reka

import (
	"reflect"
)

func call(fn interface{}, args ...interface{}) interface{} {
	fnv := reflect.ValueOf(fn)
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}
	outs := fnv.Call(in)
	if len(outs) == 0 || !outs[0].CanInterface() {
		return nil
	}
	return outs[0].Interface()
}
