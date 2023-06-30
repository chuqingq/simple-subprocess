package subprocess

import (
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"strings"

	sjson "github.com/chuqingq/simple-json"
)

// SubProcess A subprocess, wrapper for os.exec.Cmd
type SubProcess struct {
	Cmd          *exec.Cmd
	Alive        bool
	Ctx          context.Context
	Cancel       context.CancelFunc
	Stdin        io.WriteCloser
	encoder      *json.Encoder
	Stdout       io.ReadCloser
	HandleStdout StdoutHandler
	decoder      *json.Decoder
	Stderr       io.ReadCloser
	HandleStderr StderrHandler
}

// StdoutHandler 用于处理子进程的stdout输出，接收一个*sjson.Json和error
type StdoutHandler func(*sjson.Json, error)

// StderrHandler 用于处理子进程的stderr输出，接收一个io.Writer
type StderrHandler io.Writer

// New 创建一个SubProcess
func New(name string, args ...string) *SubProcess {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, name, args...)

	return &SubProcess{
		Cmd:    cmd,
		Alive:  false,
		Ctx:    ctx,
		Cancel: cancel,
	}
}

// WithStdout 设置stdout输出的Message处理函数。
// 需要在Start前调用。本库内部启动协程执行该回调。
func (s *SubProcess) WithStdout(handleStdout StdoutHandler) *SubProcess {
	s.HandleStdout = handleStdout
	return s
}

// WithStderr 设置stderr处理函数。
// 需要在Start前调用。本库内部启动协程执行该回调。
func (s *SubProcess) WithStderr(handleStderr StderrHandler) {
	s.HandleStderr = handleStderr
}

// Start 启动子进程
func (s *SubProcess) Start() error {
	var err error

	// 如果要和子进程用Message通信
	if s.HandleStdout != nil {
		s.Stdin, err = s.Cmd.StdinPipe()
		if err != nil {
			s.Cancel()
			return err
		}
		s.encoder = json.NewEncoder(s.Stdin)

		s.Stdout, err = s.Cmd.StdoutPipe()
		if err != nil {
			s.Cancel()
			return err
		}
		s.decoder = json.NewDecoder(s.Stdout)

		go s.loopRecvStdout()
	}

	// stderr如果不接收，可能会撑满
	if s.HandleStderr != nil {
		s.Stderr, err = s.Cmd.StderrPipe()
		if err != nil {
			s.Cancel()
			return err
		}
		go s.loopRecvStderr()
	}

	err = s.Cmd.Start()
	if err != nil {
		return err
	}

	s.Alive = true
	return nil
}

// loopRecvStdout 循环接收stdout消息
func (s *SubProcess) loopRecvStdout() {
	for {
		select {
		case <-s.Ctx.Done():
			return
		default:
			m, err := s.doRecvOutMsg()
			s.HandleStdout(m, err)
			if err == nil {
				continue
			} else if err == io.EOF || strings.Contains(err.Error(), "closed") {
				// file already closed
				return
			}
		}
	}
}

// loopRecvStderr 循环接收stderr内容
func (s *SubProcess) loopRecvStderr() {
	io.Copy(s.HandleStderr, s.Stderr)
}

// Wait 等待子进程结束
func (s *SubProcess) Wait() error {
	return s.Cmd.Wait()
}

// Stop 停止子进程并等待结束
func (s *SubProcess) Stop() {
	if s.Alive {
		s.Cancel()
		if s.Stdin != nil {
			s.Stdin.Close()
		}
		if s.Stdout != nil {
			s.Stdout.Close()
		}
		if s.Stderr != nil {
			s.Stderr.Close()
		}
		s.Alive = false
	}
	s.Cmd.Wait()
}

// IsAlive 判断子进程是否存活
func (s *SubProcess) IsAlive() bool {
	return s.Alive
}

// Send 向子进程发送消息
func (s *SubProcess) Send(m *sjson.Json) error {
	err := s.encoder.Encode(m)
	if err != nil && (err == io.ErrClosedPipe || err == io.EOF || strings.Contains(err.Error(), "broken pipe")) {
		s.Cancel()
		s.Alive = false
	}
	return err
}

// doRecvOutMsg 从子进程接收消息
func (s *SubProcess) doRecvOutMsg() (*sjson.Json, error) {
	m := sjson.New()
	err := s.decoder.Decode(m)
	if err != nil {
		if err == io.EOF {
			s.Cancel()
			s.Alive = false
		}
		return nil, err
	}
	return m, nil
}
