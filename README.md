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
