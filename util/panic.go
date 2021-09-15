package util

import (
	"regexp"
	"runtime/debug"
)

var re_panicHead = regexp.MustCompile(`^([\s\S]*?)runtime/panic.go(.*)\n`)
var re_panicHeadStack = regexp.MustCompile(`^([\s\S]*?)runtime/debug/stack.go(.*)\n`)
var re_panicRecover = regexp.MustCompile(`(.*)\n(.*)util/recover.go(.*)\n`)

func Recover(f func(), panicFunc func(r interface{}, stack string)) {
	defer func() {
		r := recover()
		if r != nil {
			stack := debug.Stack()
			stack = re_panicHeadStack.ReplaceAll(stack, nil)
			stack = re_panicHead.ReplaceAll(stack, nil)
			stack = re_panicRecover.ReplaceAll(stack, nil)
			panicFunc(r, string(stack))
		}
	}()
	f()
}
