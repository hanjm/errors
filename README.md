[![GoDoc](https://godoc.org/github.com/hanjm/errors?status.svg)](https://godoc.org/github.com/hanjm/errors)
[![Go Report Card](https://goreportcard.com/badge/github.com/hanjm/errors)](https://goreportcard.com/report/github.com/hanjm/errors)
[![code-coverage](http://gocover.io/_badge/github.com/hanjm/errors)](http://gocover.io/github.com/hanjm/errors)

1. 按[文件名:行号:函数名:message]格式输出
2. funcName不显示package path, fileName不显示src之前的字符

```go
// 示例代码1 非error常量的情况
// func1 调用func2 func2调用func3
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
			[errors/example.go:25:func2:i=3]
			[errors/example.go:15:func1:[1] 123]
			[errors/error_test.go:22:TestExample:]
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
```