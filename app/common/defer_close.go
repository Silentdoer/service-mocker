package common

import (
	"fmt"
	"io"
	"reflect"
)

func DeferClose(p io.Closer) {
	_ = p.Close()
}

func DeferCloseCheckError(p io.Closer) {
	err := p.Close()
	if err != nil {
		panic(fmt.Errorf("关闭%s类型对象时发生错误%s", reflect.TypeOf(p), err.Error()))
	}
}
