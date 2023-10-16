package subprocess

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"testing"
	"time"

	sjson "github.com/chuqingq/simple-json"
	"github.com/stretchr/testify/assert"
)

func TestStop(t *testing.T) {
	p := New("sh", "-c", "sleep 3")
	err := p.Start()
	assert.Nil(t, err)

	assert.Equal(t, false, p.HasFinished())
	p.Stop()
	assert.Equal(t, true, p.HasFinished())
}

func TestWait(t *testing.T) {
	p := New("sh", "-c", "sleep 1")
	err := p.Start()
	assert.Nil(t, err)

	assert.Equal(t, false, p.HasFinished())

	p.Wait()
	assert.Equal(t, true, p.HasFinished())
}

// TestAliveProcessExit 测试子进程被动停止
func TestAliveProcessExit(t *testing.T) {
	p := New("sh", "-c", "echo 123")
	err := p.Start()
	assert.Nil(t, err)
	defer p.Stop()

	assert.Equal(t, false, p.HasFinished())

	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, true, p.HasFinished())
}

func TestStdout(t *testing.T) {
	resChan := make(chan int, 1)
	defer close(resChan)

	number := rand.Int()
	p := New("sh", "-c", "echo "+strconv.Itoa(number))

	handleStdout := func(j *sjson.Json, err error) {
		if err == nil {
			// println("handleStdout: " + j.ToString())
			resChan <- j.MustInt()
		}
		//  else {
		// 	println("handleStdout error: " + err.Error())
		// }
	}
	p.WithStdout(handleStdout)
	err := p.Start()
	assert.Nil(t, err)
	p.Wait()

	res := <-resChan
	assert.Equal(t, number, res)
}

func TestStdoutInvalidJson(t *testing.T) {
	resChan := make(chan string, 1)
	defer close(resChan)

	p := New("sh", "-c", "echo abc; echo def; echo xyz")

	handleStdout := func(j *sjson.Json, err error) {
		if err == nil {
			Logger.Debugf("handleStdout: %v", j.ToString())
		} else if err == io.EOF {
			Logger.Debugf("handleStdout: io.EOF")
		} else {
			resChan <- err.Error()
			Logger.Errorf("handleStdout error: %v", err)
		}
	}
	p.WithStdout(handleStdout)
	err := p.Start()
	assert.Nil(t, err)
	p.Wait()

	var res string

	res = <-resChan
	assert.Equal(t, res, "invalid line: abc")
	res = <-resChan
	assert.Equal(t, res, "invalid line: def")
	res = <-resChan
	assert.Equal(t, res, "invalid line: xyz")
}

func TestStdin(t *testing.T) {
	resChan := make(chan int, 1)
	defer close(resChan)

	number := rand.Int()
	p := New("sh", "-c", "cat ")

	handleStdout := func(j *sjson.Json, err error) {
		if err == nil {
			println("handleStdout: " + j.ToString())
			resChan <- j.Get("number").MustInt()
		}
		//  else {
		// 	println("handleStdout error: " + err.Error())
		// }
	}
	p.WithStdout(handleStdout)
	err := p.Start()
	assert.Nil(t, err)
	defer p.Stop()

	m := &sjson.Json{}
	m.Set("number", number)
	channel := struct {
		Addr string
	}{
		Addr: "rtsp://admin:123456@localhost:554/Streaming/Channels/101",
	}
	// m.Set("channel", channel)
	m.Set("channel", sjson.FromStruct(channel))
	Logger.Debugf("before send")
	err = p.Send(m)
	assert.Nil(t, err)
	Logger.Debugf("before wait")

	Logger.Debugf("wait recv from resChan")
	res := <-resChan
	assert.Equal(t, number, res)
}

func TestStderr(t *testing.T) {
	number := rand.Int()
	input := fmt.Sprintf("%v", number)
	p := New("sh", "-c", "echo -n "+input+" >&2")

	var stderr bytes.Buffer

	p.WithStderr(&stderr)
	err := p.Start()
	assert.Nil(t, err)

	p.Wait()
	Logger.Debugf("stderr: %v", stderr.String())

	assert.Equal(t, input, stderr.String())
}
