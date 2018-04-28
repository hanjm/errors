package errors

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"strconv"
)

var goRoot = runtime.GOROOT()

var formatPartHead = []byte{'\n', '\t', '['}

const (
	formatPartColon = ':'
	formatPartTail  = ']'
)

type Err struct {
	message  string
	stdError error
	// prevErr 指向上一个Err
	prevErr     *Err
	stack       []uintptr
	once        sync.Once
	fullMessage string
}

type stackFrame struct {
	funcName string
	file     string
	line     int
	message  string
}

// Error
func (e *Err) Error() string {
	e.once.Do(func() {
		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		var (
			messages []string
			stack    []uintptr
		)
		for prev := e; prev != nil; prev = prev.prevErr {
			stack = prev.stack
			if prev.stdError != nil {
				messages = append(messages, fmt.Sprintf("%s err:%s", prev.message, prev.stdError.Error()))
			} else {
				messages = append(messages, prev.message)
			}
		}
		sf := stackFrame{}
		for i, v := range stack {
			if j := len(messages) - 1 - i; j > -1 {
				sf.message = messages[j]
			} else {
				sf.message = ""
			}
			funcForPc := runtime.FuncForPC(v)
			if funcForPc == nil {
				sf.file = "???"
				sf.line = 0
				sf.funcName = "???"
				//fmt.Fprintf(buf, "\n\t[%s:%d:%s:%s]", sf.file, sf.line, sf.funcName, sf.message)
				buf.Write(formatPartHead)
				buf.WriteString(sf.file)
				buf.WriteByte(formatPartColon)
				buf.WriteString(strconv.Itoa(sf.line))
				buf.WriteByte(formatPartColon)
				buf.WriteString(sf.funcName)
				buf.WriteByte(formatPartColon)
				buf.WriteString(sf.message)
				buf.WriteByte(formatPartTail)
				continue
			}
			sf.file, sf.line = funcForPc.FileLine(v - 1)
			// 忽略GOROOT下代码的调用栈 如/usr/local/Cellar/go/1.8.3/libexec/src/runtime/asm_amd64.s:2198:runtime.goexit:
			if strings.HasPrefix(sf.file, goRoot) {
				continue
			}
			const src = "/src/"
			if idx := strings.Index(sf.file, src); idx > 0 {
				sf.file = sf.file[idx+len(src):]
			}
			//if strings.HasPrefix(sf.file, "github.com") {
			//	continue
			//}
			// 处理函数名
			sf.funcName = funcForPc.Name()
			// 保证闭包函数名也能正确显示 如TestErrorf.func1:
			idx := strings.LastIndexByte(sf.funcName, '/')
			if idx != -1 {
				sf.funcName = sf.funcName[idx:]
				idx = strings.IndexByte(sf.funcName, '.')
				if idx != -1 {
					sf.funcName = strings.TrimPrefix(sf.funcName[idx:], ".")
				}
			}
			//fmt.Fprintf(buf, "\n\t[%s:%d:%s:%s]", sf.file, sf.line, sf.funcName, sf.message)
			buf.Write(formatPartHead)
			buf.WriteString(sf.file)
			buf.WriteByte(formatPartColon)
			buf.WriteString(strconv.Itoa(sf.line))
			buf.WriteByte(formatPartColon)
			buf.WriteString(sf.funcName)
			buf.WriteByte(formatPartColon)
			buf.WriteString(sf.message)
			buf.WriteByte(formatPartTail)
		}
		e.fullMessage = buf.String()
	})
	return e.fullMessage
}

// Prev 返回上一步传入Errorf的*Err
func (e *Err) Prev() *Err {
	return e.prevErr
}

// Inner 返回上一步传入Errorf的error, 用于判断error的值和已定义的error类型是否相等
func (e *Err) Inner() error {
	return e.stdError
}

// New 是标准库中的New函数 只能用于定义error常量使用
func New(msg string) error {
	return errors.New(msg)
}

// Errorf
//	用于包装上一步New/Errorf返回的error/*Err, 添加错误注释, 如 比"xx function error"更直接的错误说明、调用函数的参数值等
// 			如果参数error类型不为*Err(error常量或自定义error类型或nil), 用于最早出错的地方, 会收集调用栈
// 			如果参数error类型为*Err, 不会收集调用栈.
//  上层调用方可以通过GetInner/GetInnerMost得到里层/最里层被包装过的error常量
func Errorf(err error, format string, a ...interface{}) error {
	var msg string
	if len(a) == 0 {
		msg = format
	} else {
		msg = fmt.Sprintf(format, a...)
	}
	if err, ok := err.(*Err); ok {
		return &Err{
			message: msg,
			prevErr: err,
		}
	}
	newErr := new_(msg)
	newErr.stdError = err
	return newErr
}

func new_(msg string) *Err {
	pc := make([]uintptr, 200)
	length := runtime.Callers(3, pc)
	return &Err{
		message: msg,
		stack:   pc[:length],
	}
}

// GetInner
// 返回上一步传入Errorf的error常量
// 用于简化写这么一大长串类型断言 if err2, ok := err.(*Err); ok && err2.Inner() == errSomeUnexpected {}
func GetInner(err error) (error) {
	if err2, ok := err.(*Err); ok {
		return err2.Inner()
	}
	return err
}

// GetInnerMost
// 返回最早的被包装过的error常量
func GetInnerMost(err error) (error) {
	if err2, ok := err.(*Err); ok {
		var innerMost error
		for prev := err2; prev != nil; prev = prev.prevErr {
			innerMost = prev.stdError
		}
		return innerMost
	}
	return err
}