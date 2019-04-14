package reka

import "fmt"

func ByType(expected interface{}) func(x interface{}) bool {
	exp := fmt.Sprintf("%T", expected)

	return func(x interface{}) bool {
		return fmt.Sprintf("%T", x) == exp
	}
}
