package errors

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var goRoot = strings.Replace(runtime.GOROOT(),"\\","/",-1)

var formatPartHead = []byte{'\n', '\t', '['}

var defaultFilterFunc filterFunc = func(fileName string, funcName string) (filtered bool) {
	if strings.HasPrefix(fileName, "github.com") {
		return true
	}
	return false
}

// SetFilterFunc
// if f return true, this caller frame will be excluded in error string, make tidy.
// default, it excluded github.com*
func SetFilterFunc(f filterFunc) {
	defaultFilterFunc = f
}

// filterFunc
// file is the filename which trimmed */src/ prefix.
// funcName is funcName which trimmed / prefix.
type filterFunc func(fileName string, funcName string) (filtered bool)

const (
	formatPartColon = ':'
	formatPartTail  = ']'
	formatPartSpace = ' '
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
				buf.WriteByte(formatPartSpace)
				buf.WriteString(sf.file)
				buf.WriteByte(formatPartColon)
				buf.WriteString(strconv.Itoa(sf.line))
				buf.WriteByte(formatPartSpace)
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
			if defaultFilterFunc != nil && defaultFilterFunc(sf.file, sf.funcName) {
				continue
			}
			//fmt.Fprintf(buf, "\n\t[%s:%d:%s:%s]", sf.file, sf.line, sf.funcName, sf.message)
			buf.Write(formatPartHead)
			// 处理文件名行号时增加空格, 以便让IDE识别到, 可以点击跳转到源码.
			buf.WriteByte(formatPartSpace)
			buf.WriteString(sf.file)
			buf.WriteByte(formatPartColon)
			buf.WriteString(strconv.Itoa(sf.line))
			buf.WriteByte(formatPartSpace)
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
//  上层调用方可以通过GetInnerMost得到里层/最里层被包装过的error常量
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
	newErr := newErr(msg)
	newErr.stdError = err
	return newErr
}

func newErr(msg string) *Err {
	pc := make([]uintptr, 200)
	length := runtime.Callers(3, pc)
	return &Err{
		message: msg,
		stack:   pc[:length],
	}
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