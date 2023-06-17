package subprocess

import (
	"math/rand"
	"strconv"
	"testing"

	json "github.com/chuqingq/simple-json"
	"github.com/stretchr/testify/assert"
)

func TestDefault(t *testing.T) {
	p := New("sh", "-c", "sleep 10")
	err := p.Start()
	assert.Nil(t, err)
	p.Stop()
	assert.Equal(t, false, p.IsAlive())
}

func TestStdout(t *testing.T) {
	resChan := make(chan int, 1)
	defer close(resChan)

	number := rand.Int()
	p := New("sh", "-c", "echo "+strconv.Itoa(number))

	handleStdout := func(j *json.Json, err error) {
		if err == nil {
			println("handleStdout: " + j.ToString())
			resChan <- j.MustInt()
		}
		//  else {
		// 	println("handleStdout error: " + err.Error())
		// }
	}
	p.WithStdout(handleStdout)
	err := p.Start()
	assert.Nil(t, err)
	p.Cmd.Wait()

	res := <-resChan
	assert.Equal(t, number, res)
}
