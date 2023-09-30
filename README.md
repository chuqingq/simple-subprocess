# simple-subprocess [![GoDoc](https://pkg.go.dev/badge/github.com/chuqingq/simple-subprocess)](https://pkg.go.dev/github.com/chuqingq/simple-subprocess)
A simple subprocess module for Go, like os/exec.


## Features

1. Support using stdin, stdout to communicate to subprocess. Encode/decode as util.Message, like json.
2. Support capturing subprocess's stderr.
3. Support canceling/killing subprocess.


## TODO

- [ ] make sure that if parent exit or panic, subprocess will exit too.
- [x] no dempends on util.Message.
- [x] tests.

## go doc

```
package subprocess // import "github.com/chuqingq/simple-subprocess"


TYPES

type StderrHandler io.Writer
    StderrHandler 用于处理子进程的stderr输出，接收一个io.Writer

type StdoutHandler func(*sjson.Json, error)
    StdoutHandler 用于处理子进程的stdout输出，接收一个*sjson.Json和error

type SubProcess struct {
	Cmd    *exec.Cmd
	Alive  bool
	Ctx    context.Context
	Cancel context.CancelFunc
	Stdin  io.WriteCloser

	Stdout       io.ReadCloser
	HandleStdout StdoutHandler

	Stderr       io.ReadCloser
	HandleStderr StderrHandler
	// Has unexported fields.
}
    SubProcess A subprocess, wrapper for os.exec.Cmd

func New(name string, args ...string) *SubProcess
    New 创建一个SubProcess

func (s *SubProcess) IsAlive() bool
    IsAlive 判断子进程是否存活

func (s *SubProcess) Send(m *sjson.Json) error
    Send 向子进程发送消息

func (s *SubProcess) Start() error
    Start 启动子进程

func (s *SubProcess) Stop()
    Stop 停止子进程并等待结束

func (s *SubProcess) Wait() error
    Wait 等待子进程结束

func (s *SubProcess) WithStderr(handleStderr StderrHandler)
    WithStderr 设置stderr处理函数。 需要在Start前调用。本库内部启动协程执行该回调。

func (s *SubProcess) WithStdout(handleStdout StdoutHandler) *SubProcess
    WithStdout 设置stdout输出的Message处理函数。 需要在Start前调用。本库内部启动协程执行该回调。
```
