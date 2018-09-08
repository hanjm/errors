[![GoDoc](https://godoc.org/github.com/hanjm/errors?status.svg)](https://godoc.org/github.com/hanjm/errors)
[![Go Report Card](https://goreportcard.com/badge/github.com/hanjm/errors)](https://goreportcard.com/report/github.com/hanjm/errors)
[![code-coverage](http://gocover.io/_badge/github.com/hanjm/errors)](http://gocover.io/github.com/hanjm/errors)

errors包装error对象, 添加调用栈及附加自定义信息, 以便于从日志中快速定位问题.
特点:
1. 相比 github.com/pkg/errors github.com/juju/errors 开销小, 只在最早出错的地方会调用runtime.Callers, 只调用一次.
2. 对齐调用栈和附加上下文信息, 按[ 文件名:行号 函数名:message ]分行格式化输出, funcName不显示package path, fileName不显示src之前的字符. Goland等IDE可以在控制台上点击错误信息定位源码.
3. 精简调用栈输出, 默认忽略标准库和 github.com 下的包的调用栈, 也可以在 errors.go:85 处自定义一些filter, 忽略一些固定的框架的调用栈信息.

文章:
[Go如何优雅地错误处理(Error Handling and Go 1)](https://imhanjm.com/2018/07/08/go%E5%A6%82%E4%BD%95%E4%BC%98%E9%9B%85%E5%9C%B0%E9%94%99%E8%AF%AF%E5%A4%84%E7%90%86(error%20handling%20and%20go%201)/)

doc:
```go
func Errorf(err error, format string, a ...interface{}) error {
 //...
}
用于包装上一步New/Errorf返回的error/*Err, 添加错误注释, 如 比"xx function error"更直接的错误说明、调用函数的参数值等
 			如果参数error类型不为*Err(error常量或自定义error类型或nil), 用于最早出错的地方, 会收集调用栈
 			如果参数error类型为*Err, 不会收集调用栈.
上层调用方可以通过GetInnerMost得到最里层被包装过的error常量

```go
// 示例代码1 非error常量的情况
// ExampleFunc1 调用func2 func2调用func3
// 在func3使用errors.Errorf时第一个参数传nil收集最完整的调用栈,
// 其他地方用errors.Errorf时第一个参数传上一步返回的error, 最后打log
func func1() {
	requestID := "1"
	err := func2()
	if err != nil {
		err = Errorf(err, "[%s] 123", requestID)
		log.Print(err)
		// log ouuput:
		/*
			2017/09/02 18:55:35 [errors/example.go:33:func3:unexpected param]
			[ errors/example.go:25 func2:i=3]
			[ errors/example.go:15 func1:[1] 123]
			[ errors/error_test.go:22 TestExample:]
		*/
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
func func11() {
	requestID := "11"
	err := func22()
	if err != nil {
		err = Errorf(err, "[%s] 123", requestID)
		log.Print(err)
		// log output:
		/*
		2017/09/02 18:55:35 [errors/example.go:67:func33:unexpected param err:someUnexpected]
		[ errors/example.go:56 func22:i=3]
		[ errors/example.go:46 func11:[11] 123]
		[ errors/error_test.go:26 TestExample2:]
		*/
	}
	return
}

func func22() (err error) {
	i := 3
	err = func33(i)
	if err != nil {
		if err2, ok := err.(*Err); ok && err2.Inner() == errSomeUnexpected {
			fmt.Printf("==\n识别到上一步的std error:%s\n==\n", err2.Inner())
		}
		return Errorf(err, "i=%d", i)
	}
	return
}

func func33(i int) (err error) {
	return Errorf(errSomeUnexpected, "unexpected param")
}
```
