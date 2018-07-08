package errors

import (
	"errors"
	"log"
)

// 示例代码1 非error常量的情况
// ExampleFunc1 调用func2 func2调用func3
// 在func3使用errors.Errorf时第一个参数传nil收集最完整的调用栈,
// 其他地方用errors.Errorf时第一个参数传上一步返回的error, 最后打log
func ExampleFunc1() {
	requestID := "1"
	err := func2()
	if err != nil {
		err = Errorf(err, "[%s] 123", requestID)
		log.Print(err)
	}
	return
}

func func2() (err error) {
	i := 3
	err = func3(i)
	if err != nil {
		return Errorf(err, "i=%d", i)
	}
	return
}

func func3(i int) (err error) {
	return Errorf(nil, "unexpected param")
}

var (
	errSomeUnexpected = errors.New("someUnexpected")
)

// 示例代码2  error常量的情况
// ExampleFunc11 调用func22 func22调用func33
// 在func33使用errors.Errorf包装error常量,收集最完整的调用栈, 最后打log
// 调用func22处可以调GetInnerMost()方法来取到最里层被包装的error常量
func ExampleFunc11() {
	requestID := "11"
	err := func22()
	if err != nil {
		err = Errorf(err, "[%s] 123", requestID)
		log.Print(err)
	}
	if GetInnerMost(err) == errSomeUnexpected {
		log.Printf("识别到最里面的std error:%s\n", errSomeUnexpected)
	}
	return
}

func func22() (err error) {
	i := 3
	err = func33(i)
	if err != nil {
		if GetInnerMost(err) == errSomeUnexpected {
			log.Printf("识别到上一步的std error:%s\n", errSomeUnexpected)
		}
		return Errorf(err, "i=%d", i)
	}
	return
}

func func33(i int) (err error) {
	return Errorf(errSomeUnexpected, "unexpected param")
}