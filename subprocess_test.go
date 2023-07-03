package subprocess

import (
	"bytes"
	"fmt"
	"log"
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

	assert.Equal(t, true, p.HasFinished())
	p.Stop()
	assert.Equal(t, false, p.HasFinished())
}

func TestWait(t *testing.T) {
	p := New("sh", "-c", "sleep 1")
	err := p.Start()
	assert.Nil(t, err)

	assert.Equal(t, true, p.HasFinished())

	p.Wait()
	assert.Equal(t, false, p.HasFinished())
}

// TestAliveProcessExit 测试子进程被动停止
func TestAliveProcessExit(t *testing.T) {
	p := New("sh", "-c", "echo 123")
	err := p.Start()
	assert.Nil(t, err)
	defer p.Stop()

	assert.Equal(t, true, p.HasFinished())

	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, false, p.HasFinished())
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
	log.Printf("before send")
	err = p.Send(m)
	assert.Nil(t, err)
	log.Printf("before wait")

	log.Printf("wait recv from resChan")
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
	log.Printf("stderr: %v", stderr.String())

	assert.Equal(t, input, stderr.String())
}
